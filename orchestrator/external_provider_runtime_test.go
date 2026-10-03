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
