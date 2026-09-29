package rgo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type TrinitySource interface {
	RGOTrinity(context.Context, map[string]any, string) (map[string]any, error)
}

type HortaStageSink interface {
	CallWithCorrelation(context.Context, string, string, map[string]any, string) (map[string]any, error)
}

// RegisterTrinityOperation composes SARA's RGO/Trinity runtime with the existing
// N07 peer transport to N01 HortaCore. N07 remains the routing authority.
func RegisterTrinityOperation(e *orchestrator.Engine, source TrinitySource, horta HortaStageSink) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if source == nil {
		return errors.New("RGO Trinity source is required")
	}
	return e.Register("rgo.trinity.process@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		raw := strings.TrimSpace(message.Metadata["rgo_envelope_json"])
		if raw == "" {
			return protocol.Result{}, errors.New("metadata.rgo_envelope_json is required")
		}
		var envelope map[string]any
		if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
			return protocol.Result{}, errors.New("invalid RGO envelope JSON")
		}
		out, err := source.RGOTrinity(ctx, envelope, message.CorrelationID)
		if err != nil {
			return protocol.Result{
				TraceID: message.TraceID, CorrelationID: message.CorrelationID,
				Source: "N07.rgo.trinity", Target: message.Source,
				Status: "error", Error: err.Error(),
			}, err
		}

		stages, _ := out["stages"].([]any)
		hortaResults := make([]any, 0, len(stages))
		hortaOK := horta != nil
		if horta == nil {
			hortaOK = false
		}
		for _, rawStage := range stages {
			stage, ok := rawStage.(map[string]any)
			if !ok {
				hortaOK = false
				hortaResults = append(hortaResults, map[string]any{"status": "invalid_stage"})
				continue
			}
			if !hortaOK {
				hortaResults = append(hortaResults, map[string]any{"status": "BLOCKED_HORTA_SINK"})
				continue
			}
			result, sinkErr := horta.CallWithCorrelation(ctx, "N01", "rgo.hortacore.store", stage, message.CorrelationID)
			if sinkErr != nil {
				hortaOK = false
				hortaResults = append(hortaResults, map[string]any{"status": "BLOCKED_HORTA", "error": sinkErr.Error()})
				continue
			}
			hortaResults = append(hortaResults, result)
		}
		out["hortacore_persisted"] = hortaOK
		out["hortacore_results"] = hortaResults
		if !hortaOK {
			out["integration_status"] = "PARTIAL_BLOCKED_HORTA"
		} else {
			out["integration_status"] = "VALIDATED_WITH_HORTA"
		}

		rawOut, _ := json.Marshal(out)
		status := "ok"
		if !hortaOK {
			status = "blocked"
		}
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.rgo.trinity", Target: message.Source,
			Status: status,
			Metadata: map[string]string{"rgo_trinity_result_json": string(rawOut)},
		}, func() error {
			if hortaOK {
				return nil
			}
			return errors.New("RGO_TRINITY_HORTACORE_BLOCKED")
		}()
	})
}
