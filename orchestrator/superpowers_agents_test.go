package orchestrator

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestSuperpowersAgentBindingsCoverRequestedCrossfronts(t *testing.T) {
	want := map[string][]string{
		"superpowers.sara-trinity":            {"ARA", "ETR", "ITR", "RGO", "MMD"},
		"superpowers.cortex-orbital-supergpu": {"prefrontal-neocortex", "orbital-reasoning", "SuperGPU"},
		"superpowers.mesh-clareira":           {"SOUL-Mesh", "Projeto-Clareira"},
		"superpowers.octacore":                {"Octacore", "Mesh", "SuperGPU"},
		"superpowers.prefrontal-optimizer":    {"prefrontal-neocortex", "neural-network", "processing-runtime", "multi-agent-runtime", "N01.SoulNeuralGraph", "N01.N01AgentRegistry", "N01.N01AgentProcessor", "N07.neural.Network"},
	}
	got := SuperpowersAgentBindings()
	if len(got) != len(want) {
		t.Fatalf("bindings=%d want=%d", len(got), len(want))
	}
	for _, b := range got {
		expected, ok := want[b.ID]
		if !ok {
			t.Fatalf("unexpected binding %s", b.ID)
		}
		if len(expected) != len(b.Targets) {
			t.Fatalf("%s targets=%v want=%v", b.ID, b.Targets, expected)
		}
		for _, target := range expected {
			found := false
			for _, actual := range b.Targets {
				if actual == target {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("%s missing target %s in %v", b.ID, target, b.Targets)
			}
		}
		if b.ID == "superpowers.prefrontal-optimizer" {
			foundBridge := false
			for _, operation := range b.Operations {
				if operation == "neural.prefrontal.optimize@1.0.0" {
					foundBridge = true
					break
				}
			}
			if !foundBridge {
				t.Fatal("prefrontal specialist must be able to delegate to N01's native Mesh capability")
			}
		}
		if !b.FailClosed {
			t.Fatalf("%s must fail closed", b.ID)
		}
	}
}

func TestSuperpowersAgentRouteRejectsUnboundOperation(t *testing.T) {
	n, err := neural.New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	p, err := prefrontal.New(.1, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := New(n, p, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperpowersAgentOperations(e); err != nil {
		t.Fatal(err)
	}
	m := protocol.NewMessage("N01", "N07", "command", SuperpowersAgentRouteOperation, nil)
	m.Metadata["agent_id"] = "superpowers.mesh-clareira"
	m.Metadata["operation"] = "octacore.batch"
	_, err = e.Submit(context.Background(), m)
	if err == nil {
		t.Fatal("expected fail-closed rejection")
	}
}

func TestPrefrontalOptimizerAgentExecutesThroughNeuralExecutiveAndPreservesTrace(t *testing.T) {
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	p, err := prefrontal.New(.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := New(n, p, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterAdvancedOperations(e); err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperpowersAgentOperations(e); err != nil {
		t.Fatal(err)
	}

	m := protocol.NewMessage("N01", "N07", "command", SuperpowersAgentRouteOperation, []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8})
	m.Metadata["agent_id"] = "superpowers.prefrontal-optimizer"
	m.Metadata["operation"] = "prefrontal.admission@1.0.0"
	m.Metadata["candidate_json"] = `{"ID":"pfc-optimizer-e2e","Cost":0.01,"Risk":0.01,"Urgency":0.5,"Impact":0.8}`
	m.Metadata["task_id"] = "pfc-optimizer-e2e"

	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatalf("optimizer agent failed to execute prefrontal admission: %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("optimizer agent status=%q error=%q", result.Status, result.Error)
	}
	if result.TraceID == "" || result.TraceID != m.TraceID {
		t.Fatalf("root trace was not preserved: request=%q response=%q", m.TraceID, result.TraceID)
	}
	if result.Metadata["parent_trace_id"] != m.TraceID {
		t.Fatalf("parent trace linkage missing: request=%q metadata=%#v", m.TraceID, result.Metadata)
	}
	if child := result.Metadata["delegated_trace_id"]; child == "" || child == m.TraceID {
		t.Fatalf("delegated execution trace missing or not distinct: root=%q metadata=%#v", m.TraceID, result.Metadata)
	}
	if result.CorrelationID != m.CorrelationID {
		t.Fatalf("correlation not preserved: request=%q response=%q", m.CorrelationID, result.CorrelationID)
	}
	if result.Metadata["decision_id"] != "pfc-optimizer-e2e" {
		t.Fatalf("prefrontal decision not committed: %#v", result.Metadata)
	}
	if result.Metadata["superpowers_agent"] != "superpowers.prefrontal-optimizer" {
		t.Fatalf("specialist identity missing: %#v", result.Metadata)
	}
	if result.Metadata["neural_dimensions"] != "8" {
		t.Fatalf("expected 8 neural dimensions, got %#v", result.Metadata["neural_dimensions"])
	}
}

func TestPrefrontalOptimizerAgentCannotCallUnboundOperation(t *testing.T) {
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	p, err := prefrontal.New(.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := New(n, p, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterAdvancedOperations(e); err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperpowersAgentOperations(e); err != nil {
		t.Fatal(err)
	}

	m := protocol.NewMessage("N01", "N07", "command", SuperpowersAgentRouteOperation, []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8})
	m.Metadata["agent_id"] = "superpowers.prefrontal-optimizer"
	m.Metadata["operation"] = "octacore.batch"
	_, err = e.Submit(context.Background(), m)
	if err == nil {
		t.Fatal("optimizer must reject operations outside its explicit capability allowlist")
	}
}

func TestResidentAgentDescribeIsRegistered(t *testing.T) {
	n, err := neural.New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	p, err := prefrontal.New(.1, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := New(n, p, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperpowersAgentOperations(e); err != nil {
		t.Fatal(err)
	}
	m := protocol.NewMessage("N01", "N07", "command", ResidentAgentDescribeOperation, nil)
	result, err := e.Submit(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("status=%s error=%s", result.Status, result.Error)
	}
	if result.Metadata["resident_agent_json"] == "" {
		t.Fatal("missing resident agent evidence")
	}
}
