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

func TestSOULTopologyIsSevenNucleusAdjacentChain(t *testing.T) {
	top := SOULTopology()
	nuclei, ok := top["nuclei"].([]string)
	if !ok || len(nuclei) != 7 {
		t.Fatalf("invalid nuclei topology: %#v", top["nuclei"])
	}
	if top["directional"] != 12 {
		t.Fatalf("expected 12 directional channels, got %v", top["directional"])
	}
	if top["fusion_policy"] != "adjacent-only-dynamic" {
		t.Fatalf("unexpected fusion policy: %v", top["fusion_policy"])
	}
}

func TestAdvancedOperationsExecuteWithRealServices(t *testing.T) {
	n, err := neural.New(4, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperGPUOperations(e); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAdvancedOperations(e); err != nil {
		t.Fatal(err)
	}

	candidate := map[string]any{"ID": "safe-action", "Cost": 0.01, "Risk": 0.01, "Utility": 0.5, "Uncertainty": 0.01, "Urgency": 0.1, "Impact": 0.1}
	candidateJSON, _ := json.Marshal(candidate)
	result, err := e.Execute(context.Background(), "prefrontal.admission@1.0.0", []float64{1, 2, 3, 4}, map[string]string{"candidate_json": string(candidateJSON)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("unexpected admission status: %#v", result)
	}
	if len(c.Recall(1)) != 1 {
		t.Fatal("expected committed prefrontal decision")
	}

	gpuResult, err := e.Execute(context.Background(), "supergpu.federated.execute@1.0.0", []float64{2, 3}, map[string]string{"nucleus": "N01", "operation": "square"})
	if err != nil {
		t.Fatal(err)
	}
	if len(gpuResult.Payload) != 2 || gpuResult.Payload[0] != 4 || gpuResult.Payload[1] != 9 {
		t.Fatalf("unexpected federated compute: %#v", gpuResult.Payload)
	}
}

func TestPrefrontalAdmissionCanRequireJevWithoutReplacingPrefrontalAuthority(t *testing.T) {
	n, err := neural.New(4, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(n, c, supergpu.New(nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterAdvancedOperations(e); err != nil {
		t.Fatal(err)
	}

	called := false
	if err := e.Register("jev.systemone@1.0.0", func(_ context.Context, message protocol.Message) (protocol.Result, error) {
		called = true
		if message.Metadata["questions_json"] == "" {
			t.Fatal("Jev questions were not forwarded")
		}
		if message.CorrelationID == "" {
			t.Fatal("Jev correlation was not forwarded")
		}
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.jev", Target: message.Source, Status: "ok", Metadata: map[string]string{
			"decision_json": "{\"answers\":{\"needs_human\":{\"noul\":0.2}}}",
			"answers_json":  "{\"needs_human\":{\"noul\":0.2}}",
			"usage_json":    "{\"input_tokens\":4}",
			"model":         "jev-test",
		}}, nil
	}); err != nil {
		t.Fatal(err)
	}

	candidate := map[string]any{"ID": "jev-gated-action", "Cost": 0.01, "Risk": 0.1, "Utility": 0.8, "Uncertainty": 0.05, "Urgency": 0.3, "Impact": 0.2}
	candidateJSON, _ := json.Marshal(candidate)
	result, err := e.Execute(context.Background(), "prefrontal.admission@1.0.0", []float64{1, 2, 3, 4}, map[string]string{
		"candidate_json": string(candidateJSON), "correlation_id": "corr-jev-gate", "jev_required": "true",
		"jev_questions_json": "{\"needs_human\":{\"type\":\"noul\",\"instructions\":\"Does this action need human review?\"}}",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected Jev specialist to be invoked")
	}
	if result.Metadata["jev_checked"] != "true" || result.Metadata["jev_answers_json"] == "" {
		t.Fatalf("missing Jev evidence: %#v", result.Metadata)
	}
	if result.CorrelationID != "corr-jev-gate" {
		t.Fatalf("correlation not preserved: %q", result.CorrelationID)
	}
}
