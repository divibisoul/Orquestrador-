package orchestrator_test

import (
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/grf"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestGRCEExecutorRuntimeBindsRealBoundariesAndCompletes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/state":
			writeJSONTest(w, map[string]any{"ready": true, "source": "SARA"})
		case "/v1/audit":
			writeJSONTest(w, map[string]any{
				"flaws": []any{map[string]any{
					"kind": "COMPLEXIDADE_EXCESSIVA",
					"detail": "test evidence from ARA",
				}},
				"ethical": map[string]any{"approved": true},
			})
		case "/v1/regenerate":
			writeJSONTest(w, map[string]any{
				"original": "preservar estado",
				"transformed": "preservar estado
[ARA: evidence preserved]",
			})
		case "/v1/cycle":
			writeJSONTest(w, map[string]any{
				"converged": true,
				"trace_hash": "sha256:test-trace",
				"federated_context_hash": "sha256:test-context",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	proxy := backend.NewSARAProxy(backend.Config{
		SARAServiceURL: server.URL,
		SARAServiceToken: "test-token",
	})
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	compute := supergpu.New(nil)
	compute.Discover()

	executor, err := orchestrator.NewGRCEExecutorRuntime(proxy, compute, peers, orchestrator.GRCEFeedback{
		Horta: func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error { return nil },
		Vagus: func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error { return nil },
		Mesh:  func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := executor.Run(context.Background(), grf.State{
		ID: "state-test",
		EpistemicState: grf.PRESERVED,
		Payload: map[string]any{"text": "preservar estado"},
	}, grf.Context{
		TraceID: "trace-test",
		CorrelationID: "corr-test",
		SequenceIndex: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.EpistemicState != grf.ACTIVE {
		t.Fatalf("epistemic state=%s, want ACTIVE", result.EpistemicState)
	}
	if result.Outcome != "ACTIVE" {
		t.Fatalf("outcome=%q, want ACTIVE", result.Outcome)
	}
	if len(result.Provenance) < 5 {
		t.Fatalf("provenance entries=%d, want >=5", len(result.Provenance))
	}
	for i := 1; i < len(result.Provenance); i++ {
		if result.Provenance[i].SequenceIndex <= result.Provenance[i-1].SequenceIndex {
			t.Fatalf("provenance sequence not strict at %d: %d <= %d", i, result.Provenance[i].SequenceIndex, result.Provenance[i-1].SequenceIndex)
		}
	}
	if len(result.Evidence) == 0 || len(result.Capabilities) == 0 {
		t.Fatalf("expected evidence and capabilities, got evidence=%d capabilities=%d", len(result.Evidence), len(result.Capabilities))
	}
}

func writeJSONTest(w http.ResponseWriter, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
