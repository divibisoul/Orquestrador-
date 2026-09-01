package mesh

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newTestGateway(t *testing.T) *HTTPGateway {
	t.Helper()
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "true")
	t.Setenv("N07_MESH_SHARED_TOKEN", "")
	n, err := neural.New(8, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	return NewHTTPGateway(e)
}

func postEnvelope(t *testing.T, h http.Handler, envelope protocol.MeshEnvelope) (protocol.MeshEnvelope, int) {
	t.Helper()
	body, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out protocol.MeshEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out, rec.Code
}

func TestHTTPGatewayPingAndDescribe(t *testing.T) {
	h := newTestGateway(t)
	ping := protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "mesh.ping", CorrelationID: "trace-ping", Source: "N01", Target: "N07"}
	got, code := postEnvelope(t, h.Handler, ping)
	if code != http.StatusOK || got.CorrelationID != ping.CorrelationID || got.Source != "N07" || got.Target != "N01" {
		t.Fatalf("unexpected ping response: code=%d envelope=%+v", code, got)
	}
	describe := protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "mesh.describe", CorrelationID: "trace-describe", Source: "N01", Target: "N07"}
	got, code = postEnvelope(t, h.Handler, describe)
	if code != http.StatusOK || got.CorrelationID != describe.CorrelationID {
		t.Fatalf("unexpected describe response: code=%d envelope=%+v", code, got)
	}
}

func TestHTTPGatewayExecutesRealNeuralCapability(t *testing.T) {
	h := newTestGateway(t)
	envelope := protocol.MeshEnvelope{
		ContractVersion: "1.2",
		Operation:       "neural.forward",
		CorrelationID:   "trace-neural",
		Source:          "N01",
		Target:          "N07",
		Payload:         map[string]any{"values": []float64{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusOK || got.CorrelationID != envelope.CorrelationID {
		t.Fatalf("unexpected neural response: code=%d envelope=%+v", code, got)
	}
	values, ok := got.Payload["values"].([]any)
	if !ok || len(values) != 8 {
		t.Fatalf("expected eight neural outputs, got %#v", got.Payload["values"])
	}
	if got.Metadata["status"] != "ok" {
		t.Fatalf("expected successful execution metadata, got %#v", got.Metadata)
	}
}

func TestHTTPGatewayPreservesTraceIdentity(t *testing.T) {
	h := newTestGateway(t)
	envelope := protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "neural.forward", CorrelationID: "trace-preserved", Source: "N01", Target: "N07", Payload: map[string]any{"values": []float64{1, 2, 3, 4, 5, 6, 7, 8}}}
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusOK || got.CorrelationID != "trace-preserved" {
		t.Fatalf("correlation identity was not preserved: code=%d envelope=%+v", code, got)
	}
}

func TestHTTPGatewayRejectsWrongContract(t *testing.T) {
	h := newTestGateway(t)
	envelope := protocol.MeshEnvelope{ContractVersion: "9.9", Operation: "mesh.ping", CorrelationID: "trace-bad", Source: "N01", Target: "N07"}
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusBadRequest || got.Metadata["status"] != "error" {
		t.Fatalf("expected contract rejection, code=%d envelope=%+v", code, got)
	}
}

func TestHTTPGatewayRejectsUnauthenticatedWhenLocalModeDisabled(t *testing.T) {
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	t.Setenv("N07_MESH_SHARED_TOKEN", "")
	h := newTestGateway(t)
	h.AllowUnauthenticatedLocal = false
	envelope := protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "mesh.ping", CorrelationID: "trace-auth", Source: "N01", Target: "N07"}
	_, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got %d", code)
	}
}

func TestHTTPGatewayHonorsContextCancellation(t *testing.T) {
	h := newTestGateway(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewBufferString(`{"contractVersion":"1.2","operation":"neural.forward","correlationId":"trace-cancel","source":"N01","target":"N07","payload":{"values":[1,2,3,4,5,6,7,8]}}`)).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.Handler(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("cancelled request must not be reported as a successful request")
	}
}
