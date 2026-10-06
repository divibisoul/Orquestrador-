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

// recordHortaStageFailure preserves the failed stage and its provenance so a
// transport/persistence failure becomes recoverable evidence instead of loss.
func recordHortaStageFailure(stage map[string]any, status string, sinkErr error) map[string]any {
	out := map[string]any{
		"status": status,
		"stage":  stage,
	}
	if sinkErr != nil {
		out["error"] = sinkErr.Error()
	}
	for _, key := range []string{
		"cycle_id",
		"finding_id",
		"output_hash",
		"eru_snapshot_hash",
		"rgo_evidence_chain_hash",
	} {
		if value, ok := stage[key]; ok {
			out[key] = value
		}
	}
	return out
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

		finalStatus, _ := out["final_status"].(string)
		trinityValidated := finalStatus == "VALIDATED"
		stages, _ := out["stages"].([]any)
		hortaResults := make([]any, 0, len(stages))
		hortaOK := horta != nil
		for _, rawStage := range stages {
			stage, ok := rawStage.(map[string]any)
			if !ok {
				hortaOK = false
				hortaResults = append(hortaResults, map[string]any{"status": "invalid_stage"})
				continue
			}
			// Rule of Gold transformation: isolate the failure to its stage,
			// preserve its evidence, and continue the remaining stages.
			if horta == nil {
				hortaOK = false
				hortaResults = append(hortaResults, recordHortaStageFailure(stage, "BLOCKED_HORTA_SINK", nil))
				continue
			}
			result, sinkErr := horta.CallWithCorrelation(ctx, "N01", "rgo.hortacore.store", stage, message.CorrelationID)
			if sinkErr != nil {
				hortaOK = false
				hortaResults = append(hortaResults, recordHortaStageFailure(stage, "BLOCKED_HORTA", sinkErr))
				continue
			}
			hortaResults = append(hortaResults, result)
		}
		out["hortacore_persisted"] = hortaOK
		out["hortacore_results"] = hortaResults
		switch {
		case !trinityValidated:
			out["integration_status"] = "BLOCKED_TRINITY_NOT_VALIDATED"
		case !hortaOK:
			out["integration_status"] = "PARTIAL_BLOCKED_HORTA"
		default:
			out["integration_status"] = "VALIDATED_WITH_HORTA"
		}

		rawOut, _ := json.Marshal(out)
		status := "ok"
		if !trinityValidated || !hortaOK {
			status = "blocked"
		}
		return protocol.Result{
				TraceID: message.TraceID, CorrelationID: message.CorrelationID,
				Source: "N07.rgo.trinity", Target: message.Source,
				Status:   status,
				Metadata: map[string]string{"rgo_trinity_result_json": string(rawOut)},
			}, func() error {
				if !trinityValidated {
					return errors.New("RGO_TRINITY_NOT_VALIDATED")
				}
				if !hortaOK {
					return errors.New("RGO_TRINITY_HORTACORE_BLOCKED")
				}
				return nil
			}()
	})
}
