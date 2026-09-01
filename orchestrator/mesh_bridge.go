package orchestrator

import (
	"context"
	"errors"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// SubmitMesh executes a mesh envelope through the canonical N07 engine and
// converts the result back to the mesh-neutral envelope contract.
func (e *Engine) SubmitMesh(ctx context.Context, env protocol.MeshEnvelope) (protocol.MeshEnvelope, error) {
	if e == nil {
		return protocol.MeshEnvelope{}, errors.New("nil orchestrator")
	}
	msg, err := protocol.MessageFromMesh(env)
	if err != nil {
		return protocol.MeshEnvelope{}, err
	}
	result, execErr := e.Submit(ctx, msg)
	if result.TraceID == "" {
		result.TraceID = msg.TraceID
	}
	if result.Source == "" {
		result.Source = msg.Target
	}
	if result.Target == "" {
		result.Target = msg.Source
	}
	response := protocol.Message{
		Version:   protocol.Version,
		TraceID:   result.TraceID,
		Source:    result.Source,
		Target:    result.Target,
		Kind:      "result",
		Operation: msg.Operation,
		Payload:   append([]float64(nil), result.Payload...),
		Metadata:  cloneStringMap(result.Metadata),
	}
	meshResponse, convErr := protocol.MeshFromMessage(response)
	if convErr != nil {
		return protocol.MeshEnvelope{}, convErr
	}
	if execErr != nil {
		return meshResponse, execErr
	}
	return meshResponse, nil
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
