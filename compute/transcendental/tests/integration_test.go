package tests

import (
	"context"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/executor"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
	"testing"
)

func TestEngineContracts(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Enabled = true
	eng, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := executor.New(eng, models.DefaultCatalog(cfg), "auto")
	if err != nil {
		t.Fatal(err)
	}
	plan := core.Plan{Strategy: "latency_first", Workloads: []core.Workload{{ID: "a", Operation: "attention", Precision: core.FP8, MatrixSize: 512, BatchSize: 2, DataBytes: 1 << 20, MemoryNeeded: 2}, {ID: "b", Operation: "matmul", Precision: core.BF16, MatrixSize: 4096, BatchSize: 1, DataBytes: 1 << 28, MemoryNeeded: 64}}}
	ce, err := ex.EstimateCost(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if ce.EstimatedTime <= 0 || ce.Confidence <= 0 || ce.Confidence > 1 {
		t.Fatalf("invalid plan estimate %#v", ce)
	}
	r, err := ex.Execute(context.Background(), plan.Workloads[0])
	if err != nil {
		t.Fatal(err)
	}
	if ex.GetLastMetrics().Architecture != r.Metrics.Architecture {
		t.Fatal("metrics provider did not retain last metrics")
	}
	if len(ex.GetMetricsHistory(10)) != 1 {
		t.Fatal("metrics history not updated")
	}
}
