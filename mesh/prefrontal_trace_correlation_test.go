package mesh

import (
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"

	"github.com/divibisoul/Orquestrador-/orchestrator"
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

	correlationID := "corr-pfc-" + protocol.NewTraceID()
	candidateID := "pfc-" + protocol.NewTraceID()
	candidate, err := json.Marshal(map[string]any{
		"ID": candidateID, "Cost": 0.01, "Risk": 0.01, "Urgency": 0.5, "Impact": 0.8,
	})
	if err != nil {
		t.Fatal(err)
	}

	wire := canonicalRequest("request", "prefrontal.admission", correlationID, []float64{1, 2, 3, 4, 5, 6, 7, 8})
	traceID := wire["id"].(string)
	wire["metadata"] = map[string]string{
		"trace_id":       traceID,
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
	if got.CorrelationID != correlationID {
		t.Fatalf("MESH_CORRELATION_LOST expected=%s got=%s", correlationID, got.CorrelationID)
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
	t.Logf("REAL_TEST_TRACE_ID=%s CORRELATION_ID=%s", traceID, correlationID)
	t.Logf("MESH_HOPS_PASS=N01-envelope>N07-gateway>neural.forward>prefrontal.admission>Cortex-commit>N01-response")
}

func TestMeshPrefrontalAgentRoutePreservesTrace(t *testing.T) {
	t.Setenv("N07_TCE_ENABLED", "true")
	h := newTestGateway(t)
	if err := orchestrator.RegisterOrbitalReasoningOperations(h.Engine); err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterSuperpowersAgentOperations(h.Engine); err != nil {
		t.Fatal(err)
	}

	correlationID := "corr-pfc-agent-" + protocol.NewTraceID()
	candidateID := "pfc-agent-" + protocol.NewTraceID()
	workloads, err := json.Marshal([]core.Workload{{
		ID: "pfc-agent-workload", Operation: "reasoning-simulation", Precision: core.FP16,
		MatrixSize: 256, BatchSize: 1, DataBytes: 4096, MemoryNeeded: 1, Priority: 1,
	}})
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := json.Marshal(map[string]any{
		"ID": candidateID, "Cost": 0.1, "Risk": 0.1, "Utility": 0.9,
		"Uncertainty": 0.1, "Urgency": 0.2, "Impact": 0.2,
	})
	if err != nil {
		t.Fatal(err)
	}
	wire := canonicalRequest("request", orchestrator.SuperpowersAgentRouteOperation, correlationID, []float64{1, 2, 3, 4, 5, 6, 7, 8})
	traceID := wire["id"].(string)
	wire["metadata"] = map[string]string{
		"trace_id":       traceID,
		"agent_id":       "superpowers.cortex-orbital-supergpu",
		"operation":      "prefrontal.orbital.evaluate@1.0.0",
		"workloads_json": string(workloads),
		"candidate_json": string(candidate),
	}

	got, code := postWire(t, h, wire)
	if code != 200 {
		t.Fatalf("MESH_PREFrONTAL_AGENT_ROUTE_FAILED code=%d envelope=%+v", code, got)
	}
	if got.CorrelationID != correlationID {
		t.Fatalf("MESH_AGENT_CORRELATION_LOST expected=%s got=%s", correlationID, got.CorrelationID)
	}
	if got.Payload["status"] != "ok" {
		t.Fatalf("PREFRONTAL_AGENT_DECISION_FAILED: %#v", got.Payload)
	}
	metadata, ok := got.Payload["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("PREFRONTAL_AGENT_METADATA_MISSING: %#v", got.Payload["metadata"])
	}
	if metadata["decision_id"] != candidateID {
		t.Fatalf("PREFRONTAL_AGENT_DECISION_ID_MISMATCH expected=%s got=%#v", candidateID, metadata["decision_id"])
	}
	if metadata["superpowers_agent"] != "superpowers.cortex-orbital-supergpu" {
		t.Fatalf("PREFRONTAL_AGENT_NOT_IDENTIFIED: %#v", metadata["superpowers_agent"])
	}
	if metadata["simulation_is_not_hardware"] != "true" {
		t.Fatalf("ORBITAL_SIMULATION_BOUNDARY_MISSING: %#v", metadata)
	}
	if metadata["parent_trace_id"] != traceID {
		t.Fatalf("AGENT_PARENT_TRACE_NOT_PRESERVED expected=%s got=%#v", traceID, metadata["parent_trace_id"])
	}
	delegatedTrace, ok := metadata["delegated_trace_id"].(string)
	if !ok || delegatedTrace == "" || delegatedTrace == traceID {
		t.Fatalf("AGENT_CHILD_TRACE_NOT_UNIQUE parent=%s child=%#v", traceID, metadata["delegated_trace_id"])
	}
	if metadata["delegated_correlation_id"] != correlationID {
		t.Fatalf("AGENT_CORRELATION_NOT_PRESERVED expected=%s got=%#v", correlationID, metadata["delegated_correlation_id"])
	}
	t.Logf("REAL_AGENT_MESH_TRACE_ID=%s DELEGATED_TRACE_ID=%s CORRELATION_ID=%s AGENT=%s", traceID, delegatedTrace, correlationID, metadata["superpowers_agent"])
	t.Logf("AGENT_HOPS_PASS=N01-envelope>N07-Mesh>superpowers.agent.route>prefrontal.orbital.evaluate>neural.forward>Cortex-commit>N01-response")
}
