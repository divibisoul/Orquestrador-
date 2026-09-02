package backend

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	status, err := s.StatusForCID(context.Background(), cid)
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

func TestStorachaUploadUsesConfiguredSpaceAndFilename(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "guppy")
	cid := testCID
	script := "#!/bin/sh\nprintf '%s\\n' '" + cid + "'\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("STORACHA_MODE", "storacha")
	t.Setenv("STORACHA_UCAN", "unit-test-ucan")
	t.Setenv("STORACHA_GUPPY_BIN", bin)
	t.Setenv("STORACHA_SPACE", "did:key:z6Mkhunit-test-space")
	t.Setenv("STORACHA_DATA_DIR", filepath.Join(t.TempDir(), "state"))
	t.Setenv("STORACHA_GATEWAY_URL", "https://storacha.link/ipfs")

	s := NewWeb3Storage(Config{MaxUploadBytes: 1024})
	gotCID, size, err := s.Upload(context.Background(), strings.NewReader("hello"), "nested/artifact.bin")
	if err != nil {
		t.Fatal(err)
	}
	if gotCID != cid || size != 5 {
		t.Fatalf("unexpected Storacha upload cid=%q size=%d", gotCID, size)
	}
	if got := s.ObjectURL(cid); got != "https://storacha.link/ipfs/"+cid {
		t.Fatalf("unexpected Storacha object URL: %q", got)
	}
}

func TestStorachaUploadEnforcesConfiguredLimit(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "guppy")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STORACHA_MODE", "storacha")
	t.Setenv("STORACHA_UCAN", "unit-test-ucan")
	t.Setenv("STORACHA_GUPPY_BIN", bin)
	t.Setenv("STORACHA_SPACE", "did:key:z6Mkhunit-test-space")

	s := NewWeb3Storage(Config{MaxUploadBytes: 4})
	if _, _, err := s.Upload(context.Background(), strings.NewReader("hello"), "x"); err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("expected size limit error, got %v", err)
	}
}

func TestCIDValidationRejectsUnsafeValues(t *testing.T) {
	for _, cid := range []string{"", "../secret", "https://example.com/ipfs/" + testCID, "bafy-invalid"} {
		if validCID(cid) {
			t.Fatalf("expected invalid CID: %q", cid)
		}
	}
	if !validCID(testCID) {
		t.Fatal("expected fixture CID to be valid")
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
