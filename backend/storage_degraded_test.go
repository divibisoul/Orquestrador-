package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWeb3StorageMissingUCANReportsDegradedState(t *testing.T) {
	t.Setenv("WEB3_STORAGE_TOKEN", "")
	t.Setenv("STORACHA_SPACE", "")
	t.Setenv("STORACHA_MODE", "storacha")

	s := NewWeb3Storage(DefaultConfig())
	status := s.Status()
	if status["status"] != "degraded" || status["reason"] != "missing_ucan" || status["code"] != "STORAGE_NOT_CONFIGURED" {
		t.Fatalf("unexpected storage status: %#v", status)
	}
}

func TestServerStorageEndpointsReturn503WhenUCANMissing(t *testing.T) {
	t.Setenv("N07_APP_TOKEN", "test-app-token")
	t.Setenv("WEB3_STORAGE_TOKEN", "")
	t.Setenv("STORACHA_SPACE", "")
	t.Setenv("STORACHA_MODE", "storacha")

	srv := New(nil, DefaultConfig())
	h := srv.Handler()

	checks := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/v1/storage/upload"},
		{http.MethodGet, "/v1/storage/status/bafybeigdyrzt4example"},
		{http.MethodGet, "/v1/storage/object/bafybeigdyrzt4example"},
	}

	for _, check := range checks {
		req := httptest.NewRequest(check.method, check.path, nil)
		req.Header.Set("Authorization", "Bearer test-app-token")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s %s: expected 503, got %d body=%s", check.method, check.path, rec.Code, rec.Body.String())
		}
	}
}

func TestStorageStatusDoesNotRequireNetworkWhenMissingUCAN(t *testing.T) {
	t.Setenv("WEB3_STORAGE_TOKEN", "")
	t.Setenv("STORACHA_SPACE", "")
	t.Setenv("STORACHA_MODE", "storacha")

	s := NewWeb3Storage(DefaultConfig())
	_, err := s.StatusForCID(context.Background(), "bafybeigdyrzt4example")
	if err == nil || err.Error() != "STORAGE_NOT_CONFIGURED" {
		t.Fatalf("expected STORAGE_NOT_CONFIGURED, got %v", err)
	}
}
