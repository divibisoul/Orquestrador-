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
	if err := RegisterMultiAgentFacadeOperations(e); err != nil {
		t.Fatal(err)
	}
	f, err := supergpu.NewFederation(g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperGPUMeshOperations(e, f); err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperAGIFabricOperations(e); err != nil {
		t.Fatal(err)
	}
	return e
}

func newSuperAGIFabricAgentStubHarness(t *testing.T) *Engine {
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
	if err := e.Register(MultiAgentExecuteOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.multiagent", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"result_json": `{"state":"PASS","provider":"stub"}`},
		}, nil
	}); err != nil {
		t.Fatal(err)
	}
	f, err := supergpu.NewFederation(g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperGPUMeshOperations(e, f); err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperAGIFabricOperations(e); err != nil {
		t.Fatal(err)
	}
	return e
}

func TestSuperAGIFabricDescribeDeclaresCompositionWithoutAGIOverclaim(t *testing.T) {
	e := newFabricHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricDescribeOperation, nil)
	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	var described map[string]any
	if err := json.Unmarshal([]byte(result.Metadata["fabric_json"]), &described); err != nil {
		t.Fatal(err)
	}
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
	if err == nil {
		t.Fatal("missing provider must remain blocked")
	}
	if result.CorrelationID != m.CorrelationID || result.Status != "error" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSuperAGIFabricRejectsIncompleteRequestedCompute(t *testing.T) {
	e := newSuperAGIFabricAgentStubHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, nil)
	m.Metadata["agent_provider"] = "stub"
	m.Metadata["goal"] = "test"
	m.Metadata["compute_operation"] = "square"
	result, err := e.Submit(context.Background(), m)
	if err == nil || result.Error != "SUPERAGI_COMPUTE_STAGE_REQUIRED" {
		t.Fatalf("unexpected result: %#v err=%v", result, err)
	}
}

func TestSuperAGIFabricKeepsAcceleratorRequirementExplicit(t *testing.T) {
	e := newSuperAGIFabricAgentStubHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, []float64{2, 3})
	m.Metadata["agent_provider"] = "stub"
	m.Metadata["goal"] = "real accelerator required"
	m.Metadata["compute_operation"] = "square"
	m.Metadata["compute_values_json"] = "[2,3]"
	m.Metadata["require_accelerator"] = "true"
	result, err := e.Submit(context.Background(), m)
	if err == nil || result.Error != "SUPERAGI_COMPUTE_STAGE_FAILED:SUPERGPU_ACCELERATOR_UNAVAILABLE" {
		t.Fatalf("unexpected result: %#v err=%v", result, err)
	}
}
func TestSuperAGIFabricAgentArsenalBoundaryFailsClosedWithoutArtifactCoordinates(t *testing.T) {
	e := newFabricHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, nil)
	m.Metadata["agent_mode"] = "agent-arsenal"
	m.Metadata["agent_provider"] = "superpowers"
	m.Metadata["goal"] = "verify artifact boundary"
	m.Metadata["compute_operation"] = "square"
	m.Metadata["compute_values_json"] = "[2,3]"
	result, err := e.Submit(context.Background(), m)
	if err == nil || result.Error != "SUPERAGI_AGENT_ARSENAL_BOUNDARY_REQUIRED" {
		t.Fatalf("unexpected result: %#v err=%v", result, err)
	}
}

func TestSuperAGIFabricSkipsOptionalComputeStage(t *testing.T) {
	e := newSuperAGIFabricAgentStubHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, nil)
	m.CorrelationID = "corr-superagi-skip"
	m.Metadata["agent_provider"] = "stub"
	m.Metadata["goal"] = "agent-only task"
	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata["agent_stage"] != "PASS" || result.Metadata["compute_stage"] != "SKIPPED" || result.Metadata["compute_requested"] != "false" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestSuperAGIFabricAllowsCPUComputeWhenAcceleratorNotRequested(t *testing.T) {
	e := newSuperAGIFabricAgentStubHarness(t)
	m := protocol.NewMessage(protocol.N01, protocol.N07, "request", SuperAGIFabricExecuteOperation, []float64{2,3})
	m.CorrelationID = "corr-superagi-cpu"
	m.Metadata["agent_provider"] = "stub"
	m.Metadata["goal"] = "deterministic CPU compute"
	m.Metadata["compute_operation"] = "square"
	m.Metadata["compute_values_json"] = "[2,3]"
	result, err := e.Submit(context.Background(), m)
	if err != nil { t.Fatal(err) }
	if result.Metadata["compute_stage"] != "PASS" || result.Metadata["accelerator"] != "false" || result.Metadata["backend"] != "cpu" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(result.Payload) != 2 || result.Payload[0] != 4 || result.Payload[1] != 9 {
		t.Fatalf("unexpected compute payload: %#v", result.Payload)
	}
}
