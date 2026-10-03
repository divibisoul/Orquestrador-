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

func newFabricHarness(t *testing.T) *Engine {
	t.Helper()
	n, err := neural.New(8, .05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(.10, 32)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil { t.Fatal(err) }
	if err := RegisterMultiAgentFacadeOperations(e); err != nil { t.Fatal(err) }
	f, err := supergpu.NewFederation(g)
	if err != nil { t.Fatal(err) }
	if err := RegisterSuperGPUMeshOperations(e, f); err != nil { t.Fatal(err) }
	if err := RegisterSuperAGIFabricOperations(e); err != nil { t.Fatal(err) }
	return e
}

func TestSuperAGIFabricDescribeDeclaresCompositionWithoutAGIOverclaim(t *testing.T) {
	e := newFabricHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricDescribeOperation, nil)
	result, err := e.Submit(context.Background(), m)
	if err != nil { t.Fatal(err) }
	var described map[string]any
	if err := json.Unmarshal([]byte(result.Metadata["fabric_json"]), &described); err != nil { t.Fatal(err) }
	if described["claim"] != "architecture-composition; not proof of general intelligence" {
		t.Fatalf("unexpected claim: %#v", described["claim"])
	}
}

func TestSuperAGIFabricFailClosedWithoutAgentProvider(t *testing.T) {
	e := newFabricHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, nil)
	m.CorrelationID = "corr-superagi"
	m.Metadata["agent_provider"] = "crewai"
	m.Metadata["goal"] = "prepare deterministic compute"
	m.Metadata["compute_operation"] = "square"
	m.Metadata["compute_values_json"] = "[2,3]"
	result, err := e.Submit(context.Background(), m)
	if err == nil { t.Fatal("missing provider must remain blocked") }
	if result.CorrelationID != m.CorrelationID || result.Status != "error" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSuperAGIFabricRequiresComputeStage(t *testing.T) {
	e := newFabricHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, nil)
	m.Metadata["agent_provider"] = "crewai"
	m.Metadata["goal"] = "test"
	result, err := e.Submit(context.Background(), m)
	if err == nil || result.Error != "SUPERAGI_COMPUTE_STAGE_REQUIRED" {
		t.Fatalf("unexpected result: %#v err=%v", result, err)
	}
}
