package mesh

import (
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// This contract-level test exercises a generated trace through the N07 Mesh
// gateway, neural.forward, prefrontal admission, Cortex decision, and response
// envelope. It does not claim a cross-process N01/N07 deployment test.
func TestMeshPrefrontalTraceCorrelationPreserved(t *testing.T) {
	h := newTestGateway(t)
	if err := orchestrator.RegisterAdvancedOperations(h.Engine); err != nil {
		t.Fatal(err)
	}

	traceID := "trace-pfc-" + protocol.NewTraceID()
	candidateID := "pfc-" + protocol.NewTraceID()
	candidate, err := json.Marshal(map[string]any{
		"ID": candidateID, "Cost": 0.01, "Risk": 0.01, "Urgency": 0.5, "Impact": 0.8,
	})
	if err != nil {
		t.Fatal(err)
	}

	wire := canonicalRequest("request", "prefrontal.admission", traceID, []float64{1, 2, 3, 4, 5, 6, 7, 8})
	wire["metadata"] = map[string]string{
		"candidate_json": string(candidate),
		"task_id":        candidateID,
	}

	// Authentication bypass is explicitly limited to this in-memory test gateway.
	// Production gateway defaults remain fail-closed without a configured HMAC secret.
	h.Secret = ""
	h.AllowUnauthenticatedLocal = true

	got, code := postWire(t, h, wire)
	if code != 200 {
		t.Fatalf("MESH_PREFRONTAL_ADMISSION_FAILED code=%d envelope=%+v", code, got)
	}
	if got.CorrelationID != traceID {
		t.Fatalf("MESH_TRACE_CORRELATION_LOST expected=%s got=%s", traceID, got.CorrelationID)
	}
	if got.Kind != "response" || got.Source != "N07" || got.Target != "N01" {
		t.Fatalf("MESH_RESPONSE_ENVELOPE_INVALID: %+v", got)
	}
	if got.Payload["status"] != "ok" {
		t.Fatalf("PREFRONTAL_DECISION_NOT_COMMITTED: %#v", got.Payload)
	}
	metadata, ok := got.Payload["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("PREFRONTAL_DECISION_METADATA_MISSING: %#v", got.Payload["metadata"])
	}
	if metadata["decision_id"] != candidateID {
		t.Fatalf("PREFRONTAL_DECISION_ID_MISMATCH expected=%s got=%#v", candidateID, metadata["decision_id"])
	}
	if metadata["neural_dimensions"] != "8" {
		t.Fatalf("NEURAL_SIGNAL_DIMENSIONS_MISMATCH: %#v", metadata["neural_dimensions"])
	}
	t.Logf("REAL_TEST_TRACE_ID=%s", traceID)
	t.Logf("MESH_HOPS_PASS=N01-envelope>N07-gateway>neural.forward>prefrontal.admission>Cortex-commit>N01-response")
}
