package execution

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecuteParallelHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var started atomic.Int32
	tasks := make([]Task, 100)
	for i := range tasks {
		tasks[i] = func(ctx context.Context) error {
			started.Add(1)
			<-ctx.Done()
			return ctx.Err()
		}
	}

	done := make(chan []error, 1)
	go func() { done <- New(4).ExecuteParallel(ctx, tasks) }()

	deadline := time.After(2 * time.Second)
	for started.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("workers did not start")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("parallel executor did not return after cancellation")
	}
}

func TestExecuteDistributedRejectsNilDispatcher(t *testing.T) {
	errs := New(1).ExecuteDistributed(context.Background(), []Task{func(context.Context) error { return nil }}, nil)
	if len(errs) != 1 || errs[0] == nil || errs[0].Error() != "nil dispatcher" {
		t.Fatalf("expected nil dispatcher error, got %#v", errs)
	}
}

func TestExecuteDistributedPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := New(1).ExecuteDistributed(ctx, []Task{func(context.Context) error { return nil }}, func(context.Context, Task) error {
		return nil
	})
	if len(errs) != 1 || !errors.Is(errs[0], context.Canceled) {
		t.Fatalf("expected context cancellation, got %#v", errs)
	}
}

func TestCircuitBreakerUsesFailureRatio(t *testing.T) {
	e := New(1)
	if e.CircuitBreaker(false) {
		t.Fatal("circuit opened too early")
	}
	if e.CircuitBreaker(true) {
		t.Fatal("circuit opened on a 50/50 sample below minimum")
	}
	if e.CircuitBreaker(false) {
		t.Fatal("circuit opened before minimum samples")
	}
	if !e.CircuitBreaker(false) {
		t.Fatal("circuit should open at 3 failures / 4 samples")
	}
	if state := e.CircuitState(); state != string(CircuitOpen) {
		t.Fatalf("unexpected state: %s", state)
	}
	if !e.CircuitBreaker(true) {
		t.Fatal("open circuit should remain blocked during cooldown")
	}
}
