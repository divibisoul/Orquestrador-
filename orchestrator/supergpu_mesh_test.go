package orchestrator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newMeshGPUHarness(t *testing.T) *Engine {
	t.Helper()
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	f, err := supergpu.NewFederation(g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperGPUMeshOperations(e, f); err != nil {
		t.Fatal(err)
	}
	return e
}

func TestSuperGPUMeshPreservesCorrelationAndBlocksCPUWhenAcceleratorRequired(t *testing.T) {
	e := newMeshGPUHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperGPUMeshExecuteOperation, []float64{2, 3})
	m.CorrelationID = "corr-gpu-mesh"
	m.Metadata["operation"] = "square"
	result, err := e.Submit(context.Background(), m)
	if err == nil {
		t.Fatal("CPU-only environment must not claim accelerator execution")
	}
	if result.CorrelationID != m.CorrelationID || result.Status != "error" || result.Error != "SUPERGPU_ACCELERATOR_UNAVAILABLE" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSuperGPUMeshAllowsExplicitCPUCompatibilityMode(t *testing.T) {
	e := newMeshGPUHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperGPUMeshExecuteOperation, []float64{2, 3})
	m.Metadata["operation"] = "square"
	m.Metadata["require_accelerator"] = "false"
	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Payload) != 2 || result.Payload[0] != 4 || result.Payload[1] != 9 {
		t.Fatalf("unexpected CPU compatibility result: %#v", result.Payload)
	}
	if result.Metadata["accelerator"] != "false" {
		t.Fatalf("expected accelerator=false, got %q", result.Metadata["accelerator"])
	}
}

func TestSuperGPUMeshRejectsN07SelfFederation(t *testing.T) {
	e := newMeshGPUHarness(t)
	m := protocol.NewMessage(protocol.N07, protocol.N07, "request", SuperGPUMeshExecuteOperation, []float64{1})
	m.Metadata["operation"] = "identity"
	_, err := e.Submit(context.Background(), m)
	if err == nil {
		t.Fatal("N07 must use its local SuperGPU API rather than self-federating")
	}
}

func TestSuperGPUMeshParallelMetadata(t *testing.T) {
	e := newMeshGPUHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperGPUMeshParallelOperation, nil)
	m.Metadata["operation"] = "identity"
	m.Metadata["require_accelerator"] = "false"
	raw, _ := json.Marshal([][]float64{{1}, {2}})
	m.Metadata["inputs_json"] = string(raw)
	m.Metadata["workers"] = "2"
	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Metadata["accelerator"]; got != "false" {
		t.Fatalf("unexpected accelerator marker %q", got)
	}
}
