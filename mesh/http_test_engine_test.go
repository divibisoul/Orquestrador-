package mesh

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type meshHTTPTestEngine struct {
	neural  *neural.Network
	cortex  *prefrontal.Cortex
}

func newMeshHTTPTestEngine() (*meshHTTPTestEngine, error) {
	n, err := neural.New(8, .05)
	if err != nil {
		return nil, err
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		return nil, err
	}
	return &meshHTTPTestEngine{neural: n, cortex: c}, nil
}

func (e *meshHTTPTestEngine) Operations() []string {
	return []string{
		"mesh.ping@1.0.0",
		"mesh.describe@1.0.0",
		"neural.forward@1.0.0",
		"prefrontal.admission@1.0.0",
	}
}

func (e *meshHTTPTestEngine) Submit(ctx context.Context, message protocol.Message) (protocol.Result, error) {
	if e == nil || e.neural == nil || e.cortex == nil {
		return protocol.Result{}, errors.New("mesh test engine unavailable")
	}
	base := protocol.Result{
		TraceID: message.TraceID,
		CorrelationID: message.CorrelationID,
		Source: "N07",
		Target: message.Source,
		Status: "ok",
	}
	switch message.Operation {
	case "mesh.ping", "mesh.ping@1.0.0":
		base.Payload = map[string]any{"ok": true, "nucleus": "N07", "contractVersion": protocol.SoulMeshContractVersion}
		return base, nil
	case "mesh.describe", "mesh.describe@1.0.0":
		base.Payload = map[string]any{"nucleus": "N07", "operations": e.Operations()}
		return base, nil
	case "neural.forward", "neural.forward@1.0.0":
		values, err := e.neural.Forward(ctx, message.Payload)
		if err != nil {
			base.Status = "error"
			base.Error = err.Error()
			return base, err
		}
		base.Source = "N07.neural"
		base.Payload = values
		return base, nil
	case "prefrontal.admission", "prefrontal.admission@1.0.0":
		raw := message.Metadata["candidate_json"]
		var c struct {
			ID string
			Cost, Risk, Utility, Uncertainty, Urgency, Impact float64
		}
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			base.Status = "rejected"
			base.Error = err.Error()
			return base, err
		}
		neocortex, err := prefrontal.NewNeocortex(e.cortex, e.neural)
		if err != nil {
			base.Status = "error"
			base.Error = err.Error()
			return base, err
		}
		candidate, err := neocortex.Evaluate(ctx, c.ID, message.Payload, c.Risk, c.Cost, c.Urgency, c.Impact)
		if err != nil {
			base.Status = "rejected"
			base.Error = err.Error()
			return base, err
		}
		decision, err := neocortex.Commit(candidate, "mesh-test-prefrontal-admission")
		if err != nil {
			base.Status = "rejected"
			base.Error = err.Error()
			return base, err
		}
		base.Source = "N07.prefrontal"
		base.Metadata = map[string]string{
			"decision_id":      decision.ID,
			"neural_dimensions": string(mustJSON(candidate.Context["neural_dimensions"])),
		}
		return base, nil
	default:
		return protocol.Result{}, errors.New("unsupported mesh test operation: " + message.Operation)
	}
}

func mustJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}
