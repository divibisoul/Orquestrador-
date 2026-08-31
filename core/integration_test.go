package core_test

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/compute"
	"github.com/divibisoul/Orquestrador-/core/neuralfabric"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/core/superagi"
)

func TestCognitiveLoop(t *testing.T) {
	ctx := context.Background()
	d := compute.Device{ID: "local-cpu", Kind: compute.CPU, FLOPs: 1e9, MemoryMB: 16384, Precisions: []compute.Precision{compute.FP32, compute.INT8}, Ready: true}
	fabric := compute.NewLocalFabric([]compute.Device{d})
	nf := neuralfabric.NewRuntime()
	cx := prefrontal.New()

	r, err := nf.Route(ctx, neuralfabric.State{Goal: "test"}, []neuralfabric.Route{{NodeID: d.ID, DeviceID: d.ID, Precision: "int8", BatchSize: 1}})
	if err != nil || r.DeviceID != d.ID { t.Fatalf("route failed: %v", err) }
	res, err := fabric.Execute(ctx, compute.Job{ID: "test", Tokens: 1, Precision: compute.INT8}, d)
	if err != nil || res.JobID != "test" || res.DeviceID != d.ID { t.Fatalf("compute failed: %v", err) }
	cx.OptimizePolicy(0.1)
	agi := superagi.NewRuntime().WithProvider(integrationProvider{})
	if _, err := agi.GenerateEmbedding(ctx, "hello"); err != nil { t.Fatal(err) }
	e := orchestrator.NewEngine(2)
	if err := e.CreateWorkflow("w", []orchestrator.Step{{ID: "s", Run: func(context.Context) error { return nil }}}); err != nil { t.Fatal(err) }
	if err := e.ExecuteStep(ctx, "w"); err != nil { t.Fatal(err) }
}
