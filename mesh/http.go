package mesh

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

type HTTPGateway struct {
	Engine                      *orchestrator.Engine
	Federation                  *orchestrator.Federation
	Secret                      string
	AllowUnauthenticatedLocal   bool
}

type canonicalWireEnvelope struct {
	Protocol        string                 `json:"protocol"`
	ContractVersion string                 `json:"contractVersion"`
	ID              string                 `json:"id"`
	MessageID       string                 `json:"messageId"`
	CorrelationID   string                 `json:"correlationId"`
	Source          string                 `json:"source"`
	Target          string                 `json:"target"`
	Kind            string                 `json:"kind"`
	Type            string                 `json:"type"`
	Capability      string                 `json:"capability"`
	Payload         map[string]any         `json:"payload"`
	Timestamp       int64                  `json:"timestamp"`
	Nonce           string                 `json:"nonce"`
	HMAC            string                 `json:"hmac"`
	Meta            map[string]any         `json:"meta,omitempty"`
	Metadata        map[string]string      `json:"metadata,omitempty"`
	Operation       string                 `json:"operation,omitempty"`
}

func NewHTTPGateway(engine *orchestrator.Engine) *HTTPGateway {
	return &HTTPGateway{Engine: engine, Secret: strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")), AllowUnauthenticatedLocal: strings.EqualFold(strings.TrimSpace(os.Getenv("N07_MESH_ALLOW_UNAUTH_LOCAL")), "true")}
}

func (g *HTTPGateway) SetFederation(federation *orchestrator.Federation) { g.Federation = federation }
func (g *HTTPGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) { g.Handler(w, r) }
func (g *HTTPGateway) authenticated(envelope protocol.MeshEnvelope) error { if g.Secret != "" { return protocol.VerifyHMAC(envelope, g.Secret, time.Now()) }; if g.AllowUnauthenticatedLocal { return nil }; return errors.New("Mesh HMAC secret is not configured") }
func canonicalCapability(w canonicalWireEnvelope) string { if strings.TrimSpace(w.Capability) != "" { return strings.TrimSpace(w.Capability) }; if w.Payload != nil { if v, ok := w.Payload["capability"].(string); ok { return strings.TrimSpace(v) } }; return "" }
func canonicalKind(w canonicalWireEnvelope) string { if strings.TrimSpace(w.Kind) != "" { return strings.TrimSpace(w.Kind) }; switch strings.TrimSpace(w.Type) { case "PING": return "request"; case "TASK_RESULT": return "response"; case "ERROR": return "error"; default: return "" } }
func canonicalID(w canonicalWireEnvelope) string { if strings.TrimSpace(w.ID) != "" { return strings.TrimSpace(w.ID) }; return strings.TrimSpace(w.MessageID) }
func normalizedMeshEnvelope(w canonicalWireEnvelope) protocol.MeshEnvelope { payload := w.Payload; if payload == nil { payload = map[string]any{} }; if w.Capability != "" { if _, exists := payload["capability"]; !exists { payload = cloneMap(payload); payload["capability"] = w.Capability } }; return protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: w.ContractVersion, MessageID: canonicalID(w), Source: w.Source, Target: w.Target, Timestamp: w.Timestamp, Nonce: w.Nonce, CorrelationID: w.CorrelationID, Type: canonicalType(w), HMAC: w.HMAC, Payload: payload, Operation: w.Operation, Metadata: w.Metadata} }
func canonicalType(w canonicalWireEnvelope) string { switch canonicalKind(w) { case "request": if strings.EqualFold(canonicalCapability(w), "mesh.ping") || strings.EqualFold(w.Type, "PING") { return "PING" }; return "CAPABILITY_REQUEST"; case "response": return "TASK_RESULT"; case "error": return "ERROR"; default: return "" } }
func cloneMap(in map[string]any) map[string]any { out := make(map[string]any, len(in)+1); for k, v := range in { out[k] = v }; return out }
func canonicalN01Bytes(w canonicalWireEnvelope, nonce string) ([]byte, error) { return json.Marshal(struct { Protocol string `json:"protocol"`; ContractVersion string `json:"contractVersion"`; ID string `json:"id"`; CorrelationID string `json:"correlationId"`; Source string `json:"source"`; Target string `json:"target"`; Kind string `json:"kind"`; Capability string `json:"capability"`; Payload map[string]any `json:"payload"`; Timestamp int64 `json:"timestamp"`; Transport any `json:"transport"`; Meta map[string]any `json:"meta"`; Nonce string `json:"nonce"` }{w.Protocol, w.ContractVersion, canonicalID(w), w.CorrelationID, w.Source, w.Target, canonicalKind(w), canonicalCapability(w), w.Payload, func() any { if w.Meta != nil { if v, ok := w.Meta["transport"]; ok { return v } }; return nil }(), w.Meta, nonce}) }
func verifyN01HeaderHMAC(w canonicalWireEnvelope, r *http.Request, secret string) error { nonce := strings.TrimSpace(r.Header.Get("x-soul-mesh-nonce")); signature := strings.TrimSpace(r.Header.Get("x-soul-mesh-hmac")); if nonce == "" || signature == "" { return errors.New("Mesh header HMAC credentials are required") }; if w.Nonce != "" && w.Nonce != nonce { return errors.New("Mesh nonce mismatch") }; unsigned, err := canonicalN01Bytes(w, nonce); if err != nil { return err }; mac := hmac.New(sha256.New, []byte(secret)); _, _ = mac.Write(unsigned); expected := mac.Sum(nil); actual, err := hex.DecodeString(signature); if err != nil || !hmac.Equal(expected, actual) { return errors.New("Invalid Mesh header HMAC") }; return nil }
func (g *HTTPGateway) authenticateWire(w canonicalWireEnvelope, r *http.Request, envelope protocol.MeshEnvelope) error { if g.Secret == "" { return g.authenticated(envelope) }; if r.Header.Get("x-soul-mesh-hmac") != "" || r.Header.Get("x-soul-mesh-nonce") != "" { return verifyN01HeaderHMAC(w, r, g.Secret) }; return g.authenticated(envelope) }
func (g *HTTPGateway) respond(w http.ResponseWriter, status int, in protocol.MeshEnvelope, typ string, payload map[string]any) { kind := "response"; if typ == "ERROR" { kind = "error" }; capability := in.Capability(); if capability == "" { if v, ok := in.Payload["capability"].(string); ok { capability = strings.TrimSpace(v) } }; messageID := protocol.NewTraceID(); timestamp := time.Now().UnixMilli(); response := map[string]any{"protocol": "soul-mesh/1", "contractVersion": protocol.SoulMeshContractVersion, "id": messageID, "correlationId": in.CorrelationID, "source": "N07", "target": in.Source, "kind": kind, "capability": capability, "payload": payload, "timestamp": timestamp}; if g.Secret != "" { legacy := protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: protocol.SoulMeshContractVersion, MessageID: messageID, Source: "N07", Target: in.Source, Timestamp: timestamp, Nonce: protocol.NewTraceID(), CorrelationID: in.CorrelationID, Type: typ, Payload: map[string]any{"capability": capability, "payload": payload}}; if err := protocol.SignHMAC(&legacy, g.Secret); err != nil { writeMeshJSON(w, http.StatusInternalServerError, map[string]any{"error": "response signing failed"}); return }; response["nonce"] = legacy.Nonce; response["hmac"] = legacy.HMAC }; writeMeshJSON(w, status, response) }

