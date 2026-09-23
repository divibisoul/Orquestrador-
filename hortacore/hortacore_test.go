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
