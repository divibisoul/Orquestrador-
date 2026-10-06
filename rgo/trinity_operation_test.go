package rgo

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type fakeTrinitySource struct{}

func (fakeTrinitySource) RGOTrinity(_ context.Context, _ map[string]any, _ string) (map[string]any, error) {
	return map[string]any{"final_status": "VALIDATED", "stages": []any{
		map[string]any{"stage": "RGO", "cycle_id": "c1", "finding_id": "f1", "output_hash": "sha256:r", "data": map[string]any{"ok": true}},
	}}, nil
}

type fakeHortaSink struct{ calls int }

func (f *fakeHortaSink) CallWithCorrelation(_ context.Context, nucleus, capability string, payload map[string]any, _ string) (map[string]any, error) {
	if nucleus != "N01" || capability != "rgo.hortacore.store" || payload["stage"] != "RGO" {
		return nil, context.Canceled
	}
	f.calls++
	return map[string]any{"persisted": true}, nil
}

func newTestEngine(t *testing.T) *orchestrator.Engine {
	t.Helper()
	n, err := neural.New(2, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestRegisterTrinityOperationPersistsEveryStageToHortaSink(t *testing.T) {
	e := newTestEngine(t)
	sink := &fakeHortaSink{}
	if err := RegisterTrinityOperation(e, fakeTrinitySource{}, sink); err != nil {
		t.Fatal(err)
	}
	env := map[string]any{"schema_version": "1.0.0"}
	raw, _ := json.Marshal(env)
	result, err := e.Execute(context.Background(), "rgo.trinity.process@1.0.0", []float64{0}, map[string]string{
		"rgo_envelope_json": string(raw),
	})
	if err != nil || result.Status != "ok" || sink.calls != 1 {
		t.Fatalf("unexpected result: %#v err=%v calls=%d", result, err, sink.calls)
	}
}

type inconclusiveTrinitySource struct{}

func (inconclusiveTrinitySource) RGOTrinity(_ context.Context, _ map[string]any, _ string) (map[string]any, error) {
	return map[string]any{
		"final_status": "INCONCLUSIVE",
		"stages": []any{
			map[string]any{"stage": "RGO", "cycle_id": "c2", "finding_id": "f2", "output_hash": "sha256:r2", "data": map[string]any{"ok": true}},
		},
	}, nil
}

func TestRegisterTrinityOperationBlocksUnvalidatedResult(t *testing.T) {
	e := newTestEngine(t)
	sink := &fakeHortaSink{}
	if err := RegisterTrinityOperation(e, inconclusiveTrinitySource{}, sink); err != nil {
		t.Fatal(err)
	}
	result, err := e.Execute(context.Background(), "rgo.trinity.process@1.0.0", []float64{0}, map[string]string{
		"rgo_envelope_json": `{"schema_version":"1.0.0"}`,
	})
	if err == nil || result.Status != "blocked" {
		t.Fatalf("expected blocked unvalidated result: %#v err=%v", result, err)
	}
}

type multiStageTrinitySource struct{}

func (multiStageTrinitySource) RGOTrinity(_ context.Context, _ map[string]any, _ string) (map[string]any, error) {
	return map[string]any{
		"final_status": "VALIDATED",
		"stages": []any{
			map[string]any{
				"stage":                   "ARA",
				"cycle_id":                "gold-cycle",
				"finding_id":              "finding-1",
				"output_hash":             "sha256:out-1",
				"eru_snapshot_hash":       "sha256:eru-1",
				"rgo_evidence_chain_hash": "sha256:rgo-1",
			},
			map[string]any{
				"stage":                   "ETR",
				"cycle_id":                "gold-cycle",
				"finding_id":              "finding-2",
				"output_hash":             "sha256:out-2",
			},
			map[string]any{
				"stage":                   "ITR",
				"cycle_id":                "gold-cycle",
				"finding_id":              "finding-3",
				"output_hash":             "sha256:out-3",
			},
		},
	}, nil
}

type failFirstHortaSink struct{ calls int }

func (f *failFirstHortaSink) CallWithCorrelation(_ context.Context, nucleus, capability string, _ map[string]any, _ string) (map[string]any, error) {
	if nucleus != "N01" || capability != "rgo.hortacore.store" {
		return nil, errors.New("unexpected Horta route")
	}
	f.calls++
	if f.calls == 1 {
		return nil, errors.New("transient Horta failure")
	}
	return map[string]any{"persisted": true}, nil
}

func TestRegisterTrinityOperationIsolatesHortaFailureAndContinuesStages(t *testing.T) {
	e := newTestEngine(t)
	sink := &failFirstHortaSink{}
	if err := RegisterTrinityOperation(e, multiStageTrinitySource{}, sink); err != nil {
		t.Fatal(err)
	}

	result, err := e.Execute(context.Background(), "rgo.trinity.process@1.0.0", []float64{0}, map[string]string{
		"rgo_envelope_json": `{"schema_version":"1.0.0"}`,
	})
	if err == nil || result.Status != "blocked" {
		t.Fatalf("expected partial Horta block: %#v err=%v", result, err)
	}
	if sink.calls != 3 {
		t.Fatalf("expected all stages to be attempted after first failure, got %d calls", sink.calls)
	}

	var out map[string]any
	if decodeErr := json.Unmarshal([]byte(result.Metadata["rgo_trinity_result_json"]), &out); decodeErr != nil {
		t.Fatalf("invalid RGO result metadata: %v", decodeErr)
	}
	if persisted, _ := out["hortacore_persisted"].(bool); persisted {
		t.Fatalf("partial persistence must not report all stages persisted: %#v", out)
	}

	results, ok := out["hortacore_results"].([]any)
	if !ok || len(results) != 3 {
		t.Fatalf("expected three per-stage outcomes: %#v", out["hortacore_results"])
	}

	first, _ := results[0].(map[string]any)
	if first["status"] != "BLOCKED_HORTA" ||
		first["cycle_id"] != "gold-cycle" ||
		first["finding_id"] != "finding-1" ||
		first["output_hash"] != "sha256:out-1" ||
		first["eru_snapshot_hash"] != "sha256:eru-1" ||
		first["rgo_evidence_chain_hash"] != "sha256:rgo-1" ||
		first["error"] != "transient Horta failure" {
		t.Fatalf("first failure did not preserve Rule-of-Gold evidence: %#v", first)
	}

	second, _ := results[1].(map[string]any)
	if persisted, _ := second["persisted"].(bool); !persisted {
		t.Fatalf("second stage was not allowed to recover after first-stage failure: %#v", second)
	}
	third, _ := results[2].(map[string]any)
	if persisted, _ := third["persisted"].(bool); !persisted {
		t.Fatalf("third stage was not allowed to recover after first-stage failure: %#v", third)
	}
}
