package rgo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type fakeSink struct{ called bool }

func (f *fakeSink) RGOIngest(_ context.Context, env Envelope) (map[string]any, error) {
	f.called = true
	return map[string]any{"finding_id": env.FindingID, "dual": env.Dual.Status}, nil
}

func TestRegisterOperationExecutesRealEnginePath(t *testing.T) {
	n, err := neural.New(2, 0.05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(0.1, 4)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil { t.Fatal(err) }

	sink := &fakeSink{}
	if err := RegisterOperation(e, sink); err != nil { t.Fatal(err) }

	env := validEnvelope()
	raw, err := json.Marshal(env)
	if err != nil { t.Fatal(err) }
	result, err := e.Execute(context.Background(), "rgo.ingest@1.0.0", []float64{0}, map[string]string{
		"rgo_envelope_json": string(raw),
	})
	if err != nil { t.Fatal(err) }
	if result.Status != "ok" || !sink.called { t.Fatalf("RGO execution failed: %#v", result) }
}

func TestRegisterOperationRejectsMissingEnvelope(t *testing.T) {
	n, err := neural.New(2, 0.05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(0.1, 4)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil { t.Fatal(err) }
	if err := RegisterOperation(e, &fakeSink{}); err != nil { t.Fatal(err) }

	_, err = e.Submit(context.Background(), protocol.NewMessage("N01", "N07", "command", "rgo.ingest@1.0.0", []float64{0}))
	if err == nil { t.Fatal("expected missing envelope error") }
}
