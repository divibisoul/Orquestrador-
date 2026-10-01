package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/divibisoul/Orquestrador-/cognitive"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func RegisterExternalPlanningSourceOperation(e *Engine) error {
	return e.Register("cognitive.planning.sources@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		raw, err := json.Marshal(cognitive.ExternalPlanningSources)
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.cognitive", Target: m.Source, Status: "error", Error: err.Error()}, err
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.cognitive", Target: m.Source, Status: "ok",
			Metadata: map[string]string{
				"sources_json":       string(raw),
				"execution_boundary": "descriptive provenance only; upstream code is not copied or executed",
			},
		}, nil
	})
}
