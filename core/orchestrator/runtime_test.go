package orchestrator

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestExecuteStepRejectsConcurrentDuplicateExecution(t *testing.T) {
	e := NewEngine(2)
	started := make(chan struct{})
	release := make(chan struct{})
	var runs atomic.Int32

	err := e.CreateWorkflow("w", []Step{{
		ID: "s",
		Run: func(context.Context) error {
			runs.Add(1)
			closeOnce(started)
			<-release
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	firstDone := make(chan error, 1)
	go func() { firstDone <- e.ExecuteStep(context.Background(), "w") }()
	<-started

	secondErr := e.ExecuteStep(context.Background(), "w")
	if !errors.Is(secondErr, ErrWorkflowBusy) {
		t.Fatalf("expected ErrWorkflowBusy, got %v", secondErr)
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	if got := runs.Load(); got != 1 {
		t.Fatalf("expected exactly one execution, got %d", got)
	}
}

func TestRollbackPreservesProgressWhenCompensationFails(t *testing.T) {
	e := NewEngine(1)
	var compensated atomic.Int32
	compensationErr := errors.New("compensation failed")

	steps := []Step{
		{ID: "s1", Run: func(context.Context) error { return nil }, Compensate: func(context.Context) error {
			compensated.Add(1)
			return nil
		}},
		{ID: "s2", Run: func(context.Context) error { return nil }, Compensate: func(context.Context) error {
			return compensationErr
		}},
	}
	if err := e.CreateWorkflow("w", steps); err != nil {
		t.Fatal(err)
	}
	for range steps {
		if err := e.ExecuteStep(context.Background(), "w"); err != nil {
			t.Fatal(err)
		}
	}

	if err := e.RollbackWorkflow(context.Background(), "w"); !errors.Is(err, compensationErr) {
		t.Fatalf("expected compensation error, got %v", err)
	}
	state, err := e.GetWorkflow("w")
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != RollbackFailed {
		t.Fatalf("expected rollback_failed, got %s", state.Status)
	}
	if state.Current != 2 {
		t.Fatalf("expected executed-step cursor preserved at 2, got %d", state.Current)
	}
	if state.RollbackError != compensationErr.Error() {
		t.Fatalf("unexpected rollback error: %q", state.RollbackError)
	}
	if compensated.Load() != 0 {
		t.Fatalf("s1 should not compensate after s2 compensation failed, got %d", compensated.Load())
	}
}

func TestRollbackCompletesAndPausesWorkflow(t *testing.T) {
	e := NewEngine(1)
	var calls atomic.Int32
	steps := []Step{
		{ID: "s1", Run: func(context.Context) error { return nil }, Compensate: func(context.Context) error {
			calls.Add(1)
			return nil
		}},
		{ID: "s2", Run: func(context.Context) error { return nil }, Compensate: func(context.Context) error {
			calls.Add(1)
			return nil
		}},
	}
	if err := e.CreateWorkflow("w", steps); err != nil {
		t.Fatal(err)
	}
	for range steps {
		if err := e.ExecuteStep(context.Background(), "w"); err != nil {
			t.Fatal(err)
		}
	}
	if err := e.RollbackWorkflow(context.Background(), "w"); err != nil {
		t.Fatal(err)
	}
	state, err := e.GetWorkflow("w")
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != Paused || state.Current != 0 {
		t.Fatalf("unexpected state after rollback: status=%s current=%d", state.Status, state.Current)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected two compensations, got %d", calls.Load())
	}
}

func TestRetryFailedStepRequiresFailedState(t *testing.T) {
	e := NewEngine(1)
	var attempts atomic.Int32
	if err := e.CreateWorkflow("w", []Step{{ID: "s", Run: func(context.Context) error {
		if attempts.Add(1) == 1 {
			return errors.New("first failure")
		}
		return nil
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := e.ExecuteStep(context.Background(), "w"); err == nil {
		t.Fatal("expected first execution to fail")
	}
	if err := e.RetryFailedStep(context.Background(), "w"); err != nil {
		t.Fatal(err)
	}
	state, err := e.GetWorkflow("w")
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != Completed || state.Current != 1 {
		t.Fatalf("unexpected retry state: status=%s current=%d", state.Status, state.Current)
	}
}

func closeOnce(ch chan struct{}) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}
