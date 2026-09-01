package mesh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type FederatedGateway struct {
	base      *HTTPGateway
	peers     *PeerClient
	engine    *orchestrator.Engine
	catalogMu sync.RWMutex
	catalog   []FusionCapability
}

func NewFederatedHTTPGateway(engine *orchestrator.Engine) *FederatedGateway {
	peers, err := NewPeerClient(nil)
	if err != nil {
		peers = &PeerClient{}
	}
	return &FederatedGateway{base: NewHTTPGateway(engine), peers: peers, engine: engine, catalog: VerifiedN01N06Catalog()}
}

func (g *FederatedGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		g.base.Handler(w, r)
		return
	}
	if g.engine == nil {
		g.base.Handler(w, r)
		return
	}
	body, err := ioReadLimited(r.Body, 1<<20)
	if err != nil {
		writeMeshJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	r.Body = ioNopCloser(bytes.NewReader(body))
	var wire canonicalWireEnvelope
	if err := json.Unmarshal(body, &wire); err != nil {
		g.base.Handler(w, r)
		return
	}
	capability := canonicalCapability(wire)
	if capability == "" {
		g.base.Handler(w, r)
		return
	}
	if isLocalOperation(g.engine, capability) {
		r.Body = ioNopCloser(bytes.NewReader(body))
		g.base.Handler(w, r)
		return
	}
	envelope := normalizedMeshEnvelope(wire)
	if err := envelope.Validate(); err != nil {
		g.base.Handler(w, r)
		return
	}
	if err := g.base.authenticateWire(wire, r, envelope); err != nil {
		g.base.Handler(w, r)
		return
	}
	if capability == "mesh.ping" || capability == "mesh.describe" || capability == "mesh.discovery" || capability == "mesh.capabilities" || capability == "mesh.capability.resolve" || capability == "core.health" || capability == "mesh.fusion.describe" || capability == "mesh.supergpu.describe" {
		g.handleControl(w, envelope, capability)
		return
	}
	if capability == "mesh.fusion.execute" || capability == "fusion.execute" {
		g.handleComposition(w, r, envelope)
		return
	}
	peer, descErr := g.selectPeer(r.Context(), capability)
	if descErr != nil {
		g.base.respond(w, http.StatusServiceUnavailable, envelope, "ERROR", map[string]any{"error": "no federated peer capability available", "capability": capability, "details": descErr.Error()})
		return
	}
	payload := envelope.NestedPayload()
	if payload == nil {
		payload = map[string]any{}
	}
	result, err := g.peers.CallWithCorrelation(r.Context(), peer, capability, payload, envelope.CorrelationID)
	if err != nil {
		g.base.respond(w, http.StatusBadGateway, envelope, "ERROR", map[string]any{"error": "delegated capability failed", "peer": peer, "capability": capability, "details": err.Error()})
		return
	}
	g.base.respond(w, http.StatusOK, envelope, "TASK_RESULT", map[string]any{"delegated": true, "peer": peer, "capability": capability, "result": result["payload"], "completedAt": time.Now().UnixMilli()})
}

func isLocalOperation(engine *orchestrator.Engine, capability string) bool {
	for _, operation := range engine.Operations() {
		name := operation
		if i := strings.LastIndex(operation, "@"); i > 0 {
			name = operation[:i]
		}
		if name == capability {
			return true
		}
	}
	return false
}

func (g *FederatedGateway) handleControl(w http.ResponseWriter, in protocol.MeshEnvelope, capability string) {
	basePayload := map[string]any{"nucleus": protocol.N07, "contractVersion": protocol.SoulMeshContractVersion}
	switch strings.ToLower(capability) {
	case "mesh.ping":
		basePayload["ok"] = true
	case "core.health":
		for key, value := range g.engine.Health() {
			basePayload[key] = value
		}
	case "mesh.capability.resolve":
		requestedPayload := in.NestedPayload()
		requested, ok := requestedPayload["capability"].(string)
		if !ok || strings.TrimSpace(requested) == "" {
			g.base.respond(w, http.StatusBadRequest, in, "ERROR", map[string]any{"error": "mesh.capability.resolve requires payload.capability"})
			return
		}
		requested = strings.TrimSpace(requested)
		owner := ""
		local := isLocalOperation(g.engine, requested)
		if local {
			owner = protocol.N07
		} else if known := KnownFusionOwner(requested); known != "" {
			owner = known
		} else {
			for _, candidate := range g.peers.ConfiguredPeers() {
				if candidate.Circuit == CircuitOpen && time.Now().Before(candidate.RetryAfter) {
					continue
				}
				description, err := g.peers.Discover(rContextOrBackground(in), candidate.Nucleus)
				if err == nil && capabilityInDescription(description, requested) {
					owner = candidate.Nucleus
					break
				}
			}
		}
		basePayload["capability"] = requested
		basePayload["local"] = local
		basePayload["owner"] = owner
		basePayload["executable"] = local || owner != ""
	case "mesh.supergpu.describe":
		health := g.engine.Health()
		basePayload["superGPU"] = map[string]any{"local": true, "parallel": true, "federated": true, "health": health["compute"]}
	case "mesh.describe", "mesh.discovery", "mesh.capabilities", "mesh.fusion.describe":
		g.catalogMu.RLock()
		catalog := append([]FusionCapability(nil), g.catalog...)
		g.catalogMu.RUnlock()
		basePayload["operations"] = g.engine.Operations()
		basePayload["capabilityOwnership"] = catalog
		basePayload["topology"] = orchestrator.SOULTopology()
		basePayload["transports"] = []string{"IN_PROCESS", "LOOPBACK_HTTP", "HTTP", "REALTIME", "EVENT"}
	}
	g.base.respond(w, http.StatusOK, in, "TASK_RESULT", basePayload)
}

