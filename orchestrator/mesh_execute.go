package orchestrator

import (
	"context"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// ExecuteWithTrace executes a registered N07 operation while preserving the
// correlation/trace identity supplied by the external Soul Mesh request.
// Execute remains unchanged for legacy local callers that generate their own trace.
func (e *Engine) ExecuteWithTrace(ctx context.Context, traceID, source, operation string, payload []float64, metadata map[string]string) (protocol.Result, error) {
	m := protocol.NewMessage(source, "N07", "command", operation, payload)
	if traceID != "" {
		m.TraceID = traceID
	}
	m.Metadata = metadata
	if metadata != nil {
		if schema := metadata["schema"]; schema != "" {
			if err := validateSchema(schema, payload); err != nil {
				return protocol.Result{TraceID: m.TraceID, Source: "N07", Target: source, Status: "rejected", Error: err.Error()}, err
			}
		}
	}
	return e.Submit(ctx, m)
}
