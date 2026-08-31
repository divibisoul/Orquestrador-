package transcendental

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

func TestSelectorHeuristics(t *testing.T) {
	cases := []struct {
		w    trinity.Workload
		want string
	}{
		{trinity.Workload{Precision: "fp64"}, "blackwell"},
		{trinity.Workload{MemoryNeeded: 500}, "mi400"},
		{trinity.Workload{MatrixSize: 5000}, "vera_rubin"},
		{trinity.Workload{MatrixSize: 500}, "trillium"},
		{trinity.Workload{BatchSize: 100}, "blackwell"},
	}
	for _, tc := range cases {
		if got := Select(tc.w).Name; got != tc.want {
			t.Fatalf("got %s want %s", got, tc.want)
		}
	}
}

func TestComputeEngineUsesRealCPUBackend(t *testing.T) {
	e := NewEngine(trinity.ComputeConfig{})
	ctx := trinity.WithTraceID(context.Background(), "trace-compute")
	r, err := e.Execute(ctx, trinity.Workload{ID: "x", Kind: "text", Payload: "hello", BatchSize: 2}, trinity.Route{Model: "blackwell"})
	if err != nil || !r.Success {
		t.Fatalf("result=%+v err=%v", r, err)
	}
	if r.Metadata["backend"] != "cpu" || r.Metadata["execution"] != "deterministic" {
		t.Fatalf("unexpected backend metadata: %+v", r.Metadata)
	}
	if r.Metadata["trace_id"] != "trace-compute" {
		t.Fatalf("trace id lost: %+v", r.Metadata)
	}
	if r.Output == "simulated inference completed" || r.Output == "" {
		t.Fatalf("compute did not produce concrete output: %q", r.Output)
	}
}

func TestComputeEstimator(t *testing.T) {
	e := NewEngine(trinity.ComputeConfig{})
	est, err := e.Estimate(context.Background(), trinity.Workload{ID: "x", Kind: "matrix", MatrixSize: 1024, BatchSize: 8})
	if err != nil {
		t.Fatal(err)
	}
	if est.LatencyMS <= 0 || est.Memory <= 0 || est.Confidence <= 0 || est.Confidence > 1 {
		t.Fatalf("invalid estimate: %+v", est)
	}
}
