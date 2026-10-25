package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type capturedPeerRequest struct {
	ID            string `json:"id"`
	Nonce         string `json:"nonce"`
	CorrelationID string `json:"correlationId"`
}

func signedPeerResponse(secret, messageID, correlation, kind, capability string, payload map[string]any) (map[string]any, error) {
	messageType := "TASK_RESULT"
	if kind == "error" {
		messageType = "ERROR"
	}
	env := protocol.MeshEnvelope{
		Version:         protocol.SoulMeshVersion,
		ContractVersion: protocol.SoulMeshContractVersion,
		MessageID:       messageID,
		Source:          protocol.N01,
		Target:          protocol.N07,
		Timestamp:       time.Now().UnixMilli(),
		Nonce:           protocol.NewTraceID(),
		CorrelationID:   correlation,
		Type:            messageType,
		Payload: map[string]any{
			"capability": capability,
			"payload":    payload,
		},
	}
	if err := protocol.SignHMAC(&env, secret); err != nil {
		return nil, err
	}
	return map[string]any{
		"protocol":        "soul-mesh/1",
		"contractVersion": protocol.SoulMeshContractVersion,
		"id":              env.MessageID,
		"correlationId":   env.CorrelationID,
		"source":          env.Source,
		"target":          env.Target,
		"kind":            kind,
		"capability":      capability,
		"payload":         payload,
		"timestamp":       env.Timestamp,
		"nonce":            env.Nonce,
		"hmac":            env.HMAC,
	}, nil
}

func TestPeerClientAuthenticatesHTTPErrorBeforeTrustingStatus(t *testing.T) {
	const secret = "n07-auth-order-secret-0123456789"
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"unsigned temporary failure"}`))
	}))
	defer server.Close()

	t.Setenv("SOUL_MESH_N01_URL", server.URL)
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	p, err := NewPeerClient(&http.Client{Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	p.maxRetry = 3

	_, err = p.CallWithCorrelation(context.Background(), protocol.N01, "mesh.ping", map[string]any{"probe": true}, "auth-before-status-test")
	if err == nil || !strings.Contains(err.Error(), "HMAC credentials are missing") {
		t.Fatalf("expected unauthenticated response to fail before status retry, got %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one call because unsigned status is not trusted, got %d", got)
	}
}

func TestPeerClientPreservesOperationIDAcrossAuthenticatedRetry(t *testing.T) {
	const (
		secret      = "n07-stable-id-secret-0123456789"
		correlation = "stable-operation-correlation"
	)
	var calls atomic.Int32
	requests := make(chan capturedPeerRequest, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request capturedPeerRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request JSON", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()
		requests <- request

		attempt := calls.Add(1)
		kind := "response"
		status := http.StatusOK
		payload := map[string]any{"pong": true}
		if attempt == 1 {
			kind = "error"
			status = http.StatusServiceUnavailable
			payload = map[string]any{"error": "retryable"}
		}
		response, err := signedPeerResponse(secret, "n01-response-"+protocol.NewTraceID(), correlation, kind, "mesh.ping", payload)
		if err != nil {
			http.Error(w, "could not sign response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("SOUL_MESH_N01_URL", server.URL)
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	p, err := NewPeerClient(&http.Client{Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	p.maxRetry = 2

	result, err := p.CallWithCorrelation(context.Background(), protocol.N01, "mesh.ping", map[string]any{"probe": true}, correlation)
	if err != nil {
		t.Fatalf("expected authenticated retry to succeed, got %v", err)
	}
	if result["kind"] != "response" {
		t.Fatalf("expected response result, got %#v", result["kind"])
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected exactly two attempts, got %d", got)
	}

	first := <-requests
	second := <-requests
	if first.ID == "" || second.ID == "" {
		t.Fatalf("request message ID must be present on both attempts: first=%q second=%q", first.ID, second.ID)
	}
	if first.ID != second.ID {
		t.Fatalf("operation ID changed across retries: first=%q second=%q", first.ID, second.ID)
	}
	if first.Nonce == "" || second.Nonce == "" || first.Nonce == second.Nonce {
		t.Fatalf("each attempt must use a fresh nonce: first=%q second=%q", first.Nonce, second.Nonce)
	}
	if first.CorrelationID != correlation || second.CorrelationID != correlation {
		t.Fatalf("correlation ID did not survive retry: first=%q second=%q", first.CorrelationID, second.CorrelationID)
	}
}
