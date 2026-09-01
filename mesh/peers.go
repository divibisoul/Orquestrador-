package mesh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type PeerInfo struct {
	Nucleus     string
	URL         string
	Healthy     bool
	Latency     time.Duration
	LastError   string
	Failures    int
	Circuit     CircuitState
	RetryAfter  time.Time
}

type PeerClient struct {
	mu      sync.RWMutex
	peers   map[string]PeerInfo
	client  *http.Client
	secret  string
	maxRetry int
	cooldown time.Duration
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
		peers:    peers,
		client:   client,
		secret:   strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")),
		maxRetry: 3,
		cooldown: 30 * time.Second,
	}, nil
}

func (p *PeerClient) Discover(ctx context.Context, nucleus string) (map[string]any, error) {
	return p.Call(ctx, nucleus, "mesh.discovery", map[string]any{"from": protocol.N07})
}

func (p *PeerClient) Call(ctx context.Context, nucleus, capability string, payload map[string]any) (map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	nucleus = strings.TrimSpace(nucleus)
	capability = strings.TrimSpace(capability)
	if nucleus == protocol.N07 {
		return nil, errors.New("N07 cannot call itself through peer transport")
	}
	if capability == "" {
		return nil, errors.New("capability is required")
	}
	return p.call(ctx, nucleus, capability, payload)
}

func (p *PeerClient) call(ctx context.Context, nucleus, capability string, payload map[string]any) (map[string]any, error) {
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

	correlation := protocol.NewTraceID()
	var lastErr error
	for attempt := 1; attempt <= p.maxRetry; attempt++ {
		env := protocol.MeshEnvelope{
			Version:         protocol.SoulMeshVersion,
			ContractVersion: protocol.SoulMeshContractVersion,
			MessageID:       protocol.NewTraceID(),
			Source:          protocol.N07,
			Target:          nucleus,
			Timestamp:       time.Now().UnixMilli(),
			Nonce:           protocol.NewTraceID(),
			CorrelationID:   correlation,
			Type:            "CAPABILITY_REQUEST",
			Payload:         map[string]any{"capability": capability, "payload": payload},
		}
		if err := protocol.SignHMAC(&env, p.secret); err != nil {
			return nil, err
		}
		body, err := protocol.EncodeMesh(env)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, peer.URL+"/api/soul-mesh", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Soul-Contract-Version", protocol.SoulMeshContractVersion)
		started := time.Now()
		resp, err := p.client.Do(req)
		latency := time.Since(started)
		if err != nil {
			lastErr = err
			p.recordFailure(nucleus, latency, err.Error(), attempt)
			if attempt < p.maxRetry && retryableNetworkError(err) {
				if err := waitBackoff(ctx, attempt); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		var result map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			p.recordFailure(nucleus, latency, decodeErr.Error(), attempt)
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
			p.recordFailure(nucleus, latency, lastErr.Error(), attempt)
			if attempt < p.maxRetry && resp.StatusCode >= 500 {
				if err := waitBackoff(ctx, attempt); err != nil {
					return nil, err
				}
				continue
			}
			return result, lastErr
		}
		if result["contractVersion"] != protocol.SoulMeshContractVersion {
			lastErr = errors.New("mesh response contract version mismatch")
			p.recordFailure(nucleus, latency, lastErr.Error(), attempt)
			return nil, lastErr
		}
		if result["correlationId"] != correlation {
			lastErr = errors.New("peer correlation mismatch")
			p.recordFailure(nucleus, latency, lastErr.Error(), attempt)
			return nil, lastErr
		}
		p.recordSuccess(nucleus, latency)
		return result, nil
	}
	if lastErr == nil {
		lastErr = errors.New("mesh request failed")
	}
	return nil, lastErr
}

func retryableNetworkError(err error) bool {
	return err != nil
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

func (p *PeerClient) recordFailure(nucleus string, latency time.Duration, lastError string, _ int) {
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

func (p *PeerClient) ConfiguredPeers() []PeerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]PeerInfo, 0, len(p.peers))
	for _, peer := range p.peers {
		out = append(out, peer)
	}
	return out
}
