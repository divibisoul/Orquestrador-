package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/fusion"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

var canonicalFusionRegistry = fusion.NewRegistry()

// RegisterAdvancedOperations wires the canonical cognitive safety, dynamic fusion
// and federated compute surfaces into the same N07 operation registry.
func RegisterAdvancedOperations(e *Engine) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	registrations := []struct {
		name    string
		handler Handler
	}{
		{name: "mesh.synergy.describe@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			sequence := FavoriteSynergySequence()
			return protocol.Result{
				TraceID: m.TraceID, CorrelationID: m.CorrelationID,
				Source: "N07.synergy", Target: m.Source, Status: "ok",
				Metadata: map[string]string{
					"sequence_json": mustJSON(sequence),
					"sequence_id": sequence.ID,
				},
			}, nil
		}},
		{name: "mesh.synergy.execute@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
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
			result := protocol.Result{
				TraceID: m.TraceID, CorrelationID: m.CorrelationID,
				Source: "N07.synergy", Target: m.Source,
				Status: execution.Status,
				Metadata: map[string]string{"synergy_execution_json": string(raw)},
			}
			if err != nil {
				result.Error = err.Error()
			}
			return result, err
		}},
		{name: "prefrontal.admission@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var c struct {
				ID                                                string
				Cost, Risk, Utility, Uncertainty, Urgency, Impact float64
			}
			if err := decodeMetadataJSON(m.Metadata["candidate_json"], &c); err != nil {
				return protocol.Result{}, err
			}
			neocortex, err := prefrontal.NewNeocortex(e.cortex, e.neural)
			if err != nil {
				return protocol.Result{}, err
			}
			if taskID := strings.TrimSpace(m.Metadata["task_id"]); taskID != "" {
				if err := neocortex.SwitchTask(taskID); err != nil {
					return protocol.Result{}, err
				}
			}
			candidate, err := neocortex.Evaluate(ctx, c.ID, m.Payload, c.Risk, c.Cost, c.Urgency, c.Impact)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "rejected", Error: err.Error()}, err
			}
			if err := neocortex.UpdateWorkingMemory([]prefrontal.Candidate{candidate}); err != nil {
				return protocol.Result{}, err
			}
			decision, err := neocortex.Commit(candidate, "prefrontal-admission-approved")
			if err != nil {
				return protocol.Result{}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"decision_id": decision.ID, "score": floatString(decision.Score), "neural_dimensions": itoa(int(candidate.Context["neural_dimensions"].(int)))}}, nil
		}},
		{name: "prefrontal.plan@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var candidates []prefrontal.Candidate
			if err := decodeMetadataJSON(m.Metadata["candidates_json"], &candidates); err != nil {
				return protocol.Result{}, err
			}
			planned, err := e.cortex.Plan(candidates)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			b, _ := json.Marshal(planned)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"candidates_json": string(b), "count": itoa(len(planned))}}, nil
		}},
		{name: "prefrontal.prioritize@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var candidates []prefrontal.Candidate
			if err := decodeMetadataJSON(m.Metadata["candidates_json"], &candidates); err != nil {
				return protocol.Result{}, err
			}
			prioritized, err := e.cortex.Prioritize(candidates)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			b, _ := json.Marshal(prioritized)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"candidates_json": string(b), "count": itoa(len(prioritized))}}, nil
		}},
		{name: "prefrontal.inhibit@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var candidate prefrontal.Candidate
			if err := decodeMetadataJSON(m.Metadata["candidate_json"], &candidate); err != nil {
				return protocol.Result{}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"blocked": itoaBool(e.cortex.Inhibit(candidate))}}, nil
		}},
		{name: "prefrontal.select@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var candidates []prefrontal.Candidate
			if err := decodeMetadataJSON(m.Metadata["candidates_json"], &candidates); err != nil {
				return protocol.Result{}, err
			}
			selected, err := e.cortex.Select(candidates)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "rejected", Error: err.Error()}, err
			}
			b, _ := json.Marshal(selected)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"candidate_json": string(b), "selected_id": selected.ID}}, nil
		}},
		{name: "prefrontal.validate@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var candidate prefrontal.Candidate
			if err := decodeMetadataJSON(m.Metadata["candidate_json"], &candidate); err != nil {
				return protocol.Result{}, err
			}
			if err := e.cortex.ValidateAction(candidate); err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "rejected", Error: err.Error()}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"candidate_id": candidate.ID}}, nil
		}},
		{name: "prefrontal.commit@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			var candidate prefrontal.Candidate
			if err := decodeMetadataJSON(m.Metadata["candidate_json"], &candidate); err != nil {
				return protocol.Result{}, err
			}
			reason := strings.TrimSpace(m.Metadata["reason"])
			if reason == "" {
				return protocol.Result{}, errors.New("metadata.reason is required")
			}
			decision, err := e.cortex.Commit(candidate, reason)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "rejected", Error: err.Error()}, err
			}
			b, _ := json.Marshal(decision)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"decision_json": string(b), "decision_id": decision.ID}}, nil
		}},
		{name: "prefrontal.recall@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			limit := 0
			if raw := strings.TrimSpace(m.Metadata["limit"]); raw != "" {
				if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil || limit < 0 {
					return protocol.Result{}, errors.New("metadata.limit must be a non-negative integer")
				}
			}
			recalled := e.cortex.Recall(limit)
			b, _ := json.Marshal(recalled)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"decisions_json": string(b), "count": itoa(len(recalled))}}, nil
		}},
		{name: "prefrontal.task.switch@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			taskID := strings.TrimSpace(m.Metadata["task_id"])
			if err := e.cortex.SwitchTask(taskID); err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"task_id": taskID}}, nil
		}},
		{name: "prefrontal.working-memory@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			if raw := strings.TrimSpace(m.Metadata["candidates_json"]); raw != "" {
				var candidates []prefrontal.Candidate
				if err := decodeMetadataJSON(raw, &candidates); err != nil {
					return protocol.Result{}, err
				}
				if err := e.cortex.UpdateWorkingMemory(candidates); err != nil {
					return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "error", Error: err.Error()}, err
				}
			}
			limit := 0
			if raw := strings.TrimSpace(m.Metadata["limit"]); raw != "" {
				if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil || limit < 0 {
					return protocol.Result{}, errors.New("metadata.limit must be a non-negative integer")
				}
			}
			items := e.cortex.WorkingMemory(limit)
			b, _ := json.Marshal(items)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"candidates_json": string(b), "count": itoa(len(items))}}, nil
		}},
		{name: "prefrontal.outcome.observe@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			valueText := strings.TrimSpace(m.Metadata["value"])
			var value float64
			if valueText == "" {
				return protocol.Result{}, errors.New("metadata.value must be a finite number")
			}
			if parsed, err := fmt.Sscanf(valueText, "%f", &value); err != nil || parsed != 1 || math.IsNaN(value) || math.IsInf(value, 0) {
				return protocol.Result{}, errors.New("metadata.value must be a finite number")
			}
			observation, err := e.cortex.ObserveOutcome(m.Metadata["decision_id"], m.Metadata["outcome"], value)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			b, _ := json.Marshal(observation)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"observation_json": string(b)}}, nil
		}},
		{name: "prefrontal.monitor@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			state := e.cortex.Monitor(time.Now().UTC())
			b, _ := json.Marshal(state)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"state_json": string(b)}}, nil
		}},
		{name: "mesh.fusion.describe@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			b, err := json.Marshal(canonicalFusionRegistry.Snapshot())
			if err != nil {
				return protocol.Result{}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.fusion", Target: m.Source, Status: "ok", Metadata: map[string]string{"components_json": string(b), "component_count": itoa(len(canonicalFusionRegistry.Snapshot()))}}, nil
		}},
		{name: "mesh.fusion.execute@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			var ids []string
			if err := json.Unmarshal([]byte(m.Metadata["component_ids_json"]), &ids); err != nil || len(ids) < 2 {
				return protocol.Result{}, errors.New("metadata.component_ids_json must contain at least two component ids")
			}
			result, err := canonicalFusionRegistry.Fuse(ctx, ids, m.Payload)
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.fusion", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			trace, err := json.Marshal(result.Trace)
			if err != nil {
				return protocol.Result{}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.fusion", Target: m.Source, Status: "ok", Payload: result.Output, Metadata: map[string]string{"trace_json": string(trace)}}, nil
		}},
		{name: "supergpu.federated.execute@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			f, err := supergpu.NewFederation(e.compute)
			if err != nil {
				return protocol.Result{}, err
			}
			computeCtx := supergpu.WithCorrelationID(ctx, m.CorrelationID)
			result, err := f.Execute(computeCtx, supergpu.FederatedRequest{Nucleus: strings.TrimSpace(m.Metadata["nucleus"]), Operation: strings.TrimSpace(m.Metadata["operation"]), Payload: m.Payload, Device: strings.TrimSpace(m.Metadata["device"])})
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.supergpu", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.supergpu", Target: m.Source, Status: "ok", Payload: result.Output, Metadata: map[string]string{"nucleus": result.Nucleus, "device": result.Device.ID}}, nil
		}},
	}
	for _, item := range registrations {
		if err := e.Register(item.name, item.handler); err != nil {
			return err
		}
	}
	return nil
}
func RegisterFusionComponent(c fusion.Component) error { return canonicalFusionRegistry.Register(c) }
func decodeMetadataJSON(raw string, dst any) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New("metadata JSON is required")
	}
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		return errors.New("invalid metadata JSON")
	}
	return nil
}
func floatString(v float64) string { b, _ := json.Marshal(v); return string(b) }
func itoa(v int) string            { b, _ := json.Marshal(v); return strings.Trim(string(b), "\"") }
func itoaBool(v bool) string         { b, _ := json.Marshal(v); return strings.Trim(string(b), "\"") }


func mustJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
