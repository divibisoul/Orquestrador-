package mesh

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

// DynamicRouteCandidate is an observable routing choice without exposing
// internal PeerClient state to callers.
type DynamicRouteCandidate struct {
	Nucleus    string        `json:"nucleus"`
	Latency    time.Duration `json:"latency"`
	Healthy    bool          `json:"healthy"`
	Failures   int           `json:"failures"`
	Capability bool          `json:"capability"`
	Score      float64       `json:"score"`
}

// CallBestDynamic resolves a capability across all configured peers in
// parallel, scores only executable candidates, and then invokes the best one.
// It preserves the existing CallBest contract while adding health, latency and
// failure-aware routing. Discovery is bounded to the six configured nuclei.
func (p *PeerClient) CallBestDynamic(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
	if ctx == nil {
		return nil, "", errors.New("context is nil")
	}
	capability = strings.TrimSpace(capability)
	correlation = strings.TrimSpace(correlation)
	if capability == "" {
		return nil, "", errors.New("capability is required")
	}
	if correlation == "" {
		return nil, "", errors.New("correlation is required")
	}

	peers := p.ConfiguredPeers()
	if len(peers) == 0 {
		return nil, "", errors.New("no configured peers")
	}

	candidates := make([]DynamicRouteCandidate, len(peers))
	var wg sync.WaitGroup
	for i, peer := range peers {
		wg.Add(1)
		go func(index int, candidatePeer PeerInfo) {
			defer wg.Done()
			candidate := DynamicRouteCandidate{
				Nucleus:  candidatePeer.Nucleus,
				Latency:  candidatePeer.Latency,
				Healthy:  candidatePeer.Healthy,
				Failures: candidatePeer.Failures,
			}
			if candidatePeer.Circuit == CircuitOpen && time.Now().Before(candidatePeer.RetryAfter) {
				candidates[index] = candidate
				return
			}
			discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			description, err := p.Discover(discoveryCtx, candidatePeer.Nucleus)
			cancel()
			if err == nil && supportsExecutableCapability(description, capability) {
				candidate.Capability = true
				candidate.Score = routeScore(candidate)
			}
			candidates[index] = candidate
		}(i, peer)
	}
	wg.Wait()

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Capability != candidates[j].Capability {
			return candidates[i].Capability
		}
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Nucleus < candidates[j].Nucleus
	})
	for _, candidate := range candidates {
		if !candidate.Capability {
			continue
		}
		result, err := p.CallWithCorrelation(ctx, candidate.Nucleus, capability, payload, correlation)
		if err == nil {
			return result, candidate.Nucleus, nil
		}
	}
	return nil, "", errors.New("no healthy peer exposes executable capability: " + capability)
}

func routeScore(candidate DynamicRouteCandidate) float64 {
	score := 100.0
	if candidate.Healthy {
		score += 20
	}
	if candidate.Latency > 0 {
		score -= float64(candidate.Latency.Milliseconds()) / 100.0
	}
	score -= float64(candidate.Failures * 10)
	return score
}
