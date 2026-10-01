package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/cooperation"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func RegisterCooperationOperations(e *Engine, c *cooperation.Coordinator) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if c == nil {
		return errors.New("cooperation coordinator is required")
	}
	if err := e.Register("cooperation.handshake@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		target := strings.TrimSpace(m.Metadata["target"])
		required := strings.TrimSpace(m.Metadata["capability"])
		if target == "" {
			return protocol.Result{}, errors.New("metadata.target is required")
		}
		handshake, err := c.Handshake(ctx, target, required, m.CorrelationID)
		raw, _ := json.Marshal(handshake)
		result := protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.cooperation", Target: m.Source,
			Status: handshake.Status,
			Metadata: map[string]string{"handshake_json": string(raw)},
		}
		if err != nil {
			result.Error = err.Error()
		}
		return result, err
	}); err != nil {
		return err
	}
	if err := e.Register("cooperation.exchange@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		target := strings.TrimSpace(m.Metadata["target"])
		capability := strings.TrimSpace(m.Metadata["capability"])
		if target == "" || capability == "" {
			return protocol.Result{}, errors.New("metadata.target and metadata.capability are required")
		}
		payload := map[string]any{}
		if raw := strings.TrimSpace(m.Metadata["payload"]); raw != "" {
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				return protocol.Result{}, errors.New("metadata.payload must be valid JSON object")
			}
		}
		upstream, err := c.Exchange(ctx, target, capability, payload, m.CorrelationID)
		result := protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.cooperation", Target: m.Source,
			Status: "ok",
		}
		if upstream != nil {
			encoded, _ := json.Marshal(upstream)
			result.Metadata = map[string]string{"exchange_json": string(encoded)}
		}
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
		}
		return result, err
	}); err != nil {
		return err
	}
	return e.Register("cooperation.health@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.cooperation", Target: m.Source, Status: "ok",
			Metadata: map[string]string{
				"mode": "canonical-mesh",
				"protocol": cooperation.ProtocolVersion,
				"contract_version": cooperation.ContractVersion,
				"outcome_observer": "learning-machine-when-attached",
			},
		}, nil
	})
}
