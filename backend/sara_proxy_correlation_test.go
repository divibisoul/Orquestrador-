package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSARAProxyForwardsCorrelationHeader(t *testing.T) {
	const correlation = "corr-sara-proxy-001"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Correlation-ID"); got != correlation {
			t.Fatalf("X-Correlation-ID=%q want %q", got, correlation)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{"status":"ok","cycle_id":"cycle-001"}"))
	}))
	defer server.Close()

	proxy := NewSARAProxy(Config{
		SARAServiceURL:     server.URL,
		SARAServiceToken:   "test-token",
		SARARequestTimeout: 5 * time.Second,
	})

	out, err := proxy.Cycle(context.Background(), "input", "cycle-001", correlation)
	if err != nil {
		t.Fatal(err)
	}
	if out["cycle_id"] != "cycle-001" {
		t.Fatalf("unexpected response: %#v", out)
	}
}
