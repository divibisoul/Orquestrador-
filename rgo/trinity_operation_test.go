package rgo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type fakeTrinitySource struct{}
func (fakeTrinitySource) RGOTrinity(_ context.Context, _ map[string]any, _ string) (map[string]any, error) {
	return map[string]any{"stages":[]any{
		map[string]any{"stage":"RGO","cycle_id":"c1","finding_id":"f1","output_hash":"sha256:r","data":map[string]any{"ok":true}},
	}}, nil
}

type fakeHortaSink struct{ calls int }
func (f *fakeHortaSink) CallWithCorrelation(_ context.Context, nucleus, capability string, payload map[string]any, _ string) (map[string]any, error) {
	if nucleus != "N01" || capability != "rgo.hortacore.store" || payload["stage"] != "RGO" {
		return nil, context.Canceled
	}
	f.calls++
	return map[string]any{"persisted":true}, nil
}

func newTestEngine(t *testing.T) *orchestrator.Engine {
	t.Helper()
	n, err := neural.New(2, 0.05); if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(0.1, 4); if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g); if err != nil { t.Fatal(err) }
	return e
}

func TestRegisterTrinityOperationPersistsEveryStageToHortaSink(t *testing.T) {
	e := newTestEngine(t)
	sink := &fakeHortaSink{}
	if err := RegisterTrinityOperation(e, fakeTrinitySource{}, sink); err != nil { t.Fatal(err) }
	env := map[string]any{"schema_version":"1.0.0"}
	raw, _ := json.Marshal(env)
	result, err := e.Execute(context.Background(), "rgo.trinity.process@1.0.0", []float64{0}, map[string]string{
		"rgo_envelope_json": string(raw),
	})
	if err != nil || result.Status != "ok" || sink.calls != 1 {
		t.Fatalf("unexpected result: %#v err=%v calls=%d", result, err, sink.calls)
	}
}
