package octacore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestExecutiveCoreUsesSuperAGIMeshCapability(t *testing.T) {
	const secret = "octacore-superagi-mesh-secret-2026"
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wire struct {
			ID            string         `json:"id"`
			CorrelationID string         `json:"correlationId"`
			Capability    string         `json:"capability"`
			Payload       map[string]any `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		payload := map[string]any{}
		switch strings.TrimSpace(wire.Capability) {
		case "mlfg":
			payload = map[string]any{
				"task_count": 1,
				"meta_budget": 2,
				"hypotheses": []any{map[string]any{"id": "superagi-h1", "kind": "bounded-meta-learning"}},
				"evaluation_protocol": []any{map[string]any{"step": "evaluate"}},
				"selection_rules": []any{"preserve evidence"},
			}
		case "mesh.discovery", "mesh.describe":
			payload = map[string]any{
				"executableCapabilities": []any{"mesh.discovery", "mesh.describe", "mlfg"},
			}
		default:
			payload = map[string]any{"error": "unsupported test capability"}
		}

		correlation := wire.CorrelationID
		if correlation == "" {
			correlation = protocol.NewTraceID()
		}
		env := protocol.MeshEnvelope{
			Version:         protocol.SoulMeshVersion,
			ContractVersion: protocol.SoulMeshContractVersion,
			MessageID:       protocol.NewTraceID(),
			Source:          protocol.N02,
			Target:          protocol.N07,
			Timestamp:       time.Now().UnixMilli(),
			Nonce:           protocol.NewTraceID(),
			CorrelationID:   correlation,
			Type:            "TASK_RESULT",
			Payload: map[string]any{
				"capability": wire.Capability,
				"payload":    payload,
			},
		}
		if err := protocol.SignHMAC(&env, secret); err != nil {
			t.Fatalf("sign test Mesh response: %v", err)
		}
		response := map[string]any{
			"protocol":        "soul-mesh/1",
			"contractVersion": protocol.SoulMeshContractVersion,
			"id":              env.MessageID,
			"correlationId":   correlation,
			"source":          env.Source,
			"target":          env.Target,
			"kind":            "response",
			"capability":      wire.Capability,
			"payload":         payload,
			"timestamp":       env.Timestamp,
			"nonce":           env.Nonce,
			"hmac":            env.HMAC,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("SOUL_MESH_N02_URL", server.URL)
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	t.Setenv("N07_TCE_ENABLED", "true")

	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	cortex, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	gpu := supergpu.New(nil)
	gpu.Discover()
	engine, err := orchestrator.New(n, cortex, gpu)
	if err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterSuperGPUOperations(engine); err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterAdvancedOperations(engine); err != nil {
		t.Fatal(err)
	}
	tce, err := orchestrator.NewTranscendentalSimulatorFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterOrbitalReasoningOperationsWithSimulator(engine, tce); err != nil {
		t.Fatal(err)
	}
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	neocortex, err := prefrontal.NewNeocortex(cortex, n)
	if err != nil {
		t.Fatal(err)
	}
	coreProcessor, err := NewExecutiveCore(engine, neocortex, tce, gpu, peers)
	if err != nil {
		t.Fatal(err)
	}

	req := ExecutiveCoreRequest{
		JobID:         "superagi-job",
		CorrelationID: "superagi-correlation",
		TaskID:        "superagi-task",
		Operation:     "identity",
		Input:         []float64{1, 2, 3, 4},
		Workloads: []core.Workload{{
			ID: "superagi-workload",
			Operation: "identity",
			Precision: core.FP32,
			MatrixSize: 128,
			BatchSize: 1,
			DataBytes: 2048,
			MemoryNeeded: 1,
		}},
		Risk: 0,
		Cost: 0,
		Urgency: 0.2,
		Impact: 0.2,
		Collaborators: []AgentCollaboration{{
			Capability: "mlfg",
			Payload: map[string]any{
				"tasks": []any{map[string]any{"operation": "identity", "evidence": "test"}},
				"meta_budget": 2,
			},
			Required: true,
		}},
	}
	result, err := coreProcessor.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != "READY" {
		t.Fatalf("unexpected core state: %s", result.State)
	}
	if len(result.Agents) != 1 {
		t.Fatalf("expected one SuperAGI agent contribution, got %d", len(result.Agents))
	}
	agent := result.Agents[0]
	if agent.State != "REAL" || agent.Target != protocol.N02 {
		t.Fatalf("SuperAGI Mesh contribution was not real N02 execution: %#v", agent)
	}
	if agent.Output["payload"] == nil {
		t.Fatalf("expected Mesh payload from mlfg contribution: %#v", agent.Output)
	}
}
