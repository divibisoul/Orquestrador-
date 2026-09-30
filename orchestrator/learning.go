package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/learning"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/protocol"
)

// RegisterLearningOperations attaches the canonical N07 learning receiver.
// Existing neural.forward/neural.learn remain owned by Engine.registerBuiltins;
// this function adds only the missing receiver-side contracts.
func RegisterLearningOperations(e *Engine, machine *learning.Machine, n *neural.Network) error {
	if e == nil || machine == nil || n == nil {
		return errors.New("learning registration requires engine, machine and neural network")
	}
	if err := e.Register("learning.feedback@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		if len(message.Payload) != 2 {
			return protocol.Result{}, errors.New("learning feedback payload must be [reward, confidence]")
		}
		meta := message.Metadata
		exp := learning.Experience{
			ID: message.TraceID, TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: message.Source, Target: strings.TrimSpace(meta["learning_target"]),
			Capability: strings.TrimSpace(meta["learning_capability"]),
			EventType: learning.EventFeedback, Outcome: strings.TrimSpace(meta["learning_outcome"]),
			Reward: message.Payload[0], Confidence: message.Payload[1],
			Provenance: strings.TrimSpace(meta["learning_provenance"]), Metadata: meta,
		}
		if exp.Capability == "" {
			return protocol.Result{}, errors.New("learning_capability metadata is required")
		}
		result := protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.learning", Target: message.Source,
			Metadata: map[string]string{"capability": exp.Capability},
		}
		err := machine.Feedback(ctx, exp)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			return result, err
		}
		result.Status = "ok"
		result.Metadata["learned"] = "true"
		return result, nil
	}); err != nil {
		return err
	}

	if err := e.Register("neural.parameters@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		select {
		case <-ctx.Done():
			return protocol.Result{}, ctx.Err()
		default:
		}
		encoded, err := json.Marshal(n.Parameters())
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.neural", Target: message.Source, Status: "ok",
			Metadata: map[string]string{"parameters": string(encoded)},
		}, nil
	}); err != nil {
		return err
	}

	return nil
}
