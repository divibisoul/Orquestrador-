package hortacore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	OpDescribe = "hortacore.describe@1.0.0"
	OpHealth   = "hortacore.health@1.0.0"
	OpSync     = "hortacore.sync@1.0.0"
	OpDispatch = "hortacore.dispatch@1.0.0"
	OpSignal   = "hortacore.signal@1.0.0"
)

func RegisterOperations(engine *orchestrator.Engine, fusion *Fusion) error {
	if engine == nil {
		return errors.New("orchestrator engine is required")
	}
	if fusion == nil {
		return errors.New("HortaCore fusion is required")
	}
	registrations := map[string]orchestrator.Handler{
		OpDescribe: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
			raw, err := json.Marshal(fusion.Describe())
			return result(message, raw, err)
		},
		OpHealth: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			raw, err := json.Marshal(fusion.Health(ctx, message.CorrelationID))
			return result(message, raw, err)
		},
		OpSync: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			raw, err := json.Marshal(fusion.SyncMesh(ctx, message.CorrelationID))
			return result(message, raw, err)
		},
		OpSignal: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
			signal := message.Metadata["signal"]
			level := 0
			if text := message.Metadata["level"]; text != "" {
				if _, err := fmt.Sscanf(text, "%d", &level); err != nil {
					return result(message, nil, errors.New("hortacore.signal level must be integer 0..3"))
				}
			}
			state, err := fusion.ApplySignal(signal, level, message.CorrelationID)
			if err != nil {
				return result(message, nil, err)
			}
			raw, err := json.Marshal(state)
			return result(message, raw, err)
		},
		OpDispatch: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			processor := ProcessorID(message.Metadata["processor"])
			input := message.Metadata["input"]
			out := fusion.Dispatch(ctx, processor, input, message.CorrelationID)
			raw, err := json.Marshal(out)
			if err != nil {
				return result(message, nil, err)
			}
			if !out.OK && out.Error != nil {
				return result(message, raw, errors.New(out.Error.Code+":"+out.Error.Message))
			}
			return result(message, raw, nil)
		},
	}
	for name, handler := range registrations {
		if err := engine.Register(name, handler); err != nil {
			return err
		}
	}
	return nil
}

func result(message protocol.Message, raw []byte, err error) (protocol.Result, error) {
	metadata := map[string]string{"hortacore_json": ""}
	if len(raw) > 0 {
		metadata["hortacore_json"] = string(raw)
	}
	out := protocol.Result{
		TraceID:       message.TraceID,
		CorrelationID: message.CorrelationID,
		Source:        "N07.HortaCore",
		Target:        message.Source,
		Status:        "ok",
		Metadata:      metadata,
	}
	if err != nil {
		out.Status = "error"
		out.Error = err.Error()
		return out, err
	}
	return out, nil
}
