package orchestrator

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// RegisterExternalLearningOperations exposes the SOUL-29 learning boundaries.
// The external providers remain fail-closed and are never reported ONLINE from
// structural attachment alone.
func RegisterExternalLearningOperations(e *Engine) error {
	if e == nil {
		return errors.New("learning external boundary requires engine")
	}
	boundary := DefaultExternalLearningBoundary()

	describe := func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		select {
		case <-ctx.Done():
			return protocol.Result{}, ctx.Err()
		default:
		}
		raw, err := json.Marshal(map[string]any{
			"state":                    "PROJECTED",
			"federated_provider":       boundary.FederatedProvider,
			"decentralized_dependency": boundary.DecentralizedDependency,
			"control_plane":            boundary.ControlPlane,
			"persistence_boundary":     boundary.PersistenceBoundary,
			"strategy_boundary":        boundary.StrategyBoundary,
			"transport_boundary":       boundary.TransportBoundary,
			"runtime_evidence_required": true,
		})
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID:      message.TraceID,
			CorrelationID: message.CorrelationID,
			Source:       "N07.learning",
			Target:       message.Source,
			Status:       "ok",
			Metadata:     map[string]string{"learning_boundary": string(raw)},
		}, nil
	}

	if err := e.Register("learning.federated.describe@1.0.0", describe); err != nil {
		return err
	}
	return e.Register("learning.decentralized.describe@1.0.0", describe)
}
