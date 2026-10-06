package grce

import (
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/grf"
)

func TestExternalParticipantsAreFailClosedWhenRuntimeIsUnconfigured(t *testing.T) {
	for id, p := range NewExternalParticipants() {
		res, err := p.Ingest(grf.State{Payload: []byte("preserve"), Epistemic: grf.EpistemicActive}, grf.Context{CycleID: "cycle-external"})
		if err == nil || !strings.Contains(err.Error(), "EXTERNAL_RUNTIME_BLOCKED:"+id) {
			t.Fatalf("%s: expected blocked runtime, got %v", id, err)
		}
		if string(res.Output.Payload) != "preserve" {
			t.Fatalf("%s: input payload was not preserved", id)
		}
		if res.Output.Epistemic != grf.EpistemicProjected {
			t.Fatalf("%s: expected PROJECTED, got %s", id, res.Output.Epistemic)
		}
		if res.Evidence.Hash == "" || res.Provenance.ParentHash == "" {
			t.Fatalf("%s: missing evidence/provenance", id)
		}
	}
}

func TestExternalResponseContractRequiresPinnedIdentityAndEvidence(t *testing.T) {
	p := NewExternalParticipants()[BijuxDAGRuntimeID]
	input := grf.State{Payload: []byte("input"), Epistemic: grf.EpistemicActive}
	ctx := grf.Context{CycleID: "cycle-contract"}
	e := grf.Evidence{ID: "evidence-1", FailureID: "failure-1", ContextHash: ctx.Hash(), Source: BijuxDAGRuntimeID, Detail: "observed"}
	e.Hash = hashExternalEvidence(e)
	resp := externalResponse{
		StateB64:   "aW5wdXQ=",
		Provenance: grf.Provenance{ParentHash: input.Hash(), InputHash: input.Hash(), OutputHash: "output", SequenceIndex: 1, Chain: []string{"external"}},
		Evidence:   e,
		Status:     grf.EpistemicProjected,
		Source:     p.Config.Source,
		Revision:   p.Config.Revision,
	}
	if err := p.validateResponse(input, ctx, resp); err != nil {
		t.Fatalf("valid response rejected: %v", err)
	}

	resp.Source = "attacker/source"
	if err := p.validateResponse(input, ctx, resp); err == nil || !strings.Contains(err.Error(), "EXTERNAL_SOURCE_MISMATCH") {
		t.Fatalf("expected pinned source rejection, got %v", err)
	}
}
