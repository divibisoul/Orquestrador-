package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestExecutePropagatesCorrelationIDToEngine(t *testing.T) {
	n, err := neural.New(2, 0.01)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(0.1, 2)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil { t.Fatal(err) }

	const expected = "corr-http-contract-001"
	if err := e.Register("test.correlation@1.0.0", func(_ context.Context, message protocol.Message) (protocol.Result, error) {
		if message.CorrelationID != expected {
			return protocol.Result{}, &correlationMismatchError{got: message.CorrelationID}
		}
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.test", Target: message.Source, Status: "ok"}, nil
	}); err != nil { t.Fatal(err) }

	server := New(e, Config{AppToken: "n07-test-token"})
	body := strings.NewReader("{\"operation\":\"test.correlation@1.0.0\",\"payload\":[0],\"correlationId\":\"" + expected + "\"}")
	req := httptest.NewRequest(http.MethodPost, "/v1/execute", body)
	req.Header.Set("Authorization", "Bearer n07-test-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.execute(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if !strings.Contains(rec.Body.String(), expected) { t.Fatalf("response did not preserve correlation id: %s", rec.Body.String()) }
}

type correlationMismatchError struct{ got string }
func (e *correlationMismatchError) Error() string { return "correlation mismatch: " + e.got }
