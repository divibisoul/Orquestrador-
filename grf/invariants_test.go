package grf

import (
	"strings"
	"testing"
)

func TestCanonicalInvariantSetIsComplete(t *testing.T) {
	if err := ValidateInvariantSet(CanonicalInvariants); err != nil {
		t.Fatal(err)
	}
	if len(CanonicalInvariants) != 18 {
		t.Fatalf("expected 18 invariants, got %d", len(CanonicalInvariants))
	}
}

func TestMonotonicityRejectsContentLoss(t *testing.T) {
	before := State{Payload: []byte("0123456789"), Epistemic: EpistemicActive}
	after := State{Payload: []byte("0123"), Epistemic: EpistemicProjected}
	ctx := Context{CycleID: "cycle-test"}
	p := Provenance{ParentHash: before.Hash(), InputHash: before.Hash(), OutputHash: after.Hash(), SequenceIndex: 1}
	err := ValidateContextAndProvenance(before, ctx, after, p)
	if err == nil || !strings.Contains(err.Error(), "MONOTONICITY") {
		t.Fatalf("expected monotonicity error, got %v", err)
	}
}

func TestContextRequiresCycleID(t *testing.T) {
	if err := (Context{}).Validate(); err == nil {
		t.Fatal("expected missing cycle id error")
	}
}
