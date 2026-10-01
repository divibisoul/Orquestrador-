package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestOrbitalReasoningConnectsTCEToPrefrontal(t *testing.T) {
	t.Setenv("N07_TCE_ENABLED", "true")
	n, err := neural.New(4, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0, 8)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(n, c, supergpu.New(nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterOrbitalReasoningOperations(e); err != nil {
		t.Fatal(err)
	}

	workloads, _ := json.Marshal([]core.Workload{{
		ID: "orbit-task-1", Operation: "reasoning-simulation", Precision: core.FP16,
		MatrixSize: 256, BatchSize: 1, DataBytes: 4096, MemoryNeeded: 1, Priority: 1,
	}})
	candidate, _ := json.Marshal(map[string]any{
		"ID": "orbit-candidate-1", "Cost": 0.1, "Risk": 0.1, "Utility": 0.9,
		"Uncertainty": 0.1, "Urgency": 0.2, "Impact": 0.2,
	})
	m := protocol.NewMessage("N06", "N07", "command", "prefrontal.orbital.evaluate@1.0.0", []float64{1, 2, 3, 4})
	m.Metadata["workloads_json"] = string(workloads)
	m.Metadata["candidate_json"] = string(candidate)
	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("status=%q error=%q", result.Status, result.Error)
	}
	if result.Metadata["simulation_is_not_hardware"] != "true" {
		t.Fatalf("missing simulation boundary: %#v", result.Metadata)
	}
	if result.Metadata["decision_id"] != "orbit-candidate-1" {
		t.Fatalf("missing PFC decision: %#v", result.Metadata)
	}
}

func TestOrbitalReasoningFailsClosedWhenTCEDisabled(t *testing.T) {
	os.Unsetenv("N07_TCE_ENABLED")
	n, _ := neural.New(4, .05)
	c, _ := prefrontal.New(0, 8)
	e, err := New(n, c, supergpu.New(nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterOrbitalReasoningOperations(e); err != nil {
		t.Fatal(err)
	}
	m := protocol.NewMessage("N01", "N07", "command", "transcendental.estimate@1.0.0", nil)
	result, err := e.Submit(context.Background(), m)
	if err == nil || result.Error != "TCE_DISABLED" {
		t.Fatalf("expected fail-closed TCE_DISABLED, result=%#v err=%v", result, err)
	}
}
