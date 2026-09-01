package mesh

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half-open"
)

const defaultDiscoveryCacheTTL = 15 * time.Second

type PeerInfo struct {
	Nucleus    string
	URL        string
	Healthy    bool
	Latency    time.Duration
	LastError  string
	Failures   int
	Circuit    CircuitState
	RetryAfter time.Time
}

type discoveryCacheEntry struct {
	value     map[string]any
	expiresAt time.Time
}

type PeerClient struct {
	mu                sync.RWMutex
	peers             map[string]PeerInfo
	client            *http.Client
	secret            string
	maxRetry          int
	cooldown          time.Duration
	discoveryMu       sync.RWMutex
	discoveryCache    map[string]discoveryCacheEntry
	discoveryCacheTTL time.Duration
}

func NewPeerClient(client *http.Client) (*PeerClient, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	peers := map[string]PeerInfo{}
	for _, n := range []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06} {
		if u := strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_MESH_"+n+"_URL")), "/"); u != "" {
			peers[n] = PeerInfo{Nucleus: n, URL: u, Circuit: CircuitClosed}
		}
	}
	return &PeerClient{
		peers:             peers,
		client:            client,
		secret:            strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")),
		maxRetry:          3,
		cooldown:          30 * time.Second,
		discoveryCache:    make(map[string]discoveryCacheEntry),
		discoveryCacheTTL: defaultDiscoveryCacheTTL,
	}, nil
}

func clonePeerMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	copy := make(map[string]any, len(value))
	for key, item := range value {
		copy[key] = item
	}
	return copy
}

func (p *PeerClient) Discover(ctx context.Context, nucleus string) (map[string]any, error) {
	nucleus = strings.TrimSpace(nucleus)
	if nucleus == protocol.N07 {
		return nil, errors.New("N07 cannot discover itself through peer transport")
	}
	if cached, ok := p.discoveryFromCache(nucleus); ok {
		return cached, nil
	}
	result, err := p.CallWithCorrelation(ctx, nucleus, "mesh.discovery", map[string]any{"from": protocol.N07}, protocol.NewTraceID())
	if err == nil {
		p.storeDiscovery(nucleus, result)
		return clonePeerMap(result), nil
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	result, err = p.CallWithCorrelation(ctx, nucleus, "mesh.describe", map[string]any{"from": protocol.N07}, protocol.NewTraceID())
	if err != nil {
		return nil, err
	}
	p.storeDiscovery(nucleus, result)
	return clonePeerMap(result), nil
}

func (p *PeerClient) discoveryFromCache(nucleus string) (map[string]any, bool) {
	p.discoveryMu.RLock()
	entry, ok := p.discoveryCache[nucleus]
	p.discoveryMu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			p.discoveryMu.Lock()
			delete(p.discoveryCache, nucleus)
			p.discoveryMu.Unlock()
		}
		return nil, false
	}
	return clonePeerMap(entry.value), true
}

func (p *PeerClient) storeDiscovery(nucleus string, value map[string]any) {
	p.discoveryMu.Lock()
	p.discoveryCache[nucleus] = discoveryCacheEntry{value: clonePeerMap(value), expiresAt: time.Now().Add(p.discoveryCacheTTL)}
	p.discoveryMu.Unlock()
}

func (p *PeerClient) invalidateDiscovery(nucleus string) {
	p.discoveryMu.Lock()
	delete(p.discoveryCache, nucleus)
	p.discoveryMu.Unlock()
}

func (p *PeerClient) Call(ctx context.Context, nucleus, capability string, payload map[string]any) (map[string]any, error) {
	return p.CallWithCorrelation(ctx, nucleus, capability, payload, protocol.NewTraceID())
}

func (p *PeerClient) CallWithCorrelation(ctx context.Context, nucleus, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	nucleus = strings.TrimSpace(nucleus)
	capability = strings.TrimSpace(capability)
	correlation = strings.TrimSpace(correlation)
	if nucleus == protocol.N07 {
		return nil, errors.New("N07 cannot call itself through peer transport")
	}
	if capability == "" {
		return nil, errors.New("capability is required")
	}
	if correlation == "" {
		return nil, errors.New("correlation is required")
	}
	return p.call(ctx, nucleus, capability, payload, correlation)
}

