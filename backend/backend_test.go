package backend

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testCID = "bafybeibhybbpoqakv7pfj5nlrpmldkgiuksmbi3t2cnhxqxnqvbzkhyzjy"

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
			_, _ = io.WriteString(w, `{"cid":"`+testCID+`"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/status/"+testCID:
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"cid":"`+testCID+`","status":"pinned"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	s := NewWeb3Storage(Config{
		Web3StorageURL:   server.URL,
		Web3StorageToken: "secret",
		IPFSGatewayURL:   server.URL + "/ipfs",
	})
	cid, size, err := s.Upload(context.Background(), strings.NewReader("hello"), "artifact.bin")
	if err != nil || cid != testCID || size != 5 {
		t.Fatalf("unexpected upload result cid=%q size=%d err=%v", cid, size, err)
	}
	if gotAuth != "Bearer secret" || gotName != "artifact.bin" {
		t.Fatalf("unexpected upload headers auth=%q name=%q", gotAuth, gotName)
	}
	status, err := s.Status(context.Background(), cid)
	if err != nil || status["status"] != "pinned" {
		t.Fatalf("unexpected status: %#v err=%v", status, err)
	}
	if got := s.ObjectURL(cid); got != server.URL+"/ipfs/"+testCID {
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


func TestServerExposesArtifactPersistenceFailure(t *testing.T) {
	storageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "{\"cid\":\""+testCID+"\"}")
	}))
	defer storageServer.Close()

	supabaseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "persistence unavailable")
	}))
	defer supabaseServer.Close()

	server := New(nil, Config{
		AppToken:               "app-token",
		Web3StorageURL:         storageServer.URL,
		Web3StorageToken:       "storage-token",
		IPFSGatewayURL:         storageServer.URL + "/ipfs",
		SupabaseURL:            supabaseServer.URL,
		SupabaseServiceKey:     "service-key",
		SupabaseRunsTable:      "n07_runs",
		SupabaseArtifactsTable: "n07_artifacts",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/storage/upload", strings.NewReader("hello"))
	req.Header.Set("Authorization", "Bearer app-token")
	req.Header.Set("X-Filename", "artifact.txt")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload operation unexpectedly failed: status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	persistence, ok := response["persistence"].(map[string]any)
	if !ok || persistence["state"] != "BLOCKED" {
		t.Fatalf("artifact persistence failure was hidden: %#v", response)
	}
	if errText, _ := persistence["error"].(string); !strings.Contains(errText, "persistence unavailable") {
		t.Fatalf("persistence error evidence missing: %#v", persistence)
	}
}

func TestServerRejectsNonGETStorageReads(t *testing.T) {
	server := New(nil, Config{AppToken: "app-token"})
	for _, path := range []string{"/v1/storage/status/" + testCID, "/v1/storage/object/" + testCID} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Authorization", "Bearer app-token")
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 for %s, got %d", path, rec.Code)
		}
	}
}
