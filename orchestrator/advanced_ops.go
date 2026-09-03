package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

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
			candidate := prefrontal.Candidate{ID: c.ID, Cost: c.Cost, Risk: c.Risk, Utility: c.Utility, Uncertainty: c.Uncertainty, Urgency: c.Urgency, Impact: c.Impact}
			if err := e.cortex.ValidateAction(candidate); err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "rejected", Error: err.Error()}, err
			}
			decision, err := e.cortex.Commit(candidate, "prefrontal-admission-approved")
			if err != nil {
				return protocol.Result{}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal", Target: m.Source, Status: "ok", Metadata: map[string]string{"decision_id": decision.ID, "score": floatString(decision.Score)}}, nil
		}},
		{name: "mesh.fusion.describe@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			if err := ctx.Err(); err != nil {
				return protocol.Result{}, err
			}
			components := canonicalFusionRegistry.Snapshot()
			b, err := json.Marshal(components)
			if err != nil {
				return protocol.Result{}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.fusion", Target: m.Source, Status: "ok", Metadata: map[string]string{"components_json": string(b), "component_count": itoa(len(components))}}, nil
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
			trace, _ := json.Marshal(result.Trace)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.fusion", Target: m.Source, Status: "ok", Payload: result.Output, Metadata: map[string]string{"trace_json": string(trace)}}, nil
		}},
		{name: "supergpu.federated.execute@1.0.0", handler: func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			nucleus := strings.TrimSpace(m.Metadata["nucleus"])
			f, err := supergpu.NewFederation(e.compute)
			if err != nil {
				return protocol.Result{}, err
			}
			result, err := f.Execute(ctx, supergpu.FederatedRequest{Nucleus: nucleus, Operation: strings.TrimSpace(m.Metadata["operation"]), Payload: m.Payload, Device: strings.TrimSpace(m.Metadata["device"])})
			if err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.supergpu", Target: m.Source, Status: "error", Error: err.Error()}, err
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.supergpu", Target: m.Source, Status: "ok", Payload: result.Output, Metadata: map[string]string{"nucleus": result.Nucleus, "device": result.Device.ID}}, nil
		}},
	}
	for _, item := range registrations {
		if err := e.Register(item.name, item.handler); err != nil && !strings.Contains(err.Error(), "already registered") {
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
