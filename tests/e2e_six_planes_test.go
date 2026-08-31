package tests

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/neuralfabric"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/core/superagi"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/state"
)

func TestSixPlanesE2E(t *testing.T) {
	ctx := context.Background()

	// 1. Neo-Cortex: objective -> deterministic plan -> decision.
	cortex := prefrontal.New()
	plan := cortex.GeneratePlan("recover orchestrator", []string{"discover", "route", "execute"})
	decision := cortex.Decide([]prefrontal.Plan{plan}, 10)
	if decision.ID != plan.ID {
		t.Fatalf("cortex selected %q, want %q", decision.ID, plan.ID)
	}

	// 2. Orchestrator: materialize the plan as a runnable workflow.
	engine := orchestrator.NewEngine(2)
	steps := []orchestrator.Step{
		{ID: "discover", Run: func(context.Context) error { return nil }},
		{ID: "route", Run: func(context.Context) error { return nil }},
		{ID: "execute", Run: func(context.Context) error { return nil }},
	}
	if err := engine.CreateWorkflow("recovery-e2e", steps); err != nil {
		t.Fatal(err)
	}
	for range steps {
		if err := engine.ExecuteStep(ctx, "recovery-e2e"); err != nil {
			t.Fatal(err)
		}
	}
	status, err := engine.GetWorkflowStatus("recovery-e2e")
	if err != nil || status != orchestrator.Completed {
		t.Fatalf("workflow status=%q err=%v", status, err)
	}

	// 3. Mesh: announce and discover a capable execution node.
	registry := mesh.NewRegistry()
	if err := registry.Announce(mesh.Node{
		ID: "local-node",
		Status: "ready",
		Capabilities: []string{"orchestration", "inference"},
		CPU: true,
	}); err != nil {
		t.Fatal(err)
	}
	if nodes := registry.Discover(ctx, "inference"); len(nodes) != 1 {
		t.Fatalf("mesh discovered %d inference nodes, want 1", len(nodes))
	}

	// 4. Neural Fabric: score a route for the announced node.
	nf := neuralfabric.NewRuntime()
	route, err := nf.Route(ctx, neuralfabric.State{Goal: "recover orchestrator"}, []neuralfabric.Route{{
		NodeID: "local-node", DeviceID: "local-node", Precision: "int8", BatchSize: 1,
	}})
	if err != nil || route.NodeID != "local-node" {
		t.Fatalf("neural fabric route=%+v err=%v", route, err)
	}

	// 5. Super AGI: deterministic fallback remains functional without a provider.
	agi := superagi.NewRuntime()
	embedding, err := agi.GenerateEmbedding(ctx, "recovery-e2e")
	if err != nil || len(embedding) != 32 {
		t.Fatalf("embedding length=%d err=%v", len(embedding), err)
	}

	// 6. State: versioned write/read/CAS semantics.
	store := state.NewStore()
	version := store.Put("workflow", []byte("recovery-e2e"))
	value, gotVersion, ok := store.Get("workflow")
	if !ok || gotVersion != version || string(value) != "recovery-e2e" {
		t.Fatalf("state read value=%q version=%d ok=%v", value, gotVersion, ok)
	}
	if !store.CompareAndSwap("workflow", version, []byte("completed")) {
		t.Fatal("state CAS failed")
	}

	t.Log("six-plane E2E: cortex -> orchestrator -> mesh -> neural fabric -> super AGI -> state")
}
