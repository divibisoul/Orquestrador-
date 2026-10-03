package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExternalProviderStatusVLLMLive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"test-model"}]}`))
	}))
	defer srv.Close()
	t.Setenv("SOUL_VLLM_URL", srv.URL)
	status, err := externalProviderStatus(context.Background(), "vllm")
	if err != nil {
		t.Fatal(err)
	}
	if status["state"] != "LIVE" || status["live"] != true {
		t.Fatalf("status=%#v", status)
	}
}

func TestExternalProviderInvokeSGLangUsesOpenAIContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["model"] != "test-model" {
			t.Fatalf("model=%v", body["model"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer srv.Close()
	t.Setenv("SOUL_SGLANG_URL", srv.URL)
	cfg := externalProviderRuntimeConfig("sglang")
	response, err := externalOpenAIInvoke(context.Background(), cfg, "test-model", "hello", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(response["choices"].([]any)) != 1 {
		t.Fatalf("response=%#v", response)
	}
}

func TestExternalProviderStatusFailsClosedWhenUnset(t *testing.T) {
	t.Setenv("SOUL_VLLM_URL", "")
	status, err := externalProviderStatus(context.Background(), "vllm")
	if err != nil {
		t.Fatal(err)
	}
	if status["state"] != "BLOCKED" || status["code"] != "EXTERNAL_PROVIDER_ENDPOINT_NOT_CONFIGURED" {
		t.Fatalf("status=%#v", status)
	}
}

func TestExternalSwarmClawInvokeUsesA2AAndPollsRealTask(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/a2a" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer swarm-token" {
			t.Fatalf("missing bearer authentication")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			if body["method"] != "executeTask" {
				t.Fatalf("method=%v", body["method"])
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":{"taskId":"task-1","status":"submitted"}}`))
			return
		}
		if body["method"] != "getStatus" {
			t.Fatalf("method=%v", body["method"])
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"2","result":{"taskId":"task-1","status":"completed","result":"ok"}}`))
	}))
	defer srv.Close()
	t.Setenv("SOUL_SWARMCLAW_URL", srv.URL)
	t.Setenv("SOUL_SWARMCLAW_TOKEN", "swarm-token")

	cfg := externalProviderRuntimeConfig("swarmclaw")
	result, err := externalSwarmClawInvoke(context.Background(), cfg, "do work", "agent-1", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d want=2", calls)
	}
	if result["status"] != "completed" {
		t.Fatalf("result=%#v", result)
	}
}
