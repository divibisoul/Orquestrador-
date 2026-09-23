package hortacore

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/octacore"
)

func TestHortaCorePreservesFiveProcessorIdentities(t *testing.T) {
	processor, err := octacore.NewProcessor(octacore.DefaultSchedulerConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fusion, err := NewFusion(processor, nil)
	if err != nil {
		t.Fatal(err)
	}
	items := fusion.Describe()["processors"].([]map[string]any)
	if len(items) != 5 {
		t.Fatalf("expected five HortaCore processor identities, got %d", len(items))
	}
	if fusion.Describe()["octacore_slot_owner"] != "G7" {
		t.Fatalf("expected N07/Octacore ownership")
	}
}

func TestHortaCoreAuditRemainsSARAOwned(t *testing.T) {
	processor, err := octacore.NewProcessor(octacore.DefaultSchedulerConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fusion, err := NewFusion(processor, nil)
	if err != nil {
		t.Fatal(err)
	}
	result := fusion.Dispatch(context.Background(), ProcessorAudit, "audit boundary", "horta-test-corr")
	if result.OK {
		t.Fatal("audit must fail closed when SARA endpoint is not configured")
	}
	if result.Error == nil || result.Error.Message == "" {
		t.Fatal("expected explicit SARA boundary failure")
	}
}


func TestHortaCoreVagusSignalsControlSharedOctacoreProcessor(t *testing.T) {
	processor, err := octacore.NewProcessor(octacore.DefaultSchedulerConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fusion, err := NewFusion(processor, nil)
	if err != nil {
		t.Fatal(err)
	}
	state, err := fusion.ApplySignal("throttle", 2)
	if err != nil {
		t.Fatal(err)
	}
	health, ok := state["health"].(octacore.SchedulerHealth)
	if !ok {
		t.Fatalf("expected Octacore scheduler health, got %#v", state["health"])
	}
	if health.ThrottleLevel != 2 {
		t.Fatalf("expected throttle 2, got %d", health.ThrottleLevel)
	}
	if _, err := fusion.ApplySignal("halt", 0); err != nil {
		t.Fatal(err)
	}
	if processor.Health().Status != "HALTED" {
		t.Fatalf("expected shared Octacore processor to be halted")
	}
	if _, err := fusion.ApplySignal("resume", 0); err != nil {
		t.Fatal(err)
	}
	if processor.Health().Status == "HALTED" {
		t.Fatalf("expected shared Octacore processor to resume")
	}
}

func TestHortaCoreSyncMeshPreservesAllSixNucleiWithoutSecondMesh(t *testing.T) {
	processor, err := octacore.NewProcessor(octacore.DefaultSchedulerConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fusion, err := NewFusion(processor, nil)
	if err != nil {
		t.Fatal(err)
	}
	results := fusion.SyncMesh(context.Background(), "horta-sync-test")
	if len(results) != 6 {
		t.Fatalf("expected six existing Mesh peers, got %d", len(results))
	}
	for _, nucleus := range []string{"N01", "N02", "N03", "N04", "N05", "N06"} {
		if _, ok := results[nucleus]; !ok {
			t.Fatalf("missing Mesh discovery result for %s", nucleus)
		}
	}
}
