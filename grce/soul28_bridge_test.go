package grce

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/grf"
)

func TestSOUL28RuntimeSubprocessHelper(t *testing.T) {
	if os.Getenv("SOUL28_TEST_HELPER") != "1" {
		return
	}
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(21)
	}
	var request soul28ExternalRequest
	if err := json.Unmarshal(raw, &request); err != nil {
		os.Exit(22)
	}
	payload, err := base64.StdEncoding.DecodeString(request.State)
	if err != nil {
		os.Exit(23)
	}
	input := request.StateContract.Clone()
	input.Payload = payload
	status := strings.TrimSpace(os.Getenv("SOUL28_TEST_STATUS"))
	if status == "" {
		status = string(grf.PROJECTED)
	}
	output := input.Clone()
	output.Epistemic = status
	evidence := grf.SOUL28Evidence{
		ID:          "evidence-test-runtime",
		FailureID:   "failure-test-runtime",
		ContextHash: request.Context.Hash(),
		Source:      request.ParticipantID,
		Detail:      "subprocess contract test",
	}
	evidence.Hash = hashSOUL28ExternalEvidence(evidence)
	response := soul28ExternalResponse{
		StateB64: base64.StdEncoding.EncodeToString(output.Payload),
		Provenance: grf.SOUL28Provenance{
			ParentHash:    input.Hash(),
			InputHash:     input.Hash(),
			OutputHash:   output.Hash(),
			SequenceIndex: input.SequenceIndex + 1,
			Chain:         []string{"test-runtime", "projected-candidate"},
		},
		Evidence: evidence,
		Status:   status,
		Source:   request.Source,
		Revision: request.Revision,
	}
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		os.Exit(24)
	}
	os.Exit(0)
}

func TestSOUL28BridgeInvokesPinnedRunnerAndNeverPromotesExternalREAL(t *testing.T) {
	bridge := NewSOUL28ParticipantBridge()
	config := bridge.configs[SOUL28BijuxID]
	t.Setenv(config.CommandEnv, os.Args[0])
	t.Setenv(config.ArgsEnv, "-test.run=TestSOUL28RuntimeSubprocessHelper")
	t.Setenv("SOUL28_TEST_HELPER", "1")
	t.Setenv("SOUL28_TEST_STATUS", string(grf.REAL))

	input := grf.State{
		ID:             "input-state",
		EpistemicState: grf.ACTIVE,
		Payload:        map[string]any{"text": "preserve-me", "counter": 3},
	}
	cycle := grf.Context{
		TraceID: "trace-soul28-test",
		CorrelationID: "correlation-soul28-test",
		SequenceIndex: 9,
		Values: map[string]any{"intent": "bounded-test"},
	}
	output, provenance, evidence, err := bridge.InvokeParticipant(testContext(), SOUL28BijuxID, input, cycle)
	if err != nil {
		t.Fatal(err)
	}
	if output.EpistemicState != grf.PROJECTED {
		t.Fatalf("external REAL must remain a projected N07 candidate, got %s", output.EpistemicState)
	}
	if output.Payload["text"] != "preserve-me" {
		t.Fatalf("payload was not carried across the bridge: %#v", output.Payload)
	}
	if provenance.Stage != "SOUL28_PARTICIPANT:"+SOUL28BijuxID || provenance.SequenceIndex != cycle.SequenceIndex+1 {
		t.Fatalf("unexpected canonical provenance: %#v", provenance)
	}
	if evidence.State != grf.PROJECTED || evidence.Payload["external_state"] != string(grf.REAL) {
		t.Fatalf("external evidence was not retained without promotion: %#v", evidence)
	}
	if evidence.Payload["pinned_revision"] != config.Revision {
		t.Fatalf("revision pin was not preserved: %#v", evidence.Payload)
	}
}

func TestSOUL28BridgeFailsClosedAndPreservesInputWhenRunnerIsMissing(t *testing.T) {
	bridge := NewSOUL28ParticipantBridge()
	for _, config := range bridge.configs {
		t.Setenv(config.CommandEnv, "")
	}
	input := grf.State{ID: "input", EpistemicState: grf.ACTIVE, Payload: map[string]any{"v": "original"}}
	cycle := grf.Context{TraceID: "trace", CorrelationID: "correlation", SequenceIndex: 2}
	output, provenance, evidence, err := bridge.Run(testContext(), input, cycle)
	if err == nil || !strings.Contains(err.Error(), "SOUL28_RUNTIME_BLOCKED:"+SOUL28BijuxID+":runtime-unconfigured") {
		t.Fatalf("expected fail-closed runtime error, got %v", err)
	}
	inputHash, _ := input.Hash()
	outputHash, _ := output.Hash()
	if outputHash != inputHash {
		t.Fatal("the original canonical state was not preserved")
	}
	if len(provenance) != 1 || provenance[0].Stage != "SOUL28_PARTICIPANT_BLOCKED:"+SOUL28BijuxID {
		t.Fatalf("blocked provenance missing: %#v", provenance)
	}
	if len(evidence) != 1 || evidence[0].State != grf.BLOCKED || evidence[0].Hash == "" {
		t.Fatalf("blocked evidence missing: %#v", evidence)
	}
}

func TestSOUL28ResponseRejectsUnpinnedSourceAndRevision(t *testing.T) {
	bridge := NewSOUL28ParticipantBridge()
	config := bridge.configs[SOUL28BijuxID]
	state := grf.State{ID: "input", EpistemicState: grf.ACTIVE, Payload: map[string]any{"v": "x"}}
	cycle := grf.Context{TraceID: "trace", CorrelationID: "corr", SequenceIndex: 3}
	legacyInput, err := grf.SOUL28StateFromCanonical(state, cycle.SequenceIndex, 1)
	if err != nil {
		t.Fatal(err)
	}
	legacyContext, err := soul28ContextFromCanonical(cycle, legacyInput.ParentHash)
	if err != nil {
		t.Fatal(err)
	}
	output := legacyInput.Clone()
	output.Epistemic = string(grf.PROJECTED)
	evidence := grf.SOUL28Evidence{ID: "evidence", FailureID: "failure", ContextHash: legacyContext.Hash(), Source: config.ID, Detail: "test"}
	evidence.Hash = hashSOUL28ExternalEvidence(evidence)
	response := soul28ExternalResponse{
		StateB64: base64.StdEncoding.EncodeToString(output.Payload),
		Provenance: grf.SOUL28Provenance{
			ParentHash: legacyInput.Hash(), InputHash: legacyInput.Hash(),
			OutputHash: output.Hash(), SequenceIndex: legacyInput.SequenceIndex + 1,
			Chain: []string{"test"},
		},
		Evidence: evidence, Status: string(grf.PROJECTED),
		Source: config.Source, Revision: config.Revision,
	}
	response.Revision = strings.Repeat("0", 40)
	if err := validateSOUL28Response(config, legacyInput, legacyContext, response); err == nil || !strings.Contains(err.Error(), "REVISION_MISMATCH") {
		t.Fatalf("expected pinned revision rejection, got %v", err)
	}
}

func testContext() context.Context {
	return context.Background()
}
