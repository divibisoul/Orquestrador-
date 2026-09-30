package cooperation

import (
	"context"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	ProtocolVersion = protocol.SoulMeshVersion
	ContractVersion = protocol.SoulMeshContractVersion
	SourceNucleus   = protocol.N07
)

type PeerTransport interface {
	Discover(context.Context, string) (map[string]any, error)
	CallWithCorrelation(context.Context, string, string, map[string]any, string) (map[string]any, error)
}

type OutcomeObserver interface {
	ObserveRoute(context.Context, string, string, string, string, bool) error
}

type Coordinator struct {
	peers    PeerTransport
	observer OutcomeObserver
}

type Handshake struct {
	Protocol        string
	ContractVersion string
	Source          string
	Target          string
	CorrelationID   string
	Status          string
	Capabilities    []string
}

func New(peers PeerTransport, observer ...OutcomeObserver) (*Coordinator, error) {
	if peers == nil {
		return nil, errors.New("cooperation requires Mesh peer transport")
	}
	var sink OutcomeObserver
	if len(observer) > 0 {
		sink = observer[0]
	}
	return &Coordinator{peers: peers, observer: sink}, nil
}

func (c *Coordinator) Handshake(ctx context.Context, target, requiredCapability, correlation string) (Handshake, error) {
	target = strings.TrimSpace(target)
	requiredCapability = strings.TrimSpace(requiredCapability)
	correlation = strings.TrimSpace(correlation)
	if ctx == nil {
		return Handshake{}, errors.New("context is nil")
	}
	if target == "" {
		return Handshake{}, errors.New("target is required")
	}
	if target == SourceNucleus {
		return Handshake{}, errors.New("cooperation target cannot be N07")
	}
	if correlation == "" {
		correlation = protocol.NewTraceID()
	}
	description, err := c.peers.Discover(ctx, target)
	if err != nil {
		return Handshake{
			Protocol: ProtocolVersion, ContractVersion: ContractVersion,
			Source: SourceNucleus, Target: target, CorrelationID: correlation,
			Status: "error",
		}, err
	}
	capabilities := extractCapabilities(description)
	if requiredCapability != "" && !containsCapability(capabilities, requiredCapability) {
		return Handshake{
			Protocol: ProtocolVersion, ContractVersion: ContractVersion,
			Source: SourceNucleus, Target: target, CorrelationID: correlation,
			Status: "capability_unavailable", Capabilities: capabilities,
		}, errors.New("required capability unavailable: " + requiredCapability)
	}
	return Handshake{
		Protocol: ProtocolVersion, ContractVersion: ContractVersion,
		Source: SourceNucleus, Target: target, CorrelationID: correlation,
		Status: "ready", Capabilities: capabilities,
	}, nil
}

func (c *Coordinator) Exchange(ctx context.Context, target, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	correlation = strings.TrimSpace(correlation)
	if correlation == "" {
		correlation = protocol.NewTraceID()
	}
	target = strings.TrimSpace(target)
	capability = strings.TrimSpace(capability)
	if target == "" || capability == "" {
		return nil, errors.New("target and capability are required")
	}
	_, err := c.Handshake(ctx, target, capability, correlation)
	if err != nil {
		if c.observer != nil {
			_ = c.observer.ObserveRoute(ctx, SourceNucleus, target, capability, correlation, false)
		}
		return nil, err
	}
	result, err := c.peers.CallWithCorrelation(ctx, target, capability, payload, correlation)
	if c.observer != nil {
		_ = c.observer.ObserveRoute(ctx, SourceNucleus, target, capability, correlation, err == nil)
	}
	return result, err
}

func extractCapabilities(value map[string]any) []string {
	if value == nil {
		return nil
	}
	out := make([]string, 0)
	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			if raw, ok := typed["capabilities"]; ok {
				switch list := raw.(type) {
				case []any:
					for _, item := range list {
						if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
							out = append(out, strings.TrimSpace(s))
						}
					}
				case []string:
					for _, s := range list {
						if strings.TrimSpace(s) != "" {
							out = append(out, strings.TrimSpace(s))
						}
					}
				}
			}
			for key, item := range typed {
				if key == "capabilities" {
					continue
				}
				walk(item)
			}
		case []any:
			for _, item := range typed {
				walk(item)
			}
		}
	}
	walk(value)
	return uniqueStrings(out)
}

func containsCapability(capabilities []string, wanted string) bool {
	wantedName, wantedVersion := splitCapabilityVersion(wanted)
	for _, candidate := range capabilities {
		name, version := splitCapabilityVersion(candidate)
		if name == wantedName && (wantedVersion == "" || wantedVersion == version) {
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

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
