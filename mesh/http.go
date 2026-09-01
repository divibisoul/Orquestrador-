package mesh

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type HTTPGateway struct {
	Engine *orchestrator.Engine
	Token  string
}

func NewHTTPGateway(engine *orchestrator.Engine) *HTTPGateway {
	token := strings.TrimSpace(os.Getenv("N07_MESH_SHARED_TOKEN"))
	return &HTTPGateway{Engine: engine, Token: token}
}

func (g *HTTPGateway) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMeshJSON(w, http.StatusMethodNotAllowed, protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "error", CorrelationID: "", Source: "N07", Target: "", Metadata: map[string]string{"error": "POST required"}})
		return
	}
	if g.Engine == nil {
		writeMeshJSON(w, http.StatusServiceUnavailable, protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "error", CorrelationID: "", Source: "N07", Target: "", Metadata: map[string]string{"error": "N07 engine unavailable"}})
		return
	}
	if g.Token != "" && r.Header.Get("X-Soul-Mesh-Token") != g.Token {
		writeMeshJSON(w, http.StatusUnauthorized, protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "error", CorrelationID: "", Source: "N07", Target: "", Metadata: map[string]string{"error": "mesh authentication failed"}})
		return
	}
	var envelope protocol.MeshEnvelope
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&envelope); err != nil {
		writeMeshJSON(w, http.StatusBadRequest, protocol.MeshEnvelope{ContractVersion: "1.2", Operation: "error", Source: "N07", Target: "", Metadata: map[string]string{"error": err.Error()}})
		return
	}
	if envelope.ContractVersion != "1.2" {
		writeMeshJSON(w, http.StatusBadRequest, gatewayError(envelope, "unsupported contractVersion"))
		return
	}
	if err := envelope.Validate(); err != nil {
		writeMeshJSON(w, http.StatusBadRequest, gatewayError(envelope, err.Error()))
		return
	}
	if envelope.Operation == "mesh.ping" {
		writeMeshJSON(w, http.StatusOK, protocol.MeshEnvelope{ContractVersion: "1.2", Operation: envelope.Operation, Payload: map[string]any{"ok": true, "nucleus": "N07"}, CorrelationID: envelope.CorrelationID, Source: "N07", Target: envelope.Source})
		return
	}
	if envelope.Operation == "mesh.describe" {
		ops := g.Engine.Operations()
		writeMeshJSON(w, http.StatusOK, protocol.MeshEnvelope{ContractVersion: "1.2", Operation: envelope.Operation, Payload: map[string]any{"nucleus": "N07", "operations": ops, "transports": []string{"LOOPBACK_HTTP", "HTTP"}}, CorrelationID: envelope.CorrelationID, Source: "N07", Target: envelope.Source})
		return
	}
	values, err := payloadValues(envelope.Payload)
	if err != nil {
		writeMeshJSON(w, http.StatusBadRequest, gatewayError(envelope, err.Error()))
		return
	}
	result, err := g.Engine.Execute(r.Context(), envelope.Operation, values, envelope.Metadata)
	status := http.StatusOK
	if err != nil {
		status = http.StatusBadRequest
	}
	metadata := map[string]string{"status": result.Status}
	for k, v := range result.Metadata {
		metadata[k] = v
	}
	resp := protocol.MeshEnvelope{ContractVersion: "1.2", Operation: envelope.Operation, Payload: map[string]any{"values": result.Payload}, CorrelationID: envelope.CorrelationID, Source: result.Source, Target: envelope.Source, Metadata: metadata}
	if result.Error != "" {
		resp.Metadata["error"] = result.Error
	}
	writeMeshJSON(w, status, resp)
}

func payloadValues(payload map[string]any) ([]float64, error) {
	if payload == nil {
		return nil, errors.New("payload.values is required")
	}
	v, ok := payload["values"]
	if !ok {
		return nil, errors.New("payload.values is required")
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var values []float64
	if err := json.Unmarshal(encoded, &values); err != nil {
		return nil, errors.New("payload.values must be an array of numbers")
	}
	return values, nil
}

func gatewayError(in protocol.MeshEnvelope, message string) protocol.MeshEnvelope {
	return protocol.MeshEnvelope{ContractVersion: "1.2", Operation: in.Operation, CorrelationID: in.CorrelationID, Source: "N07", Target: in.Source, Metadata: map[string]string{"status": "error", "error": message}}
}

func writeMeshJSON(w http.ResponseWriter, status int, envelope protocol.MeshEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope)
}

func sorted(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
