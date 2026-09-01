package mesh

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestEnhancedFederatedGatewayParallelExecution(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")

	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request protocol.MeshEnvelope
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode peer request: %v", err)
		}
		if request.CorrelationID == "" {
			t.Fatal("peer request lost correlation id")
		}

		var payload map[string]any
		if request.Capability() == "mesh.discovery" || request.Capability() == "mesh.describe" {
			payload = map[string]any{"operations": []string{"task.alpha"}}
		} else {
			payload = map[string]any{"document": "validated"}
		}
		response := protocol.MeshEnvelope{
			Version:         protocol.SoulMeshVersion,
			ContractVersion: protocol.SoulMeshContractVersion,
			MessageID:       protocol.NewTraceID(),
			Source:          protocol.N01,
			Target:          protocol.N07,
			Timestamp:       time.Now().UnixMilli(),
			Nonce:           protocol.NewTraceID(),
			CorrelationID:   request.CorrelationID,
			Type:            "TASK_RESULT",
			Payload:         payload,
		}
		if err := protocol.SignHMAC(&response, secret); err != nil {
			t.Fatalf("sign peer response: %v", err)
		}
		body := map[string]any{
			"protocol":        "soul-mesh/1",
			"contractVersion": response.ContractVersion,
			"id":              response.MessageID,
			"correlationId":   response.CorrelationID,
			"source":          response.Source,
			"target":          response.Target,
			"kind":            "response",
			"capability":      request.Capability(),
			"payload":         response.Payload,
			"timestamp":       response.Timestamp,
			"nonce":           response.Nonce,
			"hmac":            response.HMAC,
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer peer.Close()

	n, err := neural.New(8, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := orchestrator.New(n, c, supergpu.New(nil))
	if err != nil {
		t.Fatal(err)
	}
	gateway := NewEnhancedFederatedHTTPGateway(engine)
	gateway.base.peers.mu.Lock()
	gateway.base.peers.peers[protocol.N01] = PeerInfo{Nucleus: protocol.N01, URL: peer.URL, Circuit: CircuitClosed}
	gateway.base.peers.mu.Unlock()

	request := protocol.MeshEnvelope{
		Version:         protocol.SoulMeshVersion,
		ContractVersion: protocol.SoulMeshContractVersion,
		MessageID:       protocol.NewTraceID(),
		Source:          protocol.N01,
		Target:          protocol.N07,
		Timestamp:       time.Now().UnixMilli(),
		Nonce:           protocol.NewTraceID(),
		CorrelationID:   "trace-supergpu",
		Type:            "CAPABILITY_REQUEST",
		Payload: map[string]any{
			"capability": "task.alpha",
			"payload":    map[string]any{"document": "alpha"},
		},
	}
	if err := protocol.SignHMAC(&request, secret); err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected federated execution to succeed, got %d: %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["correlationId"] != "trace-supergpu" {
		t.Fatalf("parent correlation was not preserved: %#v", out)
	}
	payload, ok := out["payload"].(map[string]any)
	if !ok || payload["execution"] != "federated-supergpu-parallel" {
		t.Fatalf("expected parallel federated execution response: %#v", out)
	}
}
