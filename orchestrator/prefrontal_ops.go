package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
)

// registerPrefrontalExecutiveOperations restores the complete executive surface
// from the recovered N07 front without replacing the existing admission handler.
// Learning statistics and decision-outcome recording remain separate contracts.
func registerPrefrontalExecutiveOperations(e *Engine) error {
	registrations := []struct {
		name    string
		handler Handler
	}{
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
			if valueText == "" {
				return protocol.Result{}, errors.New("metadata.value must be a finite number")
			}
			var value float64
			if parsed, err := fmt.Sscanf(valueText, "%f", &value); err != nil || parsed != 1 || math.IsNaN(value) || math.IsInf(value, 0) {
				return protocol.Result{}, errors.New("metadata.value must be a finite number")
			}
			observation, err := e.cortex.RecordDecisionOutcome(m.Metadata["decision_id"], m.Metadata["outcome"], value)
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
	}
	for _, item := range registrations {
		if err := e.Register(item.name, item.handler); err != nil {
			return err
		}
	}
	return nil
}
