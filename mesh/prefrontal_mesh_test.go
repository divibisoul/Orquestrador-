package mesh

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

func TestMeshRoutesIntoPrefrontalExecutiveAdmission(t *testing.T) {
	h := newTestGateway(t)
	if err := orchestrator.RegisterAdvancedOperations(h.Engine); err != nil {
		t.Fatal(err)
	}

	wire := canonicalRequest("request", "prefrontal.admission", "mesh-pfc-correlation", []float64{1, 2, 3, 4, 5, 6, 7, 8})
	wire["metadata"] = map[string]string{
		"candidate_json": `{"ID":"mesh-pfc-task","Cost":0.01,"Risk":0.01,"Urgency":0.5,"Impact":0.8}`,
		"task_id":       "mesh-pfc-task",
	}

	// Keep this unit test explicitly local and unauthenticated; production Mesh remains fail-closed.
	h.Secret = ""
	h.AllowUnauthenticatedLocal = true

	got, code := postWire(t, h, wire)
	if code != 200 {
		t.Fatalf("Mesh→Prefrontal request failed: code=%d envelope=%+v", code, got)
	}
	if got.CorrelationID != "mesh-pfc-correlation" {
		t.Fatalf("correlation was not preserved: %+v", got)
	}
	if got.Kind != "response" || got.Source != "N07" || got.Target != "N01" {
		t.Fatalf("unexpected Mesh response identity: %+v", got)
	}
	if got.Payload["status"] != "ok" {
		t.Fatalf("prefrontal admission did not execute successfully: %+v", got.Payload)
	}
	metadata, ok := got.Payload["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected decision metadata, got %#v", got.Payload["metadata"])
	}
	if metadata["decision_id"] != "mesh-pfc-task" {
		t.Fatalf("expected committed decision id, got %#v", metadata["decision_id"])
	}
	if metadata["neural_dimensions"] != "8" {
		t.Fatalf("expected 8 neural dimensions, got %#v", metadata["neural_dimensions"])
	}
}
