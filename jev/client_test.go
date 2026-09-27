package jev

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSystemOne(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != SystemOnePath { t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path) }
		if r.Header.Get("Authorization") != "Bearer test-key" { t.Fatalf("missing authorization") }
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
