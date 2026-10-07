package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbePeerAcceptsHealthyDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mesh/discovery" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	status, latency, err := probePeer(ctx, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("probe failed: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d want %d", status, http.StatusOK)
	}
	if latency < 0 {
		t.Fatalf("latency=%d", latency)
	}
}

func TestProbePeerRejectsUnavailableEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, _, err := probePeer(ctx, &http.Client{Timeout: 10 * time.Millisecond}, "http://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected unavailable endpoint to fail")
	}
}