func rContextOrBackground(in protocol.MeshEnvelope) context.Context {
	_ = in
	return context.Background()
}

func (g *FederatedGateway) handleComposition(w http.ResponseWriter, r *http.Request, in protocol.MeshEnvelope) {
	payload := in.NestedPayload()
	if payload == nil {
		payload = map[string]any{}
	}
	raw, ok := payload["steps"].([]any)
	if !ok || len(raw) == 0 {
		g.base.respond(w, http.StatusBadRequest, in, "ERROR", map[string]any{"error": "fusion.execute requires payload.steps"})
		return
	}
	results := make([]map[string]any, 0, len(raw))
	for i, item := range raw {
		step, ok := item.(map[string]any)
		if !ok {
			g.base.respond(w, http.StatusBadRequest, in, "ERROR", map[string]any{"error": "fusion step must be object"})
			return
		}
		capability, _ := step["capability"].(string)
		capability = strings.TrimSpace(capability)
		if capability == "" {
			g.base.respond(w, http.StatusBadRequest, in, "ERROR", map[string]any{"error": "fusion step capability is required"})
			return
		}
		stepPayload, _ := step["payload"].(map[string]any)
		if stepPayload == nil {
			stepPayload = map[string]any{}
		}
		started := time.Now()
		peer, err := g.selectPeer(r.Context(), capability)
		entry := map[string]any{"index": i, "capability": capability, "duration_ms": time.Since(started).Milliseconds()}
		if err != nil {
			entry["status"] = "error"
			entry["error"] = err.Error()
			results = append(results, entry)
			if required, _ := step["required"].(bool); required {
				g.base.respond(w, http.StatusBadGateway, in, "ERROR", map[string]any{"steps": results})
				return
			}
			continue
		}
		result, err := g.peers.CallWithCorrelation(r.Context(), peer, capability, stepPayload, in.CorrelationID)
		entry["owner"] = peer
		if err != nil {
			entry["status"] = "error"
			entry["error"] = err.Error()
			results = append(results, entry)
			if required, _ := step["required"].(bool); required {
				g.base.respond(w, http.StatusBadGateway, in, "ERROR", map[string]any{"steps": results})
				return
			}
			continue
		}
		entry["status"] = "ok"
		entry["result"] = result["payload"]
		results = append(results, entry)
	}
	g.base.respond(w, http.StatusOK, in, "TASK_RESULT", map[string]any{"execution": "federated-composition", "correlationId": in.CorrelationID, "steps": results})
}

func (g *FederatedGateway) selectPeer(ctx context.Context, capability string) (string, error) {
	configured := g.peers.ConfiguredPeers()
	var best string
	var bestLatency time.Duration
	var lastErr error
	for _, peer := range configured {
		if peer.Circuit == CircuitOpen && time.Now().Before(peer.RetryAfter) {
			continue
		}
		desc, err := g.peers.Discover(ctx, peer.Nucleus)
		if err != nil {
			lastErr = err
			continue
		}
		if !capabilityInDescription(desc, capability) {
			continue
		}
		if known := KnownFusionOwner(capability); known != "" && known != peer.Nucleus {
			continue
		}
		fresh, ok := g.peerSnapshot(peer.Nucleus)
		if !ok {
			continue
		}
		if best == "" || fresh.Latency < bestLatency {
			best = fresh.Nucleus
			bestLatency = fresh.Latency
		}
	}
	if best == "" {
		if lastErr != nil {
			return "", lastErr
		}
		return "", errors.New("capability not discovered on configured peers: " + capability)
	}
	return best, nil
}

func (g *FederatedGateway) peerSnapshot(nucleus string) (PeerInfo, bool) {
	for _, peer := range g.peers.ConfiguredPeers() {
		if peer.Nucleus == nucleus {
			return peer, true
		}
	}
	return PeerInfo{}, false
}

func capabilityInDescription(desc map[string]any, target string) bool {
	for _, key := range []string{"capabilities", "operations", "executableCapabilities"} {
		items, ok := desc[key].([]any)
		if !ok {
			continue
		}
		for _, raw := range items {
			if value, ok := raw.(string); ok {
				if strings.TrimSpace(strings.SplitN(value, "@", 2)[0]) == target {
					return true
				}
			}
		}
	}
	return false
}

func ioReadLimited(r io.Reader, max int64) ([]byte, error) {
	if r == nil {
		return nil, errors.New("request body is nil")
	}
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, errors.New("request body exceeds configured limit")
	}
	return data, nil
}

func ioNopCloser(r io.Reader) io.ReadCloser { return io.NopCloser(r) }
