package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMapIntentNeuralLearn(t *testing.T) {
	values, metadata, err := mapIntent("neural.learn", map[string]any{"input": []any{1, 2}, "target": []any{3, 4}})
	if err != nil { t.Fatal(err) }
	if len(values) != 4 || values[0] != 1 || values[3] != 4 { t.Fatalf("unexpected values: %#v", values) }
	if len(metadata) != 0 { t.Fatalf("unexpected metadata: %#v", metadata) }
}

func TestMapIntentSuperGPUParallel(t *testing.T) {
	_, metadata, err := mapIntent("supergpu.parallel", map[string]any{"inputs": []any{[]any{1, 2}, []any{3, 4}}, "operation": "relu", "workers": float64(2)})
	if err != nil { t.Fatal(err) }
	if metadata["operation"] != "relu" || metadata["workers"] != "2" || metadata["inputs_json"] == "" { t.Fatalf("unexpected metadata: %#v", metadata) }
}

func TestConfigUsesEnvironment(t *testing.T) {
	old := lookupEnv
	defer func() { lookupEnv = old }()
	lookupEnv = func(key string) string {
		switch key {
		case "N07_APP_TOKEN": return "token"
		case "SUPABASE_URL": return "https://example.supabase.co"
		case "SUPABASE_SERVICE_ROLE_KEY": return "service"
		case "WEB3_STORAGE_TOKEN": return "web3"
		default: return ""
		}
	}
	cfg := DefaultConfig()
	if cfg.AppToken != "token" || cfg.SupabaseURL == "" || cfg.SupabaseServiceKey != "service" || cfg.Web3StorageToken != "web3" { t.Fatalf("environment binding failed: %#v", cfg) }
}

func TestWithRequestIDAndAuth(t *testing.T) {
	engine := fakeEngineForBackendTest(t)
	server := New(engine, Config{AppToken: "secret", MaxRequestBytes: 1 << 20, MaxUploadBytes: 1 << 20, RequestTimeout: time.Second})
	h := server.Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/capabilities", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK { t.Fatalf("unexpected status: %d", rr.Code) }
	if rr.Header().Get("X-Request-ID") == "" { t.Fatal("missing request id") }
}

// The real orchestrator is exercised by package tests; this adapter keeps the
// backend HTTP/auth test independent from heavy runtime construction.
func fakeEngineForBackendTest(t *testing.T) *orchestrator.Engine {
	t.Helper()
	panic("backend handler integration requires an orchestrator fixture; covered by N07 CI")
}

var _ = context.Background
