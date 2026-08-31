package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/executor"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
)

func TestExecutorDisabledIsNonOperational(t *testing.T) {
	cfg := core.DefaultConfig()
	eng, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := executor.New(eng, models.DefaultCatalog(cfg), "auto")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ex.Execute(context.Background(), core.Workload{ID: "disabled", Operation: "matmul", Precision: core.FP8, MatrixSize: 1024, BatchSize: 1})
	if err == nil {
		t.Fatal("expected disabled error")
	}
}

func TestExecutorDeterministicAndMetrics(t *testing.T) {
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
	wl := core.Workload{ID: "det", Operation: "matmul", Precision: core.FP8, MatrixSize: 4096, BatchSize: 2, DataBytes: 1 << 28, MemoryNeeded: 64}
	a, err := ex.Estimate(context.Background(), wl)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ex.Estimate(context.Background(), wl)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("estimates are not deterministic: %#v != %#v", a, b)
	}
	r, err := ex.Execute(context.Background(), wl)
	if err != nil {
		t.Fatal(err)
	}
	if string(r.Data) != "simulated-result" {
		t.Fatalf("unexpected result data %q", r.Data)
	}
	if r.Metrics.Architecture == "" || r.Metrics.LatencyMs <= 0 {
		t.Fatalf("invalid metrics %#v", r.Metrics)
	}
}

func TestExecutePlanPreservesPerItemErrors(t *testing.T) {
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
	workloads := []core.Workload{
		{ID: "valid", Operation: "matmul", Precision: core.FP8, MatrixSize: 1024, BatchSize: 1, DataBytes: 1 << 20, MemoryNeeded: 1},
		{ID: "invalid", Operation: "matmul", Precision: core.FP8, MatrixSize: 0, BatchSize: 1, DataBytes: 1 << 20, MemoryNeeded: 1},
	}
	results := ex.ExecutePlan(context.Background(), workloads)
	if len(results) != len(workloads) {
		t.Fatalf("results=%d want %d", len(results), len(workloads))
	}
	if results[0].Error != nil {
		t.Fatalf("unexpected valid-workload error: %v", results[0].Error)
	}
	if results[1].Error == nil {
		t.Fatal("expected invalid-workload error to be preserved")
	}
}

func TestExecutePlanNilExecutorIsSafe(t *testing.T) {
	var ex *executor.SimulatedExecutor
	results := ex.ExecutePlan(context.Background(), []core.Workload{{ID: "nil", Operation: "matmul", Precision: core.FP8, MatrixSize: 1, BatchSize: 1}})
	if len(results) != 1 || results[0].Error == nil {
		t.Fatalf("expected initialization error, got %#v", results)
	}
}

func TestExecuteRejectsNilContext(t *testing.T) {
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
	_, err = ex.Execute(nil, core.Workload{ID: "nil-context", Operation: "matmul", Precision: core.FP8, MatrixSize: 1024, BatchSize: 1})
	if !errors.Is(err, context.Canceled) && err == nil {
		t.Fatal("expected nil-context error")
	}
}

func TestExecutePlanCanceledContextMarksPendingWork(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Enabled = true
	cfg.MaxParallelUnits = 1
	eng, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := executor.New(eng, models.DefaultCatalog(cfg), "auto")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := ex.ExecutePlan(ctx, []core.Workload{{ID: "a", Operation: "matmul", Precision: core.FP8, MatrixSize: 1024, BatchSize: 1}, {ID: "b", Operation: "matmul", Precision: core.FP8, MatrixSize: 1024, BatchSize: 1}})
	for _, result := range results {
		if !errors.Is(result.Error, context.Canceled) {
			t.Fatalf("expected context cancellation, got %#v", result)
		}
	}
}

func TestExecutePlanCapsWorkersAtSimulationLimit(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Enabled = true
	cfg.MaxParallelUnits = 8192
	eng, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if eng.Config.EffectiveParallelUnits() != core.MaxSimulationParallelUnits {
		t.Fatalf("effective workers=%d want %d", eng.Config.EffectiveParallelUnits(), core.MaxSimulationParallelUnits)
	}
}
