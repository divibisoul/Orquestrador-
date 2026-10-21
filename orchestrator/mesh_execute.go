package orchestrator

import (
	"context"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// ExecuteWithTrace executes a registered N07 operation while preserving the
// trace identity supplied by the external caller.
// Execute remains unchanged for legacy local callers that generate their own trace.
func (e *Engine) ExecuteWithTrace(ctx context.Context, traceID, source, operation string, payload []float64, metadata map[string]string) (protocol.Result, error) {
	return e.executeWithIdentity(ctx, traceID, "", source, operation, payload, metadata)
}

// ExecuteWithCorrelation executes a registered N07 operation while preserving
// an explicit caller correlationId. Trace identity remains independently generated
// unless a traceId is supplied through ExecuteWithTrace.
func (e *Engine) ExecuteWithCorrelation(ctx context.Context, correlationID, source, operation string, payload []float64, metadata map[string]string) (protocol.Result, error) {
	return e.executeWithIdentity(ctx, "", correlationID, source, operation, payload, metadata)
}

func (e *Engine) executeWithIdentity(ctx context.Context, traceID, correlationID, source, operation string, payload []float64, metadata map[string]string) (protocol.Result, error) {
	m := protocol.NewMessage(source, "N07", "command", operation, payload)
	if traceID != "" {
		m.TraceID = traceID
	}
	if correlationID != "" {
		m.CorrelationID = correlationID
	}
	m.Metadata = metadata
	if metadata != nil {
		if schema := metadata["schema"]; schema != "" {
			if err := validateSchema(schema, payload); err != nil {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07", Target: source, Status: "rejected", Error: err.Error()}, err
			}
		}
	}
	return e.Submit(ctx, m)
}
