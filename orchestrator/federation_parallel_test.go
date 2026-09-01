package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFederationDelegatePreservesArbitraryPayload(t *testing.T) {
	t.Setenv("SOUL_MESH_HMAC_SECRET", "")
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload := in["payload"].(map[string]any)
		if payload["document"] != "soul" {
			t.Fatalf("arbitrary document payload was not preserved: %#v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"correlationId": in["correlationId"],
			"source": "N01",
			"target": "N07",
			"payload": map[string]any{"document": "validated"},
		})
	}))
	defer server.Close()

	f := NewFederation()
	if err := f.RegisterPeer("N01", server.URL, []string{"document.validate@1.0.0"}); err != nil {
		t.Fatal(err)
	}
	result, err := f.Delegate(context.Background(), "trace-arbitrary", "document.validate", map[string]any{"document": "soul", "options": map[string]any{"strict": true}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result["payload"].(map[string]any)
	if payload["document"] != "validated" {
		t.Fatalf("unexpected delegated response: %#v", result)
	}
}

func TestFederationExecuteParallelAggregatesIndependentTasks(t *testing.T) {
	t.Setenv("SOUL_MESH_HMAC_SECRET", "")
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"correlationId": in["correlationId"],
			"source": "N02",
			"target": "N07",
			"payload": map[string]any{"ok": true},
		})
	}))
	defer server.Close()

	f := NewFederation()
	if err := f.RegisterPeer("N02", server.URL, []string{"task.a", "task.b", "task.c"}); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	results, err := f.ExecuteParallel(context.Background(), "trace-parallel", []FederatedTask{
		{ID: "a", Capability: "task.a", Payload: map[string]any{"n": 1}, Required: true},
		{ID: "b", Capability: "task.b", Payload: map[string]any{"n": 2}, Required: true},
		{ID: "c", Capability: "task.c", Payload: map[string]any{"n": 3}, Required: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, result := range results {
		if result.Status != "ok" || result.Source != "N02" {
			t.Fatalf("unexpected result: %#v", result)
		}
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("parallel federation exceeded sanity bound")
	}
}
