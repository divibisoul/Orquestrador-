package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/divibisoul/Orquestrador-/composition"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func RegisterPrimordialCompositionOperation(e *Engine) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	return e.Register("composition.primordial.resolve@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		select {
		case <-ctx.Done():
			return protocol.Result{}, ctx.Err()
		default:
		}
		path := strings.TrimSpace(message.Metadata["ledger_path"])
		if path == "" {
			path = strings.TrimSpace(os.Getenv("SOUL_PRIMORDIAL_LEDGER_PATH"))
		}
		if path == "" {
			path = "soul-nuclei.json"
		}
		ledger, err := composition.LoadLedger(path)
		if err != nil {
			return protocol.Result{}, err
		}

		mode := strings.TrimSpace(message.Metadata["mode"])
		var value any
		switch mode {
		case "", "pair":
			a := strings.TrimSpace(message.Metadata["participant_a"])
			b := strings.TrimSpace(message.Metadata["participant_b"])
			if a == "" || b == "" {
				return protocol.Result{}, errors.New("participant_a and participant_b are required")
			}
			value, err = composition.ResolvePair(ledger, a, b)
		case "matrix":
			value, err = composition.Matrix(ledger)
		case "profile":
			id := strings.TrimSpace(message.Metadata["participant"])
			value, err = composition.ResolveAIProfile(ledger, id)
		case "profiles":
			value, err = composition.AIProfiles(ledger)
		case "composition":
			rawParticipants := strings.TrimSpace(message.Metadata["participants"])
			if rawParticipants == "" {
				return protocol.Result{}, errors.New("participants are required")
			}
			parts := strings.Split(rawParticipants, ",")
			value, err = composition.ResolveComposition(ledger, parts...)
		default:
			return protocol.Result{}, errors.New("unsupported primordial composition mode")
		}
		if err != nil {
			return protocol.Result{}, err
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.primordial", Target: message.Source, Status: "ok",
			Metadata: map[string]string{"primordial_json": string(raw), "ledger_path": path, "mode": func() string {
				if mode == "" {
					return "pair"
				}
				return mode
			}()},
		}, nil
	})
}
