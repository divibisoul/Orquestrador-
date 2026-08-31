package trinity

import (
	"context"
	"sync/atomic"
	"testing"
)

type pfcStub struct {
	called atomic.Bool
	gated  atomic.Bool
}

func (p *pfcStub) Evaluate(context.Context, Workload, ComputeEngine) (Decision, error) {
	p.called.Store(true)
	return Decision{Approved: true, Strategy: Strategy{Precision: "fp32"}}, nil
}

func (p *pfcStub) GateWorkingMemory(context.Context, Workload, Result) error {
	p.gated.Store(true)
	return nil
}

type fabricStub struct {
	routed    atomic.Bool
	feedbacked atomic.Bool
}

func (f *fabricStub) Route(context.Context, Strategy, Workload) (Route, error) {
	f.routed.Store(true)
	return defaultTestRoute(), nil
}

func (f *fabricStub) Feedback(context.Context, Feedback) error {
	f.feedbacked.Store(true)
	return nil
}

type computeStub struct{}

func (computeStub) Execute(context.Context, Workload, Route) (Result, error) {
	return Result{Success: true, Metadata: map[string]string{}}, nil
}

func defaultTestRoute() Route {
	return Route{Target: "local", Model: "blackwell", Provider: "test"}
}

func TestTrinityFullFlow(t *testing.T) {
	p := &pfcStub{}
	f := &fabricStub{}
	o := &TrinityOrchestrator{
		PFC: p, Fabric: f, Compute: computeStub{},
		Config: TrinityConfig{PFCEnabled: true, FabricEnabled: true, ComputeEnabled: true},
	}
	ctx := WithTraceID(context.Background(), "trace-123")
	r, err := o.ExecuteTask(ctx, Task{ID: "1", Kind: "text", Payload: "hello"})
	if err != nil || !r.Success || !p.called.Load() || !p.gated.Load() || !f.routed.Load() || !f.feedbacked.Load() {
		t.Fatalf("flow failed r=%+v err=%v", r, err)
	}
	if r.Metadata["trace_id"] != "trace-123" {
		t.Fatalf("trace id missing: %+v", r.Metadata)
	}
}

func TestFallback(t *testing.T) {
	o := &TrinityOrchestrator{Config: TrinityConfig{FallbackMode: "legacy"}}
	_, err := o.ExecuteTask(context.Background(), Task{ID: "1", Kind: "text"})
	if err != ErrDisabled {
		t.Fatalf("got %v", err)
	}
}

func TestTraceID(t *testing.T) {
	ctx := WithTraceID(context.Background(), "abc")
	if TraceID(ctx) != "abc" {
		t.Fatal("trace id lost")
	}
}

func TestUnknownTask(t *testing.T) {
	_, err := adaptTaskToWorkload(struct{ ID string }{ID: "x"})
	if err != ErrUnknownTask {
		t.Fatalf("got %v", err)
	}
}

func TestNegativeWorkloadRejected(t *testing.T) {
	_, err := adaptTaskToWorkload(Task{ID: "1", Kind: "text", MatrixSize: -1})
	if err == nil {
		t.Fatal("expected negative resource validation error")
	}
}
