package grce

import (
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/grf"
)

func TestExternalParticipantsAreFailClosedWhenRuntimeIsUnconfigured(t *testing.T) {
	for id, p := range NewExternalParticipants() {
		input := grf.State{ID: "state-1", Payload: map[string]any{"value": "preserve"}, EpistemicState: grf.ACTIVE}
		ctx := grf.Context{TraceID: "trace-external", CorrelationID: "corr-external", SequenceIndex: 1}
		out, prov, ev := p.Ingest(input, ctx)
		if ev.State != grf.BLOCKED {
			t.Fatalf("%s: expected BLOCKED evidence, got %s", id, ev.State)
		}
		if out.EpistemicState != grf.PROJECTED {
			t.Fatalf("%s: expected PROJECTED output, got %s", id, out.EpistemicState)
		}
		if out.Payload["value"] != "preserve" {
			t.Fatalf("%s: input payload was not preserved", id)
		}
		if prov.ParentHash == "" || prov.OutputHash == "" || ev.Hash == "" {
			t.Fatalf("%s: missing evidence/provenance", id)
		}
	}
}

func TestExternalResponseContractRequiresPinnedIdentityAndEvidence(t *testing.T) {
	p := NewExternalParticipants()[BijuxDAGRuntimeID]
	input := grf.State{ID: "state-1", Payload: map[string]any{"value": "input"}, EpistemicState: grf.ACTIVE}
	ctx := grf.Context{TraceID: "trace-contract", CorrelationID: "corr-contract", SequenceIndex: 1}
	out := input
	out.EpistemicState = grf.PROJECTED
	out.ParentHash = mustStateHash(out)
	prov, err := grf.SealProvenance(mustStateHash(input), input, out, 2, "EXTERNAL_TEST", "test")
	if err != nil { t.Fatal(err) }
	resp := externalResponse{
		State: out, Provenance: prov,
		Evidence: grf.Evidence{ID: "evidence-1", FailureID: "failure-1", State: grf.PROJECTED, Hash: prov.OutputHash, InputHash: prov.InputHash, OutputHash: prov.OutputHash, SequenceIndex: 2},
		Status: grf.PROJECTED, Source: p.Config.Source, Revision: p.Config.Revision,
	}
	if err := p.validateResponse(input, ctx, resp); err != nil {
		t.Fatalf("valid response rejected: %v", err)
	}
	resp.Source = "attacker/source"
	if err := p.validateResponse(input, ctx, resp); err == nil || !strings.Contains(err.Error(), "EXTERNAL_SOURCE_MISMATCH") {
		t.Fatalf("expected pinned source rejection, got %v", err)
	}
}

func TestExternalRuntimeFailurePreservesEvidence(t *testing.T) {
	p := NewExternalParticipants()[BijuxDAGRuntimeID]
	t.Setenv(p.Config.CommandEnv, "/definitely/not/a/real/soul-runtime")
	input := grf.State{ID: "state-1", Payload: map[string]any{"value": "preserve"}, EpistemicState: grf.ACTIVE}
	ctx := grf.Context{TraceID: "trace-runtime", CorrelationID: "corr-runtime", SequenceIndex: 1}
	out, prov, ev := p.Ingest(input, ctx)
	if ev.State != grf.BLOCKED {
		t.Fatalf("expected blocked evidence, got %s", ev.State)
	}
	if prov.ParentHash == "" || prov.OutputHash == "" || out.Payload["value"] != "preserve" {
		t.Fatal("runtime failure did not preserve state/provenance")
	}
}

func mustStateHash(s grf.State) string {
	h, err := s.Hash()
	if err != nil { panic(err) }
	return h
}
