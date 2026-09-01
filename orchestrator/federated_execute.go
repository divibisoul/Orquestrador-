package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// FederatedExecution keeps the local execution path authoritative and falls
// back to a peer only when the requested operation is not locally registered.
// The external trace/correlation identity is preserved end-to-end.
type FederatedExecution struct {
	Engine     *Engine
	Federation *Federation
}

func NewFederatedExecution(engine *Engine, federation *Federation) (*FederatedExecution, error) {
	if engine == nil || federation == nil {
		return nil, errors.New("engine and federation are required")
	}
	return &FederatedExecution{Engine: engine, Federation: federation}, nil
}

func (f *FederatedExecution) Execute(ctx context.Context, traceID, source, operation string, payload []float64, metadata map[string]string) (protocol.Result, error) {
	if ctx == nil {
		return protocol.Result{}, errors.New("context is nil")
	}
	if strings.TrimSpace(operation) == "" {
		return protocol.Result{}, errors.New("operation is required")
	}

	message := protocol.NewMessage(source, protocol.N07, "command", operation, payload)
	if traceID != "" {
		message.TraceID = traceID
	}
	message.Metadata = metadata
	if _, err := f.Engine.Route(message); err == nil {
		return f.Engine.ExecuteWithTrace(ctx, message.TraceID, source, operation, payload, metadata)
	}

	requestPayload := map[string]any{"values": append([]float64(nil), payload...)}
	if metadata != nil {
		requestPayload["metadata"] = metadata
	}
	response, err := f.Federation.Delegate(ctx, message.TraceID, operation, requestPayload)
	if err != nil {
		return protocol.Result{TraceID: message.TraceID, Source: protocol.N07, Target: source, Status: "error", Error: err.Error()}, err
	}

	result := protocol.Result{TraceID: message.TraceID, Source: protocol.N07, Target: source, Status: "ok", Metadata: map[string]string{"delegated": "true"}}
	if value, ok := response["payload"].(map[string]any); ok {
		if values := decodeFloatSlice(value["values"]); values != nil {
			result.Payload = values
		}
		if status, ok := value["status"].(string); ok && status != "" {
			result.Status = status
		}
		if errText, ok := value["error"].(string); ok && errText != "" {
			result.Error = errText
			result.Status = "error"
		}
	}
	if result.Payload == nil {
		if values := decodeFloatSlice(response["values"]); values != nil {
			result.Payload = values
		}
	}
	if sourceValue, ok := response["source"].(string); ok && sourceValue != "" {
		result.Metadata["peer_source"] = sourceValue
	}
	return result, nil
}

func decodeFloatSlice(value any) []float64 {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var values []float64
	if err := json.Unmarshal(encoded, &values); err != nil {
		return nil
	}
	return values
}
