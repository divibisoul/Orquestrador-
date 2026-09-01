package mesh

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type HTTPGateway struct { Engine *orchestrator.Engine; Secret string; AllowUnauthenticatedLocal bool }

type canonicalWireEnvelope struct {
	Protocol string `json:"protocol"`
	ContractVersion string `json:"contractVersion"`
	ID string `json:"id"`
	MessageID string `json:"messageId"`
	CorrelationID string `json:"correlationId"`
	Source string `json:"source"`
	Target string `json:"target"`
	Kind string `json:"kind"`
	Type string `json:"type"`
	Capability string `json:"capability"`
	Payload map[string]any `json:"payload"`
	Timestamp int64 `json:"timestamp"`
	Nonce string `json:"nonce"`
	HMAC string `json:"hmac"`
	Meta map[string]any `json:"meta,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Operation string `json:"operation,omitempty"`
}

func NewHTTPGateway(engine *orchestrator.Engine) *HTTPGateway { return &HTTPGateway{Engine: engine, Secret: strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")), AllowUnauthenticatedLocal: strings.EqualFold(strings.TrimSpace(os.Getenv("N07_MESH_ALLOW_UNAUTH_LOCAL")), "true")} }
func (g *HTTPGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) { g.Handler(w, r) }
func (g *HTTPGateway) authenticated(envelope protocol.MeshEnvelope) error { if g.Secret != "" { return protocol.VerifyHMAC(envelope, g.Secret, time.Now()) }; if g.AllowUnauthenticatedLocal { return nil }; return errors.New("Mesh HMAC secret is not configured") }

func canonicalCapability(w canonicalWireEnvelope) string { if strings.TrimSpace(w.Capability)!="" { return strings.TrimSpace(w.Capability) }; if w.Payload!=nil { if v,ok:=w.Payload["capability"].(string); ok { return strings.TrimSpace(v) } }; return "" }
func canonicalKind(w canonicalWireEnvelope) string { if strings.TrimSpace(w.Kind)!="" { return strings.TrimSpace(w.Kind) }; switch strings.TrimSpace(w.Type) { case "PING": return "request"; case "TASK_RESULT": return "response"; case "ERROR": return "error"; default: return "request" } }
func canonicalID(w canonicalWireEnvelope) string { if strings.TrimSpace(w.ID)!="" { return strings.TrimSpace(w.ID) }; return strings.TrimSpace(w.MessageID) }
func normalizedMeshEnvelope(w canonicalWireEnvelope) protocol.MeshEnvelope {
	payload := w.Payload
	if payload == nil { payload = map[string]any{} }
	if w.Capability != "" { if _,exists:=payload["capability"]; !exists { payload=cloneMap(payload); payload["capability"]=w.Capability } }
	return protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:w.ContractVersion,MessageID:canonicalID(w),Source:w.Source,Target:w.Target,Timestamp:w.Timestamp,Nonce:w.Nonce,CorrelationID:w.CorrelationID,Type:canonicalType(w),HMAC:w.HMAC,Payload:payload,Operation:w.Operation,Metadata:w.Metadata}
}
func canonicalType(w canonicalWireEnvelope) string { switch canonicalKind(w) { case "request": if strings.EqualFold(w.Capability,"mesh.ping")||strings.EqualFold(w.Type,"PING") { return "PING" }; return "CAPABILITY_REQUEST"; case "response": return "TASK_RESULT"; case "error": return "ERROR"; default: return "TASK_RESULT" } }
func cloneMap(in map[string]any) map[string]any { out:=make(map[string]any,len(in)+1); for k,v:=range in {out[k]=v}; return out }

func (g *HTTPGateway) respond(w http.ResponseWriter, status int, in protocol.MeshEnvelope, typ string, payload map[string]any) {
	kind := "response"; if typ=="ERROR" { kind="error" }
	capability := in.Capability();
	if capability=="" { if v,ok:=in.Payload["capability"].(string); ok { capability=strings.TrimSpace(v) } }
	response := map[string]any{
		"protocol":protocol.SoulMeshVersionProtocol(),
		"contractVersion":protocol.SoulMeshContractVersion,
		"id":protocol.NewTraceID(),
		"correlationId":in.CorrelationID,
		"source":"N07",
		"target":in.Source,
		"kind":kind,
		"capability":capability,
		"payload":payload,
		"timestamp":time.Now().UnixMilli(),
	}
	if g.Secret!="" {
		legacy:=protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,MessageID:response["id"].(string),Source:"N07",Target:in.Source,Timestamp:response["timestamp"].(int64),Nonce:protocol.NewTraceID(),CorrelationID:in.CorrelationID,Type:typ,Payload:map[string]any{"capability":capability,"payload":payload}}
		if err:=protocol.SignHMAC(&legacy,g.Secret); err!=nil { writeMeshJSON(w,http.StatusInternalServerError,map[string]any{"error":"response signing failed"}); return }
		response["nonce"]=legacy.Nonce; response["hmac"]=legacy.HMAC
	}
	writeMeshJSON(w,status,response)
}

func (g *HTTPGateway) Handler(w http.ResponseWriter,r *http.Request) {
	if r.Method!=http.MethodPost { writeMeshJSON(w,http.StatusMethodNotAllowed,map[string]any{"error":"POST required"}); return }
	if g.Engine==nil { writeMeshJSON(w,http.StatusServiceUnavailable,map[string]any{"error":"N07 engine unavailable"}); return }
	var wire canonicalWireEnvelope
	if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&wire); err!=nil { writeMeshJSON(w,http.StatusBadRequest,map[string]any{"error":err.Error()}); return }
	envelope:=normalizedMeshEnvelope(wire)
	if envelope.ContractVersion=="" { envelope.ContractVersion=protocol.SoulMeshContractVersion }
	if envelope.Version=="" { envelope.Version=protocol.SoulMeshVersion }
	if envelope.Nonce=="" { envelope.Nonce=protocol.NewTraceID() }
	if envelope.Timestamp==0 { envelope.Timestamp=time.Now().UnixMilli() }
	if err:=envelope.Validate(); err!=nil { writeMeshJSON(w,http.StatusBadRequest,gatewayError(envelope,err.Error())); return }
	if err:=g.authenticated(envelope); err!=nil { g.respond(w,http.StatusUnauthorized,envelope,"ERROR",map[string]any{"error":"mesh authentication failed"}); return }
	if envelope.Target!="N07"&&envelope.Target!="BROADCAST" { g.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"target is not N07"}); return }
	capability:=canonicalCapability(wire); if capability=="" { capability=envelope.Capability() }
	if canonicalKind(wire)=="request" && (wire.Type=="PING"||capability=="mesh.ping") { g.respond(w,http.StatusOK,envelope,"TASK_RESULT",map[string]any{"ok":true,"nucleus":"N07"}); return }
	if canonicalKind(wire)=="request" && capability=="mesh.describe" { g.respond(w,http.StatusOK,envelope,"TASK_RESULT",map[string]any{"nucleus":"N07","operations":g.Engine.Operations(),"transports":[]string{"LOOPBACK_HTTP","HTTP"}}); return }
	if capability=="" { g.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"capability is required"}); return }
	values,err:=payloadValues(envelope.NestedPayload()); if err!=nil { g.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":err.Error()}); return }
	result,err:=g.Engine.ExecuteWithTrace(r.Context(),envelope.CorrelationID,envelope.Source,capability,values,envelope.NestedMetadata()); status:=http.StatusOK; if err!=nil { status=http.StatusBadRequest }
	payload:=map[string]any{"values":result.Payload,"status":result.Status}; if result.Error!="" { payload["error"]=result.Error }
	g.respond(w,status,envelope,"TASK_RESULT",payload)
}
func payloadValues(payload map[string]any)([]float64,error){if payload==nil{return nil,errors.New("payload.values is required")};v,ok:=payload["values"];if !ok{return nil,errors.New("payload.values is required")};encoded,err:=json.Marshal(v);if err!=nil{return nil,err};var values []float64;if err:=json.Unmarshal(encoded,&values);err!=nil{return nil,errors.New("payload.values must be an array of numbers")};return values,nil}
func gatewayError(in protocol.MeshEnvelope,message string)map[string]any{return map[string]any{"protocol":protocol.SoulMeshVersionProtocol(),"contractVersion":protocol.SoulMeshContractVersion,"id":protocol.NewTraceID(),"correlationId":in.CorrelationID,"source":"N07","target":in.Source,"kind":"error","capability":in.Capability(),"payload":map[string]any{"error":message},"timestamp":time.Now().UnixMilli()}}
func writeMeshJSON(w http.ResponseWriter,status int,body any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(body)}
func sorted(values []string)[]string{out:=append([]string(nil),values...);sort.Strings(out);return out}
