package mesh

import (
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
		payload := map[string]any{
			"capability": request.Capability(),
			"payload":    request.NestedPayload(),
			"status":     "ok",
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
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(0.10, 32)
	if err != nil { t.Fatal(err) }
	g, err := orchestrator.New(n, c, supergpu.New(nil))
	if err != nil { t.Fatal(err) }
	gateway := NewEnhancedFederatedHTTPGateway(g)
	federated := gateway.base.peers
	if _, err := federated.ConfiguredPeers(), error(nil); err != nil { t.Fatal(err) }

	federated.mu.Lock()
	federated.peers[protocol.N01] = PeerInfo{Nucleus: protocol.N01, URL: peer.URL, Circuit: CircuitClosed}
	federated.mu.Unlock()

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
	if err := protocol.SignHMAC(&request, secret); err != nil { t.Fatal(err) }

	body, err := json.Marshal(request)
	if err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusOK && rec.Code != http.StatusBadGateway { t.Fatalf("unexpected status: %d", rec.Code) }
}
