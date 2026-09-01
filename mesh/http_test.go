package mesh

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTestGateway(t *testing.T) *HTTPGateway {
	t.Helper()
	t.Setenv("N07_MESH_HMAC_SECRET", "")
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "true")
	n, err := neural.New(8, 0.05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(0.10, 32)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil { t.Fatal(err) }
	return NewHTTPGateway(e)
}

func newEnvelope(typ, capability, correlation string, values []float64) protocol.MeshEnvelope {
	payload := map[string]any{"capability": capability}
	if values != nil { payload["payload"] = map[string]any{"values": values} }
	return protocol.MeshEnvelope{
		Version: protocol.SoulMeshVersion,
		ContractVersion: protocol.SoulMeshContractVersion,
		MessageID: protocol.NewTraceID(),
		Source: "N01",
		Target: "N07",
		Timestamp: time.Now().UnixMilli(),
		Nonce: protocol.NewTraceID(),
		CorrelationID: correlation,
		Type: typ,
		Payload: payload,
	}
}

func postEnvelope(t *testing.T, h http.Handler, envelope protocol.MeshEnvelope) (protocol.MeshEnvelope, int) {
	t.Helper()
	body, err := json.Marshal(envelope)
	if err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out protocol.MeshEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil { t.Fatal(err) }
	return out, rec.Code
}

func TestHTTPGatewayPingAndDescribe(t *testing.T) {
	h := newTestGateway(t)
	ping := newEnvelope("PING", "mesh.ping", "trace-ping", nil)
	got, code := postEnvelope(t, h.Handler, ping)
	if code != http.StatusOK || got.CorrelationID != ping.CorrelationID || got.Source != "N07" || got.Target != "N01" { t.Fatalf("unexpected ping response: code=%d envelope=%+v", code, got) }
	describe := newEnvelope("CAPABILITY_REQUEST", "mesh.describe", "trace-describe", nil)
	got, code = postEnvelope(t, h.Handler, describe)
	if code != http.StatusOK || got.CorrelationID != describe.CorrelationID { t.Fatalf("unexpected describe response: code=%d envelope=%+v", code, got) }
}

func TestHTTPGatewayExecutesRealNeuralCapability(t *testing.T) {
	h := newTestGateway(t)
	envelope := newEnvelope("CAPABILITY_REQUEST", "neural.forward", "trace-neural", []float64{1,2,3,4,5,6,7,8})
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusOK || got.CorrelationID != envelope.CorrelationID { t.Fatalf("unexpected neural response: code=%d envelope=%+v", code, got) }
	values, ok := got.Payload["values"].([]any)
	if !ok || len(values) != 8 { t.Fatalf("expected eight neural outputs, got %#v", got.Payload["values"]) }
	if got.Payload["status"] != "ok" { t.Fatalf("expected successful execution, got %#v", got.Payload["status"]) }
}

func TestHTTPGatewayPreservesTraceIdentity(t *testing.T) {
	h := newTestGateway(t)
	envelope := newEnvelope("CAPABILITY_REQUEST", "neural.forward", "trace-preserved", []float64{1,2,3,4,5,6,7,8})
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusOK || got.CorrelationID != "trace-preserved" { t.Fatalf("correlation identity was not preserved: code=%d envelope=%+v", code, got) }
}

func TestHTTPGatewayHMACAuthentication(t *testing.T) {
	t.Setenv("N07_MESH_HMAC_SECRET", testSecret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	h := newTestGateway(t)
	h.Secret = testSecret
	envelope := newEnvelope("PING", "mesh.ping", "trace-hmac", nil)
	if err := protocol.SignHMAC(&envelope, testSecret); err != nil { t.Fatal(err) }
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusOK || got.CorrelationID != envelope.CorrelationID || got.HMAC == "" { t.Fatalf("expected signed Mesh response: code=%d envelope=%+v", code, got) }
	if err := protocol.VerifyHMAC(got, testSecret, time.Now()); err != nil { t.Fatalf("response HMAC verification failed: %v", err) }
}

func TestHTTPGatewayRejectsWrongContract(t *testing.T) {
	h := newTestGateway(t)
	envelope := newEnvelope("PING", "mesh.ping", "trace-bad", nil)
	envelope.ContractVersion = "9.9.9"
	got, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusBadRequest || got.Payload["error"] == nil { t.Fatalf("expected contract rejection, code=%d envelope=%+v", code, got) }
}

func TestHTTPGatewayRejectsUnauthenticatedWhenLocalModeDisabled(t *testing.T) {
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	t.Setenv("N07_MESH_HMAC_SECRET", "")
	h := newTestGateway(t)
	h.AllowUnauthenticatedLocal = false
	envelope := newEnvelope("PING", "mesh.ping", "trace-auth", nil)
	_, code := postEnvelope(t, h.Handler, envelope)
	if code != http.StatusUnauthorized { t.Fatalf("expected unauthorized response, got %d", code) }
}

func TestHTTPGatewayHonorsContextCancellation(t *testing.T) {
	h := newTestGateway(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	envelope := newEnvelope("CAPABILITY_REQUEST", "neural.forward", "trace-cancel", []float64{1,2,3,4,5,6,7,8})
	body, err := json.Marshal(envelope)
	if err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body)).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.Handler(rec, req)
	if rec.Code == http.StatusOK { t.Fatal("cancelled request must not be reported as a successful request") }
}
