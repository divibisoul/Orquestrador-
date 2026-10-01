package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type CapabilityResolutionSource interface {
	ResolveCapability(context.Context, string, string) (map[string]any, error)
}

func RegisterCapabilityResolutionOperation(e *Engine, resolver CapabilityResolutionSource) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if resolver == nil {
		return errors.New("capability resolution source is required")
	}
	return e.Register("mesh.capability.resolve@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		capability := strings.TrimSpace(m.Metadata["capability"])
		if capability == "" {
			return protocol.Result{}, errors.New("metadata.capability is required")
		}
		resolution, err := resolver.ResolveCapability(ctx, capability, m.CorrelationID)
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.discovery", Target: m.Source, Status: "error", Error: err.Error()}, err
		}
		raw, err := json.Marshal(resolution)
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.discovery", Target: m.Source, Status: "error", Error: err.Error()}, err
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.discovery", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"capability": capability, "resolution_json": string(raw)},
		}, nil
	})
}
