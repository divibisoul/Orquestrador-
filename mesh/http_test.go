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

type testWire struct {
	Protocol string `json:"protocol"`
	ContractVersion string `json:"contractVersion"`
	ID string `json:"id"`
	CorrelationID string `json:"correlationId"`
	Source string `json:"source"`
	Target string `json:"target"`
	Kind string `json:"kind"`
	Capability string `json:"capability"`
	Payload map[string]any `json:"payload"`
	Timestamp int64 `json:"timestamp"`
	Nonce string `json:"nonce"`
	HMAC string `json:"hmac"`
}

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

func canonicalRequest(kind, capability, correlation string, values []float64) map[string]any {
	payload := map[string]any{}
	if values != nil { payload["values"] = values }
	return map[string]any{"protocol":"soul-mesh/1","contractVersion":"1.1.0","id":protocol.NewTraceID(),"correlationId":correlation,"source":"N01","target":"N07","kind":kind,"capability":capability,"payload":payload,"timestamp":time.Now().UnixMilli(),"nonce":protocol.NewTraceID()}
}

func postWire(t *testing.T, h http.Handler, wire map[string]any) (testWire, int) {
	t.Helper()
	body, err := json.Marshal(wire)
	if err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out testWire
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil { t.Fatal(err) }
	return out, rec.Code
}

func TestHTTPGatewayPingAndDescribe(t *testing.T) { h:=newTestGateway(t); ping:=canonicalRequest("request","mesh.ping","trace-ping",nil); got,code:=postWire(t,h,ping); if code!=http.StatusOK||got.CorrelationID!="trace-ping"||got.Source!="N07"||got.Target!="N01"||got.Kind!="response"||got.Capability!="mesh.ping"||got.ContractVersion!="1.1.0"{t.Fatalf("unexpected ping response: code=%d envelope=%+v",code,got)}; describe:=canonicalRequest("request","mesh.describe","trace-describe",nil); got,code=postWire(t,h,describe); if code!=http.StatusOK||got.CorrelationID!="trace-describe"||got.Kind!="response"{t.Fatalf("unexpected describe response: code=%d envelope=%+v",code,got)} }
func TestHTTPGatewayExecutesRealNeuralCapability(t *testing.T) { h:=newTestGateway(t); wire:=canonicalRequest("request","neural.forward","trace-neural",[]float64{1,2,3,4,5,6,7,8}); got,code:=postWire(t,h,wire); if code!=http.StatusOK||got.CorrelationID!="trace-neural"||got.Kind!="response"{t.Fatalf("unexpected neural response: code=%d envelope=%+v",code,got)}; values,ok:=got.Payload["values"].([]any); if !ok||len(values)!=8{t.Fatalf("expected eight neural outputs, got %#v",got.Payload["values"])}; if got.Payload["status"]!="ok"{t.Fatalf("expected successful execution, got %#v",got.Payload["status"])} }
func TestHTTPGatewayPreservesTraceIdentity(t *testing.T) { h:=newTestGateway(t); wire:=canonicalRequest("request","neural.forward","trace-preserved",[]float64{1,2,3,4,5,6,7,8}); got,code:=postWire(t,h,wire); if code!=http.StatusOK||got.CorrelationID!="trace-preserved"{t.Fatalf("correlation identity was not preserved: code=%d envelope=%+v",code,got)} }
func TestHTTPGatewayHMACAuthentication(t *testing.T) { t.Setenv("N07_MESH_HMAC_SECRET",testSecret); t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL","false"); h:=newTestGateway(t); h.Secret=testSecret; legacy:=protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,MessageID:protocol.NewTraceID(),Source:"N01",Target:"N07",Timestamp:time.Now().UnixMilli(),Nonce:protocol.NewTraceID(),CorrelationID:"trace-hmac",Type:"PING",Payload:map[string]any{"capability":"mesh.ping"}}; if err:=protocol.SignHMAC(&legacy,testSecret);err!=nil{t.Fatal(err)}; wire:=map[string]any{"protocol":"soul-mesh/1","contractVersion":"1.1.0","id":legacy.MessageID,"correlationId":legacy.CorrelationID,"source":"N01","target":"N07","kind":"request","capability":"mesh.ping","payload":map[string]any{},"timestamp":legacy.Timestamp,"nonce":legacy.Nonce,"hmac":legacy.HMAC}; got,code:=postWire(t,h,wire); if code!=http.StatusOK||got.CorrelationID!="trace-hmac"||got.HMAC==""{t.Fatalf("expected signed Mesh response: code=%d envelope=%+v",code,got)} }
func TestHTTPGatewayRejectsWrongContract(t *testing.T) { h:=newTestGateway(t); wire:=canonicalRequest("request","mesh.ping","trace-bad",nil); wire["contractVersion"]="9.9.9"; got,code:=postWire(t,h,wire); if code!=http.StatusBadRequest||got.Payload["error"]==nil{t.Fatalf("expected contract rejection, code=%d envelope=%+v",code,got)} }
func TestHTTPGatewayRejectsUnauthenticatedWhenLocalModeDisabled(t *testing.T) { t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL","false"); t.Setenv("N07_MESH_HMAC_SECRET",""); h:=newTestGateway(t); h.AllowUnauthenticatedLocal=false; wire:=canonicalRequest("request","mesh.ping","trace-auth",nil); _,code:=postWire(t,h,wire); if code!=http.StatusUnauthorized{t.Fatalf("expected unauthorized response, got %d",code)} }
func TestHTTPGatewayHonorsContextCancellation(t *testing.T) { h:=newTestGateway(t); ctx,cancel:=context.WithCancel(context.Background()); cancel(); wire:=canonicalRequest("request","neural.forward","trace-cancel",[]float64{1,2,3,4,5,6,7,8}); body,err:=json.Marshal(wire);if err!=nil{t.Fatal(err)};req:=httptest.NewRequest(http.MethodPost,"/api/soul-mesh",bytes.NewReader(body)).WithContext(ctx);rec:=httptest.NewRecorder();h.ServeHTTP(rec,req);if rec.Code==http.StatusOK{t.Fatal("cancelled request must not be reported as a successful request")} }
