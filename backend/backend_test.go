package backend

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeb3StorageUploadAndStatus(t *testing.T) {
	var gotAuth string
	var gotName string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotName = r.Header.Get("X-Name")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/upload":
			body, _ := io.ReadAll(r.Body)
			if string(body) != "hello" {
				t.Fatalf("unexpected upload body: %q", string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"cid":"bafy-test"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/status/bafy-test":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"cid":"bafy-test","status":"pinned"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	s := NewWeb3Storage(Config{
		Web3StorageURL: server.URL,
		Web3StorageToken: "secret",
		IPFSGatewayURL: server.URL + "/ipfs",
	})
	cid, size, err := s.Upload(context.Background(), strings.NewReader("hello"), "artifact.bin")
	if err != nil || cid != "bafy-test" || size != 5 {
		t.Fatalf("unexpected upload result cid=%q size=%d err=%v", cid, size, err)
	}
	if gotAuth != "Bearer secret" || gotName != "artifact.bin" {
		t.Fatalf("unexpected upload headers auth=%q name=%q", gotAuth, gotName)
	}
	status, err := s.Status(context.Background(), cid)
	if err != nil || status["status"] != "pinned" {
		t.Fatalf("unexpected status: %#v err=%v", status, err)
	}
	if got := s.ObjectURL(cid); got != server.URL+"/ipfs/bafy-test" {
		t.Fatalf("unexpected object URL: %q", got)
	}
}

func TestWeb3StorageRejectsEmptyAndUnavailable(t *testing.T) {
	s := NewWeb3Storage(Config{})
	if _, _, err := s.Upload(context.Background(), strings.NewReader("x"), "x"); err == nil {
		t.Fatal("expected unconfigured storage to fail")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "bad upload")
	}))
	defer server.Close()
	s = NewWeb3Storage(Config{Web3StorageURL: server.URL, Web3StorageToken: "secret"})
	if _, _, err := s.Upload(context.Background(), nil, "x"); err == nil {
		t.Fatal("expected nil body to fail")
	}
	if _, _, err := s.Upload(context.Background(), strings.NewReader("x"), "x"); err == nil {
		t.Fatal("expected remote upload error")
	}
}

func TestWeb3StorageHonorsCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	s := NewWeb3Storage(Config{Web3StorageURL: server.URL, Web3StorageToken: "secret"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := s.Upload(ctx, strings.NewReader("x"), "x")
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestSupabaseRecordRun(t *testing.T) {
	var gotAPIKey, gotAuthorization, gotPrefer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("apikey")
		gotAuthorization = r.Header.Get("Authorization")
		gotPrefer = r.Header.Get("Prefer")
		if r.Method != http.MethodPost || r.URL.Path != "/rest/v1/n07_runs" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	s := NewSupabaseStore(Config{SupabaseURL: server.URL, SupabaseServiceKey: "service-key", SupabaseRunsTable: "n07_runs"})
	if err := s.RecordRun(context.Background(), map[string]any{"trace_id": "t1", "status": "ok"}); err != nil {
		t.Fatal(err)
	}
	if gotAPIKey != "service-key" || gotAuthorization != "Bearer service-key" || gotPrefer != "return=minimal" {
		t.Fatalf("unexpected Supabase headers: apikey=%q auth=%q prefer=%q", gotAPIKey, gotAuthorization, gotPrefer)
	}
}

func TestSupabaseValidation(t *testing.T) {
	s := NewSupabaseStore(Config{})
	if err := s.RecordRun(context.Background(), nil); err == nil {
		t.Fatal("expected nil run row to fail")
	}
	if err := s.RecordArtifact(context.Background(), map[string]any{"x": 1}); err == nil {
		t.Fatal("expected unconfigured store to fail")
	}
}
