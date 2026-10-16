package grce

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/grf"
)

func TestGRCERejectsIncompleteProvenance(t *testing.T) {
	err:=validateProvenanceChain("parent",[]grf.Provenance{{
		ParentHash:"parent",
		InputHash:"input",
		OutputHash:"",
		SequenceIndex:1,
		Stage:"DETECT",
	}},grf.State{ParentHash:"parent"})
	if err==nil || err.Error()!="GRCE_PROVENANCE_INCOMPLETE" {
		t.Fatalf("unexpected error: %v",err)
	}
}

func TestGRCECanonicalInvariantsAreEighteen(t *testing.T) {
	set:=grf.CanonicalInvariantSet()
	if err:=grf.ValidateInvariantSet(set);err!=nil{t.Fatal(err)}
	if len(set.Invariants)!=18{t.Fatalf("invariants=%d",len(set.Invariants))}
}
