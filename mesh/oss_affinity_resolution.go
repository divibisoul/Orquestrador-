package mesh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type OSSAffinityTargetState string

const (
	OSSAffinityReal         OSSAffinityTargetState = "REAL"
	OSSAffinityProjected    OSSAffinityTargetState = "PROJECTED"
	OSSAffinityBlocked      OSSAffinityTargetState = "BLOCKED"
	OSSAffinityUnmeasurable OSSAffinityTargetState = "UNMEASURABLE"
)

type OSSAffinityTargetReport struct {
	Capability string                 `json:"capability"`
	Nucleus    string                 `json:"nucleus"`
	Rank       int                    `json:"rank"`
	Configured bool                   `json:"configured"`
	Executable bool                   `json:"executable"`
	State      OSSAffinityTargetState `json:"state"`
	Error      string                 `json:"error,omitempty"`
}

type OSSAffinityResolution struct {
	Capability string                    `json:"capability"`
	Targets    []OSSAffinityTargetReport `json:"targets"`
}

func (p *PeerClient) ResolveOSSAffinity(ctx context.Context, capability, correlation string) (OSSAffinityResolution, error) {
	if p == nil {
		return OSSAffinityResolution{}, errors.New("peer client is nil")
	}
	if ctx == nil {
		return OSSAffinityResolution{}, errors.New("context is nil")
	}
	capability = strings.TrimSpace(capability)
	correlation = strings.TrimSpace(correlation)
	if capability == "" {
		return OSSAffinityResolution{}, errors.New("capability is required")
	}
	if correlation == "" {
		return OSSAffinityResolution{}, errors.New("correlation is required")
	}

	targets := runtimeOSSAffinityTargets(capability)
	if len(targets) == 0 {
		return OSSAffinityResolution{Capability: capability, Targets: []OSSAffinityTargetReport{}}, nil
	}

	configured := make(map[string]struct{})
	for _, peer := range p.ConfiguredPeers() {
		configured[peer.Nucleus] = struct{}{}
	}

	reports := make([]OSSAffinityTargetReport, len(targets))
	var wg sync.WaitGroup
	for rank, nucleus := range targets {
		wg.Add(1)
		go func(rank int, nucleus string) {
			defer wg.Done()
			report := OSSAffinityTargetReport{
				Capability: capability,
				Nucleus:    nucleus,
				Rank:       rank,
				State:      OSSAffinityBlocked,
			}

			var description map[string]any
			if _, ok := configured[nucleus]; ok {
				report.Configured = true
				discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				description, err = p.Discover(discoveryCtx, nucleus)
				cancel()
				if err != nil {
					report.State = OSSAffinityUnmeasurable
					report.Error = err.Error()
					reports[rank] = report
					return
				}
			} else if probe := p.affinityProbe(nucleus); probe != nil {
				report.Configured = true
				probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				description, err = probe(probeCtx, correlation)
				cancel()
				if err != nil {
					report.State = OSSAffinityUnmeasurable
					report.Error = err.Error()
					reports[rank] = report
					return
				}
			} else {
				report.Error = "peer transport is not configured"
				reports[rank] = report
				return
			}

			report.Executable = supportsExecutableCapability(description, capability)
			if report.Executable {
				report.State = OSSAffinityReal
			} else {
				report.State = OSSAffinityProjected
				report.Error = fmt.Sprintf("target is reachable but does not advertise executable capability %q", capability)
			}
			reports[rank] = report
		}(rank, nucleus)
	}
	wg.Wait()

	return OSSAffinityResolution{Capability: capability, Targets: reports}, nil
}

type CapabilityResolutionSource interface {
	ResolveCapability(ctx context.Context, capability, correlation string) (map[string]any, error)
}

func (p *PeerClient) ResolveCapability(ctx context.Context, capability, correlation string) (map[string]any, error) {
	resolution, err := p.ResolveOSSAffinity(ctx, capability, correlation)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(resolution)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	return document, nil
}
