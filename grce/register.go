package grce

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

// Register installs the canonical GRCE operation into the existing N07 engine.
func Register(e *orchestrator.Engine, participant Participant) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	executor := New(participant)
	return e.Register("grce.cycle.execute@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		input := strings.TrimSpace(message.Metadata["grce_input"])
		if input == "" {
			input = strings.TrimSpace(message.Metadata["input"])
		}
		if input == "" {
			return protocol.Result{
				TraceID: message.TraceID, CorrelationID: message.CorrelationID,
				Source: "N07.grce", Target: message.Source, Status: "rejected",
			}, errors.New("GRCE_INPUT_REQUIRED")
		}

		cycle := executor.Execute(ctx, input, message.CorrelationID)
		raw, err := json.Marshal(cycle)
		if err != nil {
			return protocol.Result{}, err
		}
		metadata := map[string]string{
			"grce_state":        string(cycle.State),
			"grce_cycle_id":     cycle.CycleID,
			"grce_stages":       stringValue(len(cycle.Stages)),
			"grce_output_hash":  cycle.FinalOutputHash,
			"grce_evidence_json": string(raw),
		}
		status := "ok"
		if cycle.State != StateReal {
			status = "blocked"
		}
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.grce", Target: message.Source, Status: status,
			Metadata: metadata,
		}, func() error {
			if cycle.State != StateReal {
				if cycle.Error != "" {
					return errors.New(cycle.Error)
				}
				return errors.New("GRCE_NOT_REAL")
			}
			return nil
		}()
	})
}

func stringValue(v int) string {
	b, _ := json.Marshal(v)
	return string(b)
}
