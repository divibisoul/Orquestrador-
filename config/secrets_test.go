package config

import "testing"

func TestInspectSecretsDegradedWhenRequiredSecretsMissing(t *testing.T) {
	t.Setenv("N07_APP_TOKEN", "")
	t.Setenv("WEB3_STORAGE_TOKEN", "")
	t.Setenv("SUPABASE_URL", "")
	t.Setenv("SUPABASE_SERVICE_ROLE_KEY", "")

	status := InspectSecrets()
	if status.Status != "degraded" || status.Mode != "restricted" {
		t.Fatalf("expected degraded restricted mode, got status=%q mode=%q", status.Status, status.Mode)
	}
	if len(status.Missing) != 4 {
		t.Fatalf("expected four missing secrets, got %d", len(status.Missing))
	}
}

func TestInspectSecretsReadyWhenRequiredSecretsPresent(t *testing.T) {
	t.Setenv("N07_APP_TOKEN", "test-app-token")
	t.Setenv("WEB3_STORAGE_TOKEN", "test-web3-token")
	t.Setenv("SUPABASE_URL", "https://example.supabase.co")
	t.Setenv("SUPABASE_SERVICE_ROLE_KEY", "test-service-role-key")

	status := InspectSecrets()
	if status.Status != "ready" || status.Mode != "full" {
		t.Fatalf("expected ready full mode, got status=%q mode=%q", status.Status, status.Mode)
	}
	if len(status.Missing) != 0 {
		t.Fatalf("expected no missing secrets, got %v", status.Missing)
	}
}
