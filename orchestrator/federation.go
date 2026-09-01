package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type Peer struct {
	Nucleus      string
	URL          string
	Capabilities []string
	Healthy      bool
	Latency      time.Duration
	SuccessRate  float64
	InFlight     int
	Adapter      *mesh.Adapter
}

type peerState struct {
	Peer
	failures  int
	openUntil time.Time
	probes    uint64
	calls     uint64
	success   uint64
	latency   time.Duration
}

type Candidate struct {
	Nucleus     string
	Capability  string
	Score       float64
	Latency     time.Duration
	SuccessRate float64
	Healthy     bool
	InFlight    int
}

type Federation struct {
	mu              sync.RWMutex
	peers            map[string]*peerState
	failureThreshold int
	breakerWindow    time.Duration
	maxRetries       int
	baseBackoff      time.Duration
}

func NewFederation() *Federation {
	return &Federation{
		peers:            map[string]*peerState{},
		failureThreshold: 3,
		breakerWindow:    5 * time.Second,
		maxRetries:       2,
		baseBackoff:      50 * time.Millisecond,
	}
}

func (f *Federation) RegisterPeer(nucleus, baseURL string, capabilities []string) error {
	nucleus = strings.TrimSpace(nucleus)
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if nucleus == "" || baseURL == "" {
		return errors.New("nucleus and base URL are required")
	}
	if nucleus == protocol.N07 {
		return errors.New("N07 cannot register itself as a peer")
	}
	if !isBaseNucleus(nucleus) {
		return fmt.Errorf("unsupported peer nucleus: %s", nucleus)
	}
	adapter, err := mesh.New(baseURL)
	if err != nil {
		return err
	}
	caps := append([]string(nil), capabilities...)
	f.mu.Lock()
	f.peers[nucleus] = &peerState{Peer: Peer{
		Nucleus:      nucleus,
		URL:          baseURL,
		Capabilities: uniqueStrings(caps),
		Healthy:      true,
		SuccessRate:  1,
		Adapter:      adapter,
	}}
	f.mu.Unlock()
	return nil
}

func (f *Federation) RemovePeer(nucleus string) {
	f.mu.Lock()
	delete(f.peers, strings.TrimSpace(nucleus))
	f.mu.Unlock()
}

func (f *Federation) Snapshot() []Peer {
	f.mu.RLock()
	out := make([]Peer, 0, len(f.peers))
	for _, state := range f.peers {
		p := state.Peer
		p.Capabilities = append([]string(nil), state.Capabilities...)
		out = append(out, p)
	}
	f.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Nucleus < out[j].Nucleus })
	return out
}

func (f *Federation) DiscoverLive(ctx context.Context) ([]Peer, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	peers := f.Snapshot()
	if len(peers) == 0 {
		return nil, errors.New("no peers registered")
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for _, peer := range peers {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			payload, err := f.callWithRetry(ctx, peer.Nucleus, func() (map[string]any, error) {
				correlation := protocol.NewTraceID()
				envelope, err := peer.Adapter.Envelope("mesh.discovery", correlation, peer.Nucleus, map[string]any{})
				if err != nil {
					return nil, err
				}
				return peer.Adapter.Send(ctx, peer.URL, envelope)
			})
			f.record(peer.Nucleus, time.Since(start), err == nil)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("%s discovery failed: %w", peer.Nucleus, err)
				}
				mu.Unlock()
				return
			}
			caps := extractOperations(payload)
			f.mu.Lock()
			if state := f.peers[peer.Nucleus]; state != nil {
				if len(caps) > 0 {
					state.Capabilities = uniqueStrings(caps)
				}
				state.Healthy = true
			}
			f.mu.Unlock()
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return f.Snapshot(), firstErr
	}
	return f.Snapshot(), nil
}

func (f *Federation) Discover(capability string) []Candidate {
	capability = normalizeCapability(capability)
	f.mu.RLock()
	candidates := make([]Candidate, 0, len(f.peers))
	for _, state := range f.peers {
		if capability != "" && !hasCapability(state.Capabilities, capability) {
			continue
		}
		success := state.SuccessRate
		if state.calls > 0 {
			success = float64(state.success) / float64(state.calls)
		}
		latency := state.Latency
		if latency <= 0 {
			latency = state.latency
		}
		score := routeScore(state.Healthy, success, latency, state.InFlight)
		candidates = append(candidates, Candidate{
			Nucleus:     state.Nucleus,
			Capability:  capability,
			Score:       score,
			Latency:     latency,
			SuccessRate: success,
			Healthy:     state.Healthy,
			InFlight:    state.InFlight,
		})
	}
	f.mu.RUnlock()
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Nucleus < candidates[j].Nucleus
		}
		return candidates[i].Score > candidates[j].Score
	})
	return candidates
}

