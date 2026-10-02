package agentarsenal

import (
	"context"
	"testing"
)

func TestProxyConfigurationIsFailClosed(t *testing.T) {
	t.Setenv("SOUL_AGENT_ARSENAL_URL", "")
	p := NewFromEnv()
	if p.Configured() {
		t.Fatal("proxy must remain unconfigured without an endpoint")
	}
	if _, err := p.Catalog(context.Background(), "", "", 100, 0); err == nil {
		t.Fatal("catalog must fail closed without an endpoint")
	}
}

func TestProxyRejectsNilContext(t *testing.T) {
	t.Setenv("SOUL_AGENT_ARSENAL_URL", "http://127.0.0.1:1")
	p := NewFromEnv()
	if _, err := p.Catalog(nil, "", "", 1, 0); err == nil || err.Error() != "context is nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}
