package api

import (
	"net/http/httptest"
	"testing"
)

func TestNormalizePrompt(t *testing.T) {
	got, err := normalizePrompt([]chatMessage{
		{Role: "system", Content: "system rules"},
		{Role: "user", Content: "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "[system]\nsystem rules\n\n[user]\nhello"
	if got != want {
		t.Fatalf("normalizePrompt = %q, want %q", got, want)
	}
}

func TestContentTextSupportsTextParts(t *testing.T) {
	got, err := contentText([]any{
		map[string]any{"text": "one"},
		map[string]any{"text": "two"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "one\ntwo" {
		t.Fatalf("contentText = %q", got)
	}
}

func TestMeshExecutableRecognizesVersionedCapability(t *testing.T) {
	description := map[string]any{
		"executableCapabilities": []any{"ai.generate@1.0.0", "mesh.describe@1.0.0"},
	}
	if !meshExecutable(description, "ai.generate") {
		t.Fatal("expected executable capability")
	}
	if meshExecutable(description, "ai.generate.ollama") {
		t.Fatal("unexpected Ollama capability")
	}
}

func TestRequestCorrelationIDUsesExternalRequestIDWhenValid(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("X-Request-ID", "langgraph-task-7")
	if got := requestCorrelationID(req); got != "langgraph-task-7" {
		t.Fatalf("requestCorrelationID = %q", got)
	}
}

func TestRequestCorrelationIDRejectsOversizedExternalRequestID(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("X-Request-ID", "0123456789"+string(make([]byte, 201)))
	if got := requestCorrelationID(req); got == req.Header.Get("X-Request-ID") {
		t.Fatal("oversized request id must not become correlation id")
	}
}
