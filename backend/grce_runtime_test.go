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
			var auditRequest map[string]any
			if err := json.NewDecoder(r.Body).Decode(&auditRequest); err != nil {
				t.Errorf("decode ARA request: %v", err)
				writeGRCEJSONTest(w, map[string]any{"error": "invalid ARA request"})
				return
			}
			input, _ := auditRequest["input"].(string)
			findingKind := "TEST_EVIDENCE"
			findingDetail := "free-text mention of COMPLEXIDADE_EXCESSIVA is not itself a typed ARA finding"
			if len(input) > 10000 {
				findingKind = "COMPLEXIDADE_EXCESSIVA"
				findingDetail = "ARA classified the original high-complexity input"
			}
			writeGRCEJSONTest(w, map[string]any{
				"flaws": []any{map[string]any{
					"kind":   findingKind,
					"detail": findingDetail,
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
		case "/v1/rgo/trinity":
			// Assert the boundary validates the hash-linked candidate envelope,
			// not a second copy of the entire original input. Original state must
			// remain traceable through hashes and artifact evidence references.
			var request map[string]any
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			finding, ok := request["finding"].(map[string]any)
			if !ok {
				t.Fatalf("RGO finding missing: %#v", request)
			}
			failure, ok := finding["failure"].(map[string]any)
			if !ok {
				t.Fatalf("RGO failure envelope missing: %#v", finding)
			}
			description, _ := failure["description"].(string)
			if len(description) == 0 || len(description) > 1024 {
				t.Fatalf("RGO description must be compact, got %d bytes", len(description))
			}
			if !strings.Contains(description, "original_state_hash=") {
				t.Fatalf("RGO description does not bind the original state hash: %q", description)
			}
			extensions, ok := finding["extensions"].(map[string]any)
			if !ok || extensions["original_state_hash"] == "" || extensions["candidate_hash"] == "" {
				t.Fatalf("RGO extensions must preserve original and candidate hashes: %#v", finding["extensions"])
			}
			evidence, ok := finding["evidence"].([]any)
			if !ok || len(evidence) == 0 {
				t.Fatalf("RGO evidence references missing: %#v", finding["evidence"])
			}
			// Contract fixture mirrors the evidence fields returned by SARA's
			// existing RGO/Trinity processor. Runtime validation still requires
			// actual TRINITY, ERU, and MMD stage names; it must not infer them.
			writeGRCEJSONTest(w, map[string]any{
				"final_status": "VALIDATED",
				"finding_id": "grce:corr-test",
				"stages": []any{
					map[string]any{"stage": "RGO", "finding_id": "grce:corr-test", "rgo_evidence_chain_hash": "sha256:rgo", "output_hash": "sha256:rgo-output"},
					map[string]any{"stage": "TRINITY::ARA", "finding_id": "grce:corr-test", "rgo_evidence_chain_hash": "sha256:trinity", "output_hash": "sha256:trinity-output"},
					map[string]any{"stage": "ERU", "finding_id": "grce:corr-test", "rgo_evidence_chain_hash": "sha256:eru", "output_hash": "sha256:eru-output"},
					map[string]any{"stage": "MMD", "finding_id": "grce:corr-test", "rgo_evidence_chain_hash": "sha256:mmd", "output_hash": "sha256:mmd-output"},
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

	// The detector category, not free-text prose, controls the safety transition.
	// The benign fixture above deliberately mentions the marker in detail while
	// still reaching ACTIVE because its structured kind is TEST_EVIDENCE.
	complexInput := strings.Repeat("a", 11000)
	preserved, preserveErr := executor.Run(context.Background(), grf.State{
		ID:              "state-complex",
		EpistemicState: grf.REAL,
		Payload:         map[string]any{"text": complexInput},
	}, grf.Context{
		TraceID:       "trace-complex",
		CorrelationID: "corr-complex",
		SequenceIndex: 2,
	})
	if preserveErr == nil || !strings.Contains(preserveErr.Error(), "GRCE_COMPLEXITY_EXCESSIVE_INPUT_PRESERVED") {
		t.Fatalf("expected typed ARA complexity finding to fail closed, got result=%#v err=%v", preserved, preserveErr)
	}
	if preserved.EpistemicState != grf.PRESERVED || preserved.Outcome != "ROLLBACK_PRESERVED:DETECT" {
		t.Fatalf("complex input state=%s outcome=%q, want PRESERVED/ROLLBACK_PRESERVED:DETECT", preserved.EpistemicState, preserved.Outcome)
	}
	gotText, ok := preserved.NextState.Payload["text"].(string)
	if !ok || gotText != complexInput {
		t.Fatal("original high-complexity input was not preserved byte-for-byte")
	}
	if len(preserved.Evidence) == 0 {
		t.Fatal("typed complexity finding must retain evidence and causal identity")
	}
}


func TestGRCEEthicalReviewInputPreservesContentAndStatesPrinciples(t *testing.T) {
	const source = "a\na\na\nCOMPLEXIDADE_EXCESSIVA; original trace sha256:abc123"
	review := grceEthicalReviewInput("validar o estado candidato antes da ativação", source)
	for _, required := range []string{
		"autonomia",
		"transparência",
		"responsabilidade",
		"cuidado com a comunidade",
		"histórico causal",
		"proveniência",
		source,
	} {
		if !strings.Contains(review, required) {
			t.Fatalf("ethical review context lost required principle/source %q", required)
		}
	}
}

func writeGRCEJSONTest(w http.ResponseWriter, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func TestBuildGRCERGOFindingMatchesSARAEnvelopeSchema(t *testing.T) {
	const originalText = "preservar 🔒IDENTITY, evidências e proveniência"
	const transformedText = "candidate-transformed-output-preserving-history"
	state := grf.State{
		ID:              "grce-candidate-001",
		EpistemicState: grf.PRESERVED,
		Payload: map[string]any{
			"original_state": map[string]any{"text": originalText},
			"artifacts": []grf.Artifact{{
				ID:    "candidate-artifact-001",
				Stage: "ITR",
				State: grf.REAL,
				Payload: map[string]any{
					"transformed": transformedText,
				},
			}},
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
	if !ok {
		t.Fatalf("RGO failure record is missing: %#v", finding["failure"])
	}
	description, ok := failure["description"].(string)
	if !ok {
		t.Fatalf("RGO candidate description must remain a complete serialized string: %T", failure["description"])
	}
	var reviewed map[string]any
	if err := json.Unmarshal([]byte(description), &reviewed); err != nil {
		t.Fatalf("RGO candidate representation is not valid JSON: %v", err)
	}
	if reviewed["candidate_id"] != state.ID || reviewed["candidate_epistemic_state"] != string(grf.PRESERVED) {
		t.Fatalf("candidate identity/epistemic state was lost: %#v", reviewed)
	}
	candidatePayload, ok := reviewed["source_and_transformations"].(map[string]any)
	if !ok {
		t.Fatalf("original state and transformations were dropped from RGO payload: %#v", reviewed["source_and_transformations"])
	}
	original, ok := candidatePayload["original_state"].(map[string]any)
	if !ok || original["text"] != originalText {
		t.Fatalf("original source content/protected tag was not preserved: %#v", candidatePayload["original_state"])
	}
	artifacts, ok := candidatePayload["artifacts"].([]any)
	if !ok || len(artifacts) != 1 {
		t.Fatalf("candidate transformation artifacts were not preserved: %#v", candidatePayload["artifacts"])
	}
	artifact, ok := artifacts[0].(map[string]any)
	if !ok || artifact["id"] != "candidate-artifact-001" {
		t.Fatalf("candidate artifact identity was lost: %#v", artifacts[0])
	}
	artifactPayload, ok := artifact["payload"].(map[string]any)
	if !ok || artifactPayload["transformed"] != transformedText {
		t.Fatalf("transformed candidate content was lost: %#v", artifact["payload"])
	}
}

