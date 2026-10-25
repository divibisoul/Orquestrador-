package backend

import (
	"time"
	"fmt"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/grce"
	"github.com/divibisoul/Orquestrador-/grf"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestGRCEExecutorRuntimeCompletesWithHTTPBoundaryFixture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/state":
			writeGRCEJSONTest(w, map[string]any{"ready": true, "source": "SARA"})
		case "/v1/audit":
			writeGRCEJSONTest(w, map[string]any{
				"flaws": []any{map[string]any{
					"kind":   "COMPLEXIDADE_EXCESSIVA",
					"detail": "test evidence from ARA",
				}},
				"ethical": map[string]any{"approved": true},
			})
		case "/v1/regenerate":
			writeGRCEJSONTest(w, map[string]any{
				"original":    "preservar estado",
				"transformed": "preservar estado\n[ARA: evidence preserved]",
			})
		case "/v1/cycle":
			writeGRCEJSONTest(w, map[string]any{
				"converged":              true,
				"trace_hash":             "sha256:test-trace",
				"federated_context_hash": "sha256:test-context",
			})
		case "/v1/grce/hooks":
			writeGRCEJSONTest(w, map[string]any{
				"final_status": "VALIDATED",
				"finding_id": "grce:corr-test",
				"all_hooks_observed": true,
				"hook_evidence": map[string]any{
					"RGO": true, "ARA": true, "ITR": true, "ETR": true, "ERU": true, "MMD": true,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	proxy := NewSARAProxy(Config{
		SARAServiceURL:   server.URL,
		SARAServiceToken: "test-token",
	})
	compute := supergpu.New(nil)
	compute.Discover()

	executor, err := NewGRCEExecutorRuntime(proxy, compute, GRCEFeedback{
		Horta: func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error { return nil },
		Vagus: func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error { return nil },
		Mesh:  func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := executor.Run(context.Background(), grf.State{
		ID:              "state-test",
		EpistemicState: grf.PRESERVED,
		Payload:         map[string]any{"text": "preservar estado"},
	}, grf.Context{
		TraceID:       "trace-test",
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

func writeGRCEJSONTest(w http.ResponseWriter, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}


func TestGRCEEthicalReviewInputPreservesFullContent(t *testing.T) {
	content := "simulação de auditoria com modificar 🔒IDENTITY; manter o original intacto"
	got := grceEthicalReviewInput("avaliar a necessidade e os limites da transformação", content)
	for _, required := range []string{
		"Princípios do ciclo GRCE",
		"autonomia",
		"transparência",
		"responsabilidade",
		"cuidado com a comunidade",
		"histórico causal",
		"proveniência",
		"Finalidade da avaliação: avaliar a necessidade e os limites da transformação",
		content,
	} {
		if !strings.Contains(got, required) {
			t.Fatalf("ethical review input lost required context %q: %q", required, got)
		}
	}
}

func TestBuildGRCERGOFindingMatchesSARAEnvelopeSchema(t *testing.T) {
	state := grf.State{
		ID:              "grce-candidate-001",
		EpistemicState: grf.PRESERVED,
		Payload: map[string]any{
			"original_state": map[string]any{"text": "preservar 🔒IDENTITY, evidências e proveniência"},
			"artifacts":      []grf.Artifact{},
		},
	}
	context := grf.Context{TraceID: "trace-rgo-001", CorrelationID: "corr-rgo-001", SequenceIndex: 7}
	finding, err := buildGRCERGOFinding(state, context)
	if err != nil {
		t.Fatal(err)
	}
	requiredStrings := map[string]string{
		"schema_version": "1.0.0",
		"finding_id":     "grce:corr-rgo-001",
		"object_id":      state.ID,
		"correlation_id": context.CorrelationID,
		"trace_id":       context.TraceID,
	}
	for key, want := range requiredStrings {
		if got, ok := finding[key].(string); !ok || got != want {
			t.Fatalf("%s=%v, want %q", key, finding[key], want)
		}
	}
	timestamp, ok := finding["timestamp"].(string)
	if !ok {
		t.Fatalf("timestamp missing or non-string: %T", finding["timestamp"])
	}
	if _, err := time.Parse(time.RFC3339Nano, timestamp); err != nil {
		t.Fatalf("timestamp does not satisfy RFC3339Nano: %q: %v", timestamp, err)
	}
	source, ok := finding["source"].(map[string]any)
	if !ok || source["system"] != "SOUL" || source["module"] != "N07.GRCE" || source["version"] != grce.Version {
		t.Fatalf("invalid SARA source contract: %#v", finding["source"])
	}
	epistemic, ok := finding["epistemic"].(map[string]any)
	if !ok || epistemic["mode"] != "EXECUTION" || epistemic["verification_state"] != "UNVERIFIED" {
		t.Fatalf("invalid or overstated epistemic state: %#v", finding["epistemic"])
	}
	actionability, ok := finding["actionability"].(map[string]any)
	if !ok || actionability["status"] != "ACTIONABLE" {
		t.Fatalf("invalid actionability contract: %#v", finding["actionability"])
	}
	dual, ok := finding["dual"].(map[string]any)
	if !ok || dual["status"] != "UNRESOLVED" {
		t.Fatalf("invalid dual contract: %#v", finding["dual"])
	}
	evidence, ok := finding["evidence"].([]any)
	if !ok || len(evidence) != 1 {
		t.Fatalf("expected one state-hash evidence reference, got %#v", finding["evidence"])
	}
	ref, ok := evidence[0].(map[string]any)
	if !ok || ref["id"] == "" || ref["kind"] != "GRCE_STATE_HASH" || !strings.HasPrefix(fmt.Sprint(ref["ref"]), "sha256:") {
		t.Fatalf("invalid state-hash evidence reference: %#v", evidence[0])
	}
	provenance, ok := finding["provenance"].(map[string]any)
	if !ok || provenance["origin"] != "N07.GRCE" || provenance["input_hash"] == "" {
		t.Fatalf("invalid GRCE provenance: %#v", finding["provenance"])
	}
	failure, ok := finding["failure"].(map[string]any)
	if !ok || !strings.Contains(fmt.Sprint(failure["description"]), "🔒IDENTITY") {
		t.Fatalf("RGO payload did not preserve the original source content: %#v", finding["failure"])
	}
}
