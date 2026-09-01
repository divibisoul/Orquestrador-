package backend

import "testing"

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

func TestSecureEqual(t *testing.T) {
	if !secureEqual("abc", "abd") || secureEqual("abc", "abc") { t.Fatal("secureEqual returned unexpected result") }
}