func (f *Federation) Delegate(ctx context.Context, traceID, capability string, payload map[string]any) (map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	capability = strings.TrimSpace(capability)
	if capability == "" {
		return nil, errors.New("capability is required")
	}
	candidates := f.Discover(capability)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no peer advertises capability: %s", capability)
	}
	var lastErr error
	for _, candidate := range candidates {
		peer, ok := f.getPeer(candidate.Nucleus)
		if !ok || !f.allow(peer.Nucleus) {
			continue
		}
		f.inFlight(peer.Nucleus, 1)
		start := time.Now()
		result, err := f.callWithRetry(ctx, peer.Nucleus, func() (map[string]any, error) {
			correlation := strings.TrimSpace(traceID)
			if correlation == "" {
				correlation = protocol.NewTraceID()
			}
			envelope, err := peer.Adapter.Envelope(capability, correlation, peer.Nucleus, payload)
			if err != nil {
				return nil, err
			}
			return peer.Adapter.Send(ctx, peer.URL, envelope)
		})
		f.inFlight(peer.Nucleus, -1)
		f.record(peer.Nucleus, time.Since(start), err == nil)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("all eligible peers are unavailable")
	}
	return nil, lastErr
}

func (f *Federation) getPeer(nucleus string) (Peer, bool) {
	f.mu.RLock()
	state, ok := f.peers[nucleus]
	if !ok {
		f.mu.RUnlock()
		return Peer{}, false
	}
	p := state.Peer
	p.Capabilities = append([]string(nil), state.Capabilities...)
	f.mu.RUnlock()
	return p, true
}

func (f *Federation) allow(nucleus string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	state, ok := f.peers[nucleus]
	if !ok {
		return false
	}
	if time.Now().Before(state.openUntil) {
		return false
	}
	return true
}

func (f *Federation) inFlight(nucleus string, delta int) {
	f.mu.Lock()
	if state, ok := f.peers[nucleus]; ok {
		state.InFlight += delta
		if state.InFlight < 0 {
			state.InFlight = 0
		}
	}
	f.mu.Unlock()
}

func (f *Federation) record(nucleus string, latency time.Duration, success bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state, ok := f.peers[nucleus]
	if !ok {
		return
	}
	state.calls++
	if success {
		state.success++
		state.failures = 0
		state.Healthy = true
	} else {
		state.failures++
		state.Healthy = false
		if state.failures >= f.failureThreshold {
			state.openUntil = time.Now().Add(f.breakerWindow)
			state.failures = 0
		}
	}
	if state.latency <= 0 {
		state.latency = latency
	} else {
		state.latency = time.Duration(float64(state.latency)*0.7 + float64(latency)*0.3)
	}
	state.Latency = state.latency
	state.SuccessRate = float64(state.success) / float64(state.calls)
}

func (f *Federation) callWithRetry(ctx context.Context, nucleus string, call func() (map[string]any, error)) (map[string]any, error) {
	var lastErr error
	for attempt := 0; attempt <= f.maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			backoff := f.baseBackoff * time.Duration(1<<(attempt-1))
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}
		result, err := call()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func routeScore(healthy bool, success float64, latency time.Duration, inFlight int) float64 {
	if !healthy {
		return 0
	}
	latencyMs := float64(latency) / float64(time.Millisecond)
	latencyPenalty := 0.0
	if latencyMs > 0 {
		latencyPenalty = math.Min(latencyMs/1000.0, 1)
	}
	loadPenalty := math.Min(float64(maxInt(inFlight, 0))/16.0, 1)
	return 0.55*success + 0.30*(1-latencyPenalty) + 0.15*(1-loadPenalty)
}

func extractOperations(payload map[string]any) []string {
	v, ok := payload["operations"]
	if !ok {
		return nil
	}
	values, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, item := range values {
		if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
			out = append(out, strings.TrimSpace(text))
		}
	}
	return out
}

func normalizeCapability(value string) string {
	value = strings.TrimSpace(value)
	if i := strings.LastIndex(value, "@"); i > 0 {
		return strings.TrimSpace(value[:i])
	}
	return value
}

func hasCapability(values []string, target string) bool {
	target = normalizeCapability(target)
	for _, value := range values {
		if normalizeCapability(value) == target {
			return true
		}
	}
	return false
}

func isBaseNucleus(n string) bool {
	switch n {
	case protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06:
		return true
	default:
		return false
	}
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func maxInt(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
