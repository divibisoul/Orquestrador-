package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func RegisterCapabilityResolutionOperation(e *Engine, peers *mesh.PeerClient) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if peers == nil {
		return errors.New("mesh peer client is required")
	}
	return e.Register("mesh.capability.resolve@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		capability := strings.TrimSpace(m.Metadata["capability"])
		if capability == "" {
			return protocol.Result{}, errors.New("metadata.capability is required")
		}
		resolution, err := peers.ResolveOSSAffinity(ctx, capability, m.CorrelationID)
		if err != nil {
			return protocol.Result{
				TraceID: m.TraceID, CorrelationID: m.CorrelationID,
				Source: "N07.discovery", Target: m.Source, Status: "error", Error: err.Error(),
			}, err
		}
		raw, err := json.Marshal(resolution)
		if err != nil {
			return protocol.Result{
				TraceID: m.TraceID, CorrelationID: m.CorrelationID,
				Source: "N07.discovery", Target: m.Source, Status: "error", Error: err.Error(),
			}, err
		}
		metadata := map[string]string{
			"capability": capability,
			"resolution_json": string(raw),
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.discovery", Target: m.Source, Status: "ok", Metadata: metadata,
		}, nil
	})
}
