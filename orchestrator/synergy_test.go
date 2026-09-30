package orchestrator

import (
	"context"
	"strings"
	"testing"
)

type recordingSynergyInvoker struct{ calls []string }

func (r *recordingSynergyInvoker) Invoke(_ context.Context, target, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	r.calls = append(r.calls, target+":"+capability+":"+correlation)
	return map[string]any{"target": target, "accepted": payload["synergy_sequence_id"]}, nil
}

func TestFavoriteSynergySequencePreservesCanonicalTopology(t *testing.T) {
	sequence := FavoriteSynergySequence()
	if err := ValidateSynergySequence(sequence); err != nil {
		t.Fatal(err)
	}
	if sequence.Nodes[0].Target != "N01" || sequence.Nodes[1].Target != "N05" ||
		sequence.Nodes[2].Target != "SARA" || sequence.Nodes[3].Target != "N02" ||
		sequence.Nodes[4].Target != "N03" || sequence.Nodes[5].Target != "N06" ||
		sequence.Nodes[6].Target != "N04" || sequence.Nodes[7].Target != "N07" {
		t.Fatalf("favorite sequence changed: %+v", sequence.Nodes)
	}
	if len(sequence.Edges) != 7 {
		t.Fatalf("expected 7 composition edges, got %d", len(sequence.Edges))
	}
}

func TestExecuteSynergyRouteUsesSharedCorrelationAndFinalizesAtN07(t *testing.T) {
	invoker := &recordingSynergyInvoker{}
	capabilities := map[string]string{"N01":"mesh.describe","N05":"mesh.describe","SARA":"sara.state","N02":"mesh.describe","N03":"mesh.describe","N06":"mesh.describe","N04":"mesh.describe"}
	result, err := ExecuteSynergyRoute(context.Background(), FavoriteSynergySequence(), invoker, "synergy-correlation", capabilities, map[string]any{"text":"probe","prompt":"probe","input":"probe","audioBase64":"ZmFrZS1hdWRpbw==","mimeType":"audio/wav","tool":"getWeather"})
	if err != nil { t.Fatal(err) }
	if result.FinalTarget != "N07" || result.Status != "ok" { t.Fatalf("unexpected synergy result: %+v", result) }
	if len(result.Trace) != 8 || len(invoker.calls) != 7 { t.Fatalf("expected 7 remote stages plus local N07 finalization: trace=%d calls=%d", len(result.Trace), len(invoker.calls)) }
	for _, call := range invoker.calls {
		if !strings.HasSuffix(call, "synergy-correlation") { t.Fatalf("correlation not preserved: %s", call) }
	}
}

func TestExecuteSynergyRouteBlocksMissingStagePrerequisite(t *testing.T) {
	_, err := ExecuteSynergyRoute(context.Background(), FavoriteSynergySequence(), &recordingSynergyInvoker{}, "missing-input-correlation", map[string]string{}, map[string]any{"text":"probe"})
	if err == nil || !strings.Contains(err.Error(), "required field prompt") {
		t.Fatalf("unexpected prerequisite result: %v", err)
	}
}