func (p *PeerClient) CallBest(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
	if ctx == nil {
		return nil, "", errors.New("context is nil")
	}
	for _, peer := range p.ConfiguredPeers() {
		if peer.Circuit == CircuitOpen && time.Now().Before(peer.RetryAfter) {
			continue
		}
		discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		description, err := p.Discover(discoveryCtx, peer.Nucleus)
		cancel()
		if err != nil || !supportsExecutableCapability(description, capability) {
			continue
		}
		result, err := p.CallWithCorrelation(ctx, peer.Nucleus, capability, payload, correlation)
		if err == nil {
			return result, peer.Nucleus, nil
		}
	}
	return nil, "", fmt.Errorf("no healthy peer exposes executable capability: %s", capability)
}

func supportsExecutableCapability(description map[string]any, capability string) bool {
	capability = strings.TrimSpace(capability)
	if capability == "" || description == nil {
		return false
	}
	raw, ok := description["executableCapabilities"]
	if !ok {
		if nested, ok := description["payload"].(map[string]any); ok {
			raw = nested["executableCapabilities"]
		}
	}
	items, ok := raw.([]any)
	if !ok {
		if values, ok := raw.([]string); ok {
			items = make([]any, len(values))
			for i, value := range values {
				items[i] = value
			}
		} else {
			return false
		}
	}
	requestedName, requestedVersion := splitPeerCapabilityVersion(capability)
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			continue
		}
		name, version := splitPeerCapabilityVersion(strings.TrimSpace(value))
		if name == requestedName && (requestedVersion == "" || requestedVersion == version) {
			return true
		}
	}
	return false
}

func splitPeerCapabilityVersion(value string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(value), "@", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func (p *PeerClient) call(ctx context.Context, nucleus, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	p.mu.RLock()
	peer, ok := p.peers[nucleus]
	p.mu.RUnlock()
	if !ok {
		return nil, errors.New("peer not configured: " + nucleus)
	}
	if peer.Circuit == CircuitOpen {
		if time.Now().Before(peer.RetryAfter) {
			return nil, fmt.Errorf("peer circuit open: %s", nucleus)
		}
		p.mu.Lock()
		peer = p.peers[nucleus]
		peer.Circuit = CircuitHalfOpen
		p.peers[nucleus] = peer
		p.mu.Unlock()
	}
	if p.secret == "" {
		return nil, errors.New("SOUL_MESH_HMAC_SECRET is not configured")
	}

	var lastErr error
	for attempt := 1; attempt <= p.maxRetry; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		messageID := protocol.NewTraceID()
		nonce := protocol.NewTraceID()
		wirePayload := payload
		if wirePayload == nil {
			wirePayload = map[string]any{}
		}
		wire := canonicalWireEnvelope{Protocol: "soul-mesh/1", ContractVersion: protocol.SoulMeshContractVersion, ID: messageID, CorrelationID: correlation, Source: protocol.N07, Target: nucleus, Kind: "request", Capability: capability, Payload: wirePayload, Timestamp: time.Now().UnixMilli(), Nonce: nonce}
		unsigned, err := canonicalN01Bytes(wire, nonce)
		if err != nil {
			return nil, err
		}
		mac := hmac.New(sha256.New, []byte(p.secret))
		_, _ = mac.Write(unsigned)
		wire.HMAC = hex.EncodeToString(mac.Sum(nil))
		body, err := json.Marshal(wire)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, peer.URL+"/api/soul-mesh", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Soul-Contract-Version", protocol.SoulMeshContractVersion)
		req.Header.Set("X-Soul-Correlation-Id", correlation)
		req.Header.Set("X-Soul-Mesh-Nonce", nonce)
		req.Header.Set("X-Soul-Mesh-HMAC", wire.HMAC)
		started := time.Now()
		resp, err := p.client.Do(req)
		latency := time.Since(started)
		if err != nil {
			lastErr = err
			p.recordFailure(nucleus, latency, err.Error())
			if attempt < p.maxRetry {
				if err := waitBackoff(ctx, attempt); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}
		var result map[string]any
		decodeErr := jsonDecodeLimited(resp, &result)
		resp.Body.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			p.recordFailure(nucleus, latency, decodeErr.Error())
			if attempt < p.maxRetry {
				if err := waitBackoff(ctx, attempt); err != nil {
					return nil, err
				}
				continue
			}
			return nil, decodeErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("peer request failed: %s", resp.Status)
			p.recordFailure(nucleus, latency, lastErr.Error())
			if attempt < p.maxRetry && (resp.StatusCode == 408 || resp.StatusCode == 429 || resp.StatusCode >= 500) {
				if err := waitBackoff(ctx, attempt); err != nil {
					return nil, err
				}
				continue
			}
			return result, lastErr
		}
		if result["contractVersion"] != protocol.SoulMeshContractVersion {
			lastErr = errors.New("mesh response contract version mismatch")
			p.recordFailure(nucleus, latency, lastErr.Error())
			return nil, lastErr
		}
		if result["protocol"] != "soul-mesh/1" {
			lastErr = errors.New("mesh response protocol mismatch")
			p.recordFailure(nucleus, latency, lastErr.Error())
			return nil, lastErr
		}
		if result["correlationId"] != correlation {
			lastErr = errors.New("peer correlation mismatch")
			p.recordFailure(nucleus, latency, lastErr.Error())
			return nil, lastErr
		}
		if result["kind"] != "response" && result["kind"] != "error" {
			lastErr = errors.New("mesh response kind mismatch")
			p.recordFailure(nucleus, latency, lastErr.Error())
			return nil, lastErr
		}
		if err := verifyResponseHMAC(result, p.secret); err != nil {
			lastErr = err
			p.recordFailure(nucleus, latency, err.Error())
			return nil, err
		}
		p.recordSuccess(nucleus, latency)
		return result, nil
	}
	if lastErr == nil {
		lastErr = errors.New("mesh request failed")
	}
	return nil, lastErr
}

