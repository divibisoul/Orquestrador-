package jev

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSystemOne(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != SystemOnePath { t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path) }
		if r.Header.Get("Authorization") != "Bearer test-key" { t.Fatalf("missing authorization") }
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { t.Fatalf("invalid request: %v", err) }
		if req.Model != "jev-latest" { t.Fatalf("unexpected model: %q", req.Model) }
		if req.Questions["urgent"]["type"] != "noul" { t.Fatalf("unexpected question: %#v", req.Questions) }
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-1.13.0","answers":{"urgent":{"type":"noul","noul":0.91}},"usage":{"input_tokens":10}}`))
	}))
	defer server.Close()
	client := &Client{BaseURL:server.URL, APIKey:"test-key", Model:"jev-latest", HTTPClient:server.Client()}
	out, err := client.SystemOne(context.Background(), "incident", map[string]map[string]any{"urgent":{"type":"noul","instructions":"Is this urgent?"}})
	if err != nil { t.Fatal(err) }
	if out.Answers["urgent"]["noul"] != 0.91 { t.Fatalf("unexpected answer: %#v", out.Answers) }
}

func TestSystemOneRejectsMissingQuestions(t *testing.T) {
	client := &Client{APIKey:"test-key", BaseURL:"http://unused"}
	if _, err := client.SystemOne(context.Background(), "state", nil); err == nil { t.Fatal("expected validation error") }
}


func TestSystemOneRetriesRateLimit(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-1.13.0","answers":{"route":{"type":"choice","choice":"technical","probabilities":{"technical":1},"confidence":1}}}`))
	}))
	defer server.Close()
	client := &Client{BaseURL: server.URL, APIKey: "test-key", Model: "jev-latest", HTTPClient: server.Client(), MaxRetries: 1, Backoff: time.Millisecond}
	out, err := client.SystemOne(context.Background(), "state", map[string]map[string]any{"route":{"type":"choice","instructions":"route","criteria":map[string]string{"technical":"technical"}}})
	if err != nil { t.Fatal(err) }
	if calls != 2 { t.Fatalf("expected retry, got %d calls", calls) }
	if out.Model != "jev-1.13.0" { t.Fatalf("unexpected model: %q", out.Model) }
}
