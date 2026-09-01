package tests

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

func TestSuperAGI(t *testing.T) {
	a := superagi.NewWithProvider(testProvider{})
	v := a.GenerateEmbedding("hello")
	if len(v) != 4 { t.Fatalf("embedding size=%d", len(v)) }
	a.SemanticMemory("earth", "planet")
	if !a.VerifyFact("earth") { t.Fatal("semantic memory") }
	if !a.VerifySafety("normal output") { t.Fatal("safety false positive") }
	a.WorkingMemory("a", "b")
	if len(a.ReplayExperience(1)) != 0 { t.Fatal("unexpected replay") }
}
