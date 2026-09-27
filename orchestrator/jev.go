package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/jev"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func RegisterJevOperations(e *Engine, client *jev.Client) error {
	if e == nil { return errors.New("engine is required") }
	if client == nil { return errors.New("Jev client is required") }
	return e.Register("jev.systemone@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		state := strings.TrimSpace(message.Metadata["state"])
		questionsJSON := strings.TrimSpace(message.Metadata["questions_json"])
		if state == "" { return protocol.Result{}, errors.New("metadata.state is required") }
		if questionsJSON == "" { return protocol.Result{}, errors.New("metadata.questions_json is required") }
		var questions map[string]map[string]any
		if err := json.Unmarshal([]byte(questionsJSON), &questions); err != nil { return protocol.Result{}, errors.New("metadata.questions_json is invalid") }
		if len(questions) == 0 { return protocol.Result{}, errors.New("metadata.questions_json must contain at least one question") }
		response, err := client.SystemOne(ctx, state, questions)
		if err != nil { return protocol.Result{}, err }
		raw, err := json.Marshal(response); if err != nil { return protocol.Result{}, err }
		return protocol.Result{Source:"N07.jev", Target:message.Source, Status:"ok", Metadata:map[string]string{"decision_json":string(raw),"model":response.Model}}, nil
	})
}