func (g *HTTPGateway) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeMeshJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"}); return }
	if g.Engine == nil { writeMeshJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "N07 engine unavailable"}); return }
	var wire canonicalWireEnvelope
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&wire); err != nil { writeMeshJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()}); return }
	if wire.Protocol == "" || wire.ContractVersion == "" || canonicalID(wire) == "" || wire.CorrelationID == "" || wire.Source == "" || wire.Target == "" || canonicalKind(wire) == "" || wire.Timestamp <= 0 { writeMeshJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid canonical Mesh identity"}); return }
	envelope := normalizedMeshEnvelope(wire)
	if err := envelope.Validate(); err != nil { writeMeshJSON(w, http.StatusBadRequest, gatewayError(envelope, err.Error())); return }
	if err := g.authenticateWire(wire, r, envelope); err != nil { g.respond(w, http.StatusUnauthorized, envelope, "ERROR", map[string]any{"error": "mesh authentication failed"}); return }
	if envelope.Target != "N07" && envelope.Target != "BROADCAST" { g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "target is not N07"}); return }
	capability := canonicalCapability(wire)
	if capability == "" { g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "capability is required"}); return }
	if canonicalKind(wire) == "request" && (wire.Type == "PING" || capability == "mesh.ping") { g.respond(w, http.StatusOK, envelope, "TASK_RESULT", map[string]any{"ok": true, "nucleus": "N07"}); return }
	if canonicalKind(wire) == "request" && (capability == "mesh.discovery" || capability == "mesh.describe") {
		payload := map[string]any{"nucleus": "N07", "operations": g.Engine.Operations(), "transports": []string{"IN_PROCESS", "LOOPBACK_HTTP", "HTTP", "REALTIME"}}
		if g.Federation != nil { peers := g.Federation.Snapshot(); snapshot := make([]map[string]any, 0, len(peers)); for _, peer := range peers { snapshot = append(snapshot, map[string]any{"nucleus": peer.Nucleus, "url": peer.URL, "capabilities": peer.Capabilities, "healthy": peer.Healthy, "latency_ms": float64(peer.Latency) / float64(time.Millisecond), "in_flight": peer.InFlight}) }; payload["peers"] = snapshot }
		g.respond(w, http.StatusOK, envelope, "TASK_RESULT", payload); return
	}

	routeMessage := protocol.Message{Version: protocol.Version, TraceID: envelope.CorrelationID, Source: envelope.Source, Target: "N07", Kind: "command", Operation: capability, Payload: nil}
	if _, routeErr := g.Engine.Route(routeMessage); routeErr == nil {
		values, err := payloadValues(envelope.NestedPayload())
		if err != nil { g.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": err.Error()}); return }
		result, err := g.Engine.ExecuteWithTrace(r.Context(), envelope.CorrelationID, envelope.Source, capability, values, envelope.NestedMetadata())
		status := http.StatusOK
		if err != nil { status = http.StatusBadRequest }
		payload := map[string]any{"values": result.Payload, "status": result.Status, "execution": "local"}
		if result.Error != "" { payload["error"] = result.Error }
		g.respond(w, status, envelope, "TASK_RESULT", payload)
		return
	}

	if g.Federation == nil { g.respond(w, http.StatusNotImplemented, envelope, "ERROR", map[string]any{"error": "capability has no local handler and N07 federation is unavailable", "capability": capability}); return }
	remotePayload := envelope.NestedPayload()
	if remotePayload == nil { remotePayload = map[string]any{} }
	result, err := g.Federation.Delegate(r.Context(), envelope.CorrelationID, capability, remotePayload)
	if err != nil { g.respond(w, http.StatusBadGateway, envelope, "ERROR", map[string]any{"error": err.Error(), "capability": capability}); return }
	g.respond(w, http.StatusOK, envelope, "TASK_RESULT", map[string]any{"execution": "federated", "peer_result": result})
}

func payloadValues(payload map[string]any) ([]float64, error) { if payload == nil { return nil, errors.New("payload.values is required") }; v, ok := payload["values"]; if !ok { return nil, errors.New("payload.values is required") }; encoded, err := json.Marshal(v); if err != nil { return nil, err }; var values []float64; if err := json.Unmarshal(encoded, &values); err != nil { return nil, errors.New("payload.values must be an array of numbers") }; return values, nil }
func gatewayError(in protocol.MeshEnvelope, message string) map[string]any { return map[string]any{"protocol": "soul-mesh/1", "contractVersion": protocol.SoulMeshContractVersion, "id": protocol.NewTraceID(), "correlationId": in.CorrelationID, "source": "N07", "target": in.Source, "kind": "error", "capability": in.Capability(), "payload": map[string]any{"error": message}, "timestamp": time.Now().UnixMilli()} }
func writeMeshJSON(w http.ResponseWriter, status int, body any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(body) }
func sorted(values []string) []string { out := append([]string(nil), values...); sort.Strings(out); return out }
