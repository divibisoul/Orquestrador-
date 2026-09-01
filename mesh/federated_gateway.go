package mesh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

type FederatedGateway struct {
	base   *HTTPGateway
	peers  *PeerClient
	engine *orchestrator.Engine
}

func NewFederatedHTTPGateway(engine *orchestrator.Engine) *FederatedGateway {
	return &FederatedGateway{base: NewHTTPGateway(engine), peers: mustPeerClient(), engine: engine}
}

func mustPeerClient() *PeerClient {
	p, err := NewPeerClient(nil)
	if err != nil {
		return &PeerClient{}
	}
	return p
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
	if capability == "mesh.ping" || capability == "mesh.describe" || capability == "core.health" {
		r.Body = ioNopCloser(bytes.NewReader(body))
		g.base.Handler(w, r)
		return
	}
	peer, descErr := g.selectPeer(r.Context(), capability)
	if descErr != nil {
		g.base.respond(w, http.StatusServiceUnavailable, envelope, "ERROR", map[string]any{
			"error":      "no federated peer capability available",
			"capability": capability,
			"details":    descErr.Error(),
		})
		return
	}
	payload := envelope.NestedPayload()
	if payload == nil {
		payload = map[string]any{}
	}
	result, err := g.peers.CallWithCorrelation(r.Context(), peer, capability, payload, envelope.CorrelationID)
	if err != nil {
		g.base.respond(w, http.StatusBadGateway, envelope, "ERROR", map[string]any{
			"error":      "delegated capability failed",
			"peer":       peer,
			"capability": capability,
			"details":    err.Error(),
		})
		return
	}
	g.base.respond(w, http.StatusOK, envelope, "TASK_RESULT", map[string]any{
		"delegated":  true,
		"peer":       peer,
		"capability": capability,
		"result":     result["payload"],
		"completedAt": time.Now().UnixMilli(),
	})
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

func (g *FederatedGateway) selectPeer(ctx context.Context, capability string) (string, error) {
	configured := g.peers.ConfiguredPeers()
	var best string
	bestLatency := time.Duration(1<<63 - 1)
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
		if best == "" || peer.Latency < bestLatency {
			best = peer.Nucleus
			bestLatency = peer.Latency
		}
	}
	if best == "" {
		if lastErr != nil {
			return "", lastErr
		}
		return "", errors.New("capability not discovered as executable on configured peers: " + capability)
	}
	return best, nil
}

func capabilityInDescription(desc map[string]any, target string) bool {
	target = strings.TrimSpace(target)
	raw, ok := desc["executableCapabilities"]
	if !ok {
		return false
	}
	items, ok := raw.([]any)
	if !ok {
		return false
	}
	requestedName, requestedVersion := splitCapabilityVersion(target)
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			continue
		}
		name, version := splitCapabilityVersion(strings.TrimSpace(value))
		if name != requestedName {
			continue
		}
		if requestedVersion == "" || requestedVersion == version {
			return true
		}
	}
	return false
}

func splitCapabilityVersion(value string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(value), "@", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
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
