package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func registerSynergyOperations(e *Engine) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	registrations := []struct {
		name string
		handler Handler
	}{
		{"mesh.synergy.describe@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			sequence := FavoriteSynergySequence()
			raw, _ := json.Marshal(sequence)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.synergy", Target: m.Source, Status: "ok", Metadata: map[string]string{"sequence_json": string(raw), "sequence_id": sequence.ID}}, nil
		}},
		{"mesh.synergy.execute@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			invoker := e.synergy()
			if invoker == nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.synergy", Target: m.Source, Status: "blocked", Error: "synergy invoker is not configured"}, errors.New("synergy invoker is not configured")
			}
			sequence := FavoriteSynergySequence()
			if raw := strings.TrimSpace(m.Metadata["synergy_sequence_json"]); raw != "" {
				if err := json.Unmarshal([]byte(raw), &sequence); err != nil {
					return protocol.Result{}, errors.New("invalid synergy_sequence_json")
				}
			}
			var capabilities map[string]string
			if err := decodeMetadataJSON(m.Metadata["synergy_capabilities_json"], &capabilities); err != nil {
				return protocol.Result{}, err
			}
			payload := map[string]any{}
			if raw := strings.TrimSpace(m.Metadata["synergy_payload_json"]); raw != "" {
				if err := json.Unmarshal([]byte(raw), &payload); err != nil {
					return protocol.Result{}, errors.New("invalid synergy_payload_json")
				}
			} else {
				payload["values"] = append([]float64(nil), m.Payload...)
			}
			execution, err := ExecuteSynergyRoute(ctx, sequence, invoker, m.CorrelationID, capabilities, payload)
			raw, marshalErr := json.Marshal(execution)
			if marshalErr != nil {
				return protocol.Result{}, marshalErr
			}
			result := protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.synergy", Target: m.Source, Status: execution.Status, Metadata: map[string]string{"synergy_execution_json": string(raw)}}
			if err != nil {
				result.Error = err.Error()
			}
			return result, err
		}},
	}
	for _, item := range registrations {
		if err := e.Register(item.name, item.handler); err != nil {
			return err
		}
	}
	return nil
}
