package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/divibisoul/Orquestrador-/jev"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestJevOperationPreservesCorrelationAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/systemone" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var request jev.Request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.Model != "jev-latest" {
			t.Fatalf("unexpected model: %q", request.Model)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-1.13.0","answers":{"needs_human":{"type":"noul","noul":0.8}},"usage":{"input_tokens":12,"output_tokens":4}}`))
	}))
	defer server.Close()

	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.1, 16)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(n, c, supergpu.New(nil))
	if err != nil {
		t.Fatal(err)
	}

	client := &jev.Client{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		Model:      "jev-latest",
		HTTPClient: server.Client(),
	}
	if err := RegisterJevOperations(e, client); err != nil {
		t.Fatal(err)
	}

	message := protocol.NewMessage("N06", "N07", "command", "jev.systemone@1.0.0", nil)
	message.Metadata["state"] = "A destructive action is proposed."
	message.Metadata["questions_json"] = `{"needs_human":{"type":"noul","instructions":"Does this need human review?"}}`

	result, err := e.Submit(context.Background(), message)
	if err != nil {
		t.Fatal(err)
	}
	if result.TraceID != message.TraceID {
		t.Fatalf("trace id not preserved: got %q want %q", result.TraceID, message.TraceID)
	}
	if result.CorrelationID != message.CorrelationID {
		t.Fatalf("correlation id not preserved: got %q want %q", result.CorrelationID, message.CorrelationID)
	}
	if result.Metadata["answers_json"] == "" || result.Metadata["usage_json"] == "" {
		t.Fatalf("structured Jev metadata missing: %#v", result.Metadata)
	}
}
