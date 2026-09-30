package superagi

import (
	"context"
	"strings"
	"testing"
)

func TestNewRequiresExternalRuntime(t *testing.T) {
	if _, err := New(Config{}); err == nil || !strings.Contains(err.Error(), "SUPERAGI_BLOCKED_INFRASTRUCTURE") {
		t.Fatalf("expected fail-closed configuration error, got %v", err)
	}

	if _, err := New(Config{BaseURL: "https://example.invalid"}); err == nil || !strings.Contains(err.Error(), "SUPERAGI_BLOCKED_INFRASTRUCTURE") {
		t.Fatalf("expected fail-closed API key error, got %v", err)
	}
}

func TestRequestValidationIsLocalAndDeterministic(t *testing.T) {
	c, err := New(Config{BaseURL: "https://example.invalid", APIKey: "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateAgent(context.Background(), CreateAgentRequest{}); err == nil {
		t.Fatal("expected create-agent validation error")
	}
	if _, err := c.StartRun(context.Background(), RunAgentRequest{}); err == nil {
		t.Fatal("expected start-run validation error")
	}
	if c.CreatePath != "/agents/create" {
		t.Fatalf("unexpected create path: %s", c.CreatePath)
	}
	if c.RunPath != "/agents/run" {
		t.Fatalf("unexpected run path: %s", c.RunPath)
	}
}
