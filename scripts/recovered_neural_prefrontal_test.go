package main

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
)

func TestRecoveredNeuralAndPrefrontalRuntime(t *testing.T) {
	n, err := neural.New(4, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	if err := n.AddEdge(0, 1, 0.2); err != nil {
		t.Fatal(err)
	}
	if err := n.AddEdge(1, 2, 0.15); err != nil {
		t.Fatal(err)
	}
	out, err := n.Forward(context.Background(), []float64{0.2, 0.1, 0.0, 0.3})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 4 {
		t.Fatalf("unexpected neural output length: %d", len(out))
	}
	if n.Health()["status"] != "ready" {
		t.Fatalf("neural runtime is not ready: %#v", n.Health())
	}

	c, err := prefrontal.New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	neo, err := prefrontal.NewNeocortex(c, n)
	if err != nil {
		t.Fatal(err)
	}

	candidate, err := neo.Evaluate(context.Background(), "recovered-neural-path", []float64{0.2, 0.1, 0.0, 0.3}, 0.0, 0.1, 0.2, 0.3)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.ID != "recovered-neural-path" {
		t.Fatalf("unexpected candidate: %#v", candidate)
	}

	decision, err := neo.Commit(candidate, "recovered historical prefrontal decision path")
	if err != nil {
		t.Fatal(err)
	}
	if decision.CandidateID != candidate.ID {
		t.Fatalf("decision did not preserve candidate identity: %#v", decision)
	}

	health := neo.Health()
	if health["module"] != "PrefrontalNeocortex" {
		t.Fatalf("unexpected prefrontal health: %#v", health)
	}
	if health["neural_binding"] != true {
		t.Fatalf("neural binding not reported: %#v", health)
	}
}
