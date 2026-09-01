package mesh

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	sg "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestEnhancedFederatedGatewayParallelExecution(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")

	var active int32
	var maxActive int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request protocol.MeshEnvelope
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode peer request: %v", err)
		}
		if err := protocol.VerifyHMAC(request, secret, time.Now()); err != nil {
			t.Fatalf("peer received invalid request HMAC: %v", err)
		}
		current := atomic.AddInt32(&active, 1)
		for {
			old := atomic.LoadInt32(&maxActive)
			if current <= old || atomic.CompareAndSwapInt32(&maxActive, old, current) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&active, -1)
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
			Payload:         map[string]any{"value": request.CorrelationID},
		}
		if err := protocol.SignHMAC(&response, secret); err != nil {
			t.Fatalf("sign peer response: %v", err)
		}
		_ = json.NewEncoder(w).Encode(response)
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
	engine, err := orchestrator.New(n, c, sg.New(nil))
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
		Nonce:            protocol.NewTraceID(),
		CorrelationID:   "trace-supergpu",
		Type:            "CAPABILITY_REQUEST",
		Payload: map[string]any{
			"capability": "mesh.supergpu.parallel",
			"payload": map[string]any{
				"tasks": []map[string]any{{
					"id":         "task-alpha",
					"capability": "task.alpha",
					"payload":    map[string]any{"document": "alpha"},
					"required":   true,
				}},
			},
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
		t.Fatalf("expected federated parallel execution to succeed, got %d: %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	payload, ok := out["payload"].(map[string]any)
	if !ok || payload["execution"] != "federated-supergpu-parallel" {
		t.Fatalf("unexpected federated response: %#v", out)
	}
	if out["correlationId"] != "trace-supergpu" {
		t.Fatalf("parent correlation was not preserved: %#v", out)
	}
	if atomic.LoadInt32(&maxActive) != 1 {
		t.Fatalf("single-task path should use one peer call, observed max concurrency %d", maxActive)
	}
}

func TestEnhancedFederatedGatewayRejectsDuplicateTaskIDs(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	n, err := neural.New(8, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := orchestrator.New(n, c, sg.New(nil))
	if err != nil {
		t.Fatal(err)
	}
	gateway := NewEnhancedFederatedHTTPGateway(engine)
	request := protocol.MeshEnvelope{
		Version:         protocol.SoulMeshVersion,
		ContractVersion: protocol.SoulMeshContractVersion,
		MessageID:       protocol.NewTraceID(),
		Source:          protocol.N01,
		Target:          protocol.N07,
		Timestamp:       time.Now().UnixMilli(),
		Nonce:            protocol.NewTraceID(),
		CorrelationID:   "trace-duplicate",
		Type:            "CAPABILITY_REQUEST",
		Payload: map[string]any{
			"capability": "mesh.supergpu.parallel",
			"payload": map[string]any{
				"tasks": []map[string]any{
					{"id": "same", "capability": "task.a", "payload": map[string]any{}},
					{"id": "same", "capability": "task.b", "payload": map[string]any{}},
				},
			},
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
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate IDs to be rejected, got %d: %s", rec.Code, rec.Body.String())
	}
}