func jsonDecodeLimited(resp *http.Response, out *map[string]any) error {
	if resp == nil || resp.Body == nil {
		return errors.New("empty peer response")
	}
	if resp.ContentLength > 1<<20 {
		return errors.New("peer response exceeds configured limit")
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
}

func waitBackoff(ctx context.Context, attempt int) error {
	delay := time.Duration(1<<uint(attempt-1)) * 250 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (p *PeerClient) recordSuccess(nucleus string, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer, ok := p.peers[nucleus]
	if !ok {
		return
	}
	peer.Healthy = true
	peer.Latency = latency
	peer.LastError = ""
	peer.Failures = 0
	peer.Circuit = CircuitClosed
	peer.RetryAfter = time.Time{}
	p.peers[nucleus] = peer
}

func (p *PeerClient) recordFailure(nucleus string, latency time.Duration, lastError string) {
	p.invalidateDiscovery(nucleus)
	p.mu.Lock()
	defer p.mu.Unlock()
	peer, ok := p.peers[nucleus]
	if !ok {
		return
	}
	peer.Healthy = false
	peer.Latency = latency
	peer.LastError = lastError
	peer.Failures++
	if peer.Failures >= 3 {
		peer.Circuit = CircuitOpen
		peer.RetryAfter = time.Now().Add(p.cooldown)
	}
	p.peers[nucleus] = peer
}

func verifyResponseHMAC(result map[string]any, secret string) error {
	hmacValue, _ := result["hmac"].(string)
	nonce, _ := result["nonce"].(string)
	if hmacValue == "" || nonce == "" {
		return errors.New("peer response HMAC credentials are missing")
	}
	if _, err := hex.DecodeString(hmacValue); err != nil {
		return fmt.Errorf("invalid peer response HMAC encoding: %w", err)
	}
	timestamp, ok := result["timestamp"].(float64)
	if !ok {
		return errors.New("peer response timestamp is invalid")
	}
	if delta := time.Now().UnixMilli() - int64(timestamp); delta > 30000 || delta < -30000 {
		return errors.New("peer response timestamp outside accepted clock skew")
	}
	id, _ := result["id"].(string)
	source, _ := result["source"].(string)
	target, _ := result["target"].(string)
	correlation, _ := result["correlationId"].(string)
	capability, _ := result["capability"].(string)
	kind, _ := result["kind"].(string)
	typ := "TASK_RESULT"
	if kind == "error" {
		typ = "ERROR"
	}
	payload := map[string]any{}
	if value, ok := result["payload"].(map[string]any); ok {
		payload = value
	}
	env := protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: protocol.SoulMeshContractVersion, MessageID: id, Source: source, Target: target, Timestamp: int64(timestamp), Nonce: nonce, CorrelationID: correlation, Type: typ, HMAC: hmacValue, Payload: map[string]any{"capability": capability, "payload": payload}}
	return protocol.VerifyHMAC(env, secret, time.Now())
}

func (p *PeerClient) ConfiguredPeers() []PeerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]PeerInfo, 0, len(p.peers))
	for _, peer := range p.peers {
		out = append(out, peer)
	}
	return out
}
