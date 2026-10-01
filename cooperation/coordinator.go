package cooperation

import (
	"context"
	"errors"
	"strings"
	"time"

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
	Protocol        string   `json:"protocol"`
	ContractVersion string   `json:"contract_version"`
	Source          string   `json:"source"`
	Target          string   `json:"target"`
	CorrelationID   string   `json:"correlation_id"`
	Status          string   `json:"status"`
	Capabilities    []string `json:"capabilities,omitempty"`
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
	if ctx == nil {
		return Handshake{}, errors.New("context is nil")
	}
	target = strings.TrimSpace(target)
	requiredCapability = strings.TrimSpace(requiredCapability)
	correlation = strings.TrimSpace(correlation)
	if target == "" {
		return Handshake{}, errors.New("target is required")
	}
	if target == SourceNucleus {
		return Handshake{}, errors.New("cooperation cannot target its own N07 transport")
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
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	target = strings.TrimSpace(target)
	capability = strings.TrimSpace(capability)
	correlation = strings.TrimSpace(correlation)
	if target == "" || capability == "" {
		return nil, errors.New("target and capability are required")
	}
	if correlation == "" {
		correlation = protocol.NewTraceID()
	}
	_, err := c.Handshake(ctx, target, capability, correlation)
	if err != nil {
		c.observe(ctx, target, capability, correlation, false)
		return nil, err
	}
	result, err := c.peers.CallWithCorrelation(ctx, target, capability, payload, correlation)
	c.observe(ctx, target, capability, correlation, err == nil)
	return result, err
}

func (c *Coordinator) observe(ctx context.Context, target, capability, correlation string, success bool) {
	if c.observer == nil {
		return
	}
	observeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = c.observer.ObserveRoute(observeCtx, SourceNucleus, target, capability, correlation, success)
}

func extractCapabilities(value map[string]any) []string {
	if value == nil {
		return nil
	}
	out := make([]string, 0, 16)
	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			for key, item := range typed {
				if key == "capabilities" || key == "executableCapabilities" {
					switch list := item.(type) {
					case []any:
						for _, v := range list {
							if s, ok := v.(string); ok {
								out = append(out, strings.TrimSpace(s))
							}
						}
					case []string:
						out = append(out, list...)
					}
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
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
