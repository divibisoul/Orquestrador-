package rgo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type SaraSink interface {
	RGOIngest(context.Context, Envelope) (map[string]any, error)
}

func RegisterOperation(e *orchestrator.Engine, sink SaraSink) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	return e.Register("rgo.ingest@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		raw := strings.TrimSpace(message.Metadata["rgo_envelope_json"])
		if raw == "" {
			return protocol.Result{}, errors.New("metadata.rgo_envelope_json is required")
		}
		var env Envelope
		if err := json.Unmarshal([]byte(raw), &env); err != nil {
			return protocol.Result{}, errors.New("invalid RGO envelope JSON")
		}
		if err := env.Validate(); err != nil {
			return protocol.Result{}, err
		}
		env = DeriveDual(env)
		if sink == nil {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.rgo", Target: message.Source, Status: "blocked", Error: "RGO_SARA_SINK_UNAVAILABLE"}, errors.New("RGO_SARA_SINK_UNAVAILABLE")
		}
		out, err := sink.RGOIngest(ctx, env)
		if err != nil {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.rgo", Target: message.Source, Status: "error", Error: err.Error()}, err
		}
		rawOut, _ := json.Marshal(out)
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.rgo", Target: message.Source, Status: "ok",
			Metadata: map[string]string{"rgo_result_json": string(rawOut)},
		}, nil
	})
}
