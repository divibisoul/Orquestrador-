package main

import (
	"context"
	"fmt"

	"github.com/divibisoul/Orquestrador-/core/neuralfabric"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/core/superagi"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/state"
)

func main() {
	ctx := context.Background()
	goal := "recover orchestrator"

	cortex := prefrontal.New()
	plan := cortex.GeneratePlan(goal, []string{"plan", "route", "execute"})
	decision := cortex.Decide([]prefrontal.Plan{plan}, 10)
	fmt.Printf("[cortex] selected %s\n", decision.ID)

	engine := orchestrator.NewEngine(2)
	steps := []orchestrator.Step{
		{ID: "plan", Run: func(context.Context) error { fmt.Println("[orchestrator] plan"); return nil }},
		{ID: "route", Run: func(context.Context) error { fmt.Println("[orchestrator] route"); return nil }},
		{ID: "execute", Run: func(context.Context) error { fmt.Println("[orchestrator] execute"); return nil }},
	}
	if err := engine.CreateWorkflow("demo", steps); err != nil {
		fmt.Printf("orchestrator setup failed: %v\n", err)
		return
	}
	for range steps {
		if err := engine.ExecuteStep(ctx, "demo"); err != nil {
			fmt.Printf("workflow execution failed: %v\n", err)
			return
		}
	}

	registry := mesh.NewRegistry()
	if err := registry.Announce(mesh.Node{ID: "local", Status: "ready", Capabilities: []string{"inference"}, CPU: true}); err != nil {
		fmt.Printf("mesh setup failed: %v\n", err)
		return
	}
	fmt.Printf("[mesh] nodes=%d\n", len(registry.Discover(ctx, "inference")))

	nf := neuralfabric.NewRuntime()
	if _, err := nf.Route(ctx, neuralfabric.State{Goal: goal}, []neuralfabric.Route{{NodeID: "local", DeviceID: "local", Precision: "int8", BatchSize: 1}}); err != nil {
		fmt.Printf("neural fabric routing failed: %v\n", err)
		return
	}
	fmt.Println("[neural-fabric] route accepted")

	agi := superagi.NewRuntime()
	if _, err := agi.GenerateEmbedding(ctx, goal); err != nil {
		fmt.Printf("super AGI failed: %v\n", err)
		return
	}
	fmt.Println("[super-agi] deterministic embedding generated")

	store := state.NewStore()
	version := store.Put("workflow", []byte("completed"))
	fmt.Printf("[state] workflow version=%d\n", version)
	fmt.Println("six-plane orchestrator demo completed")
}
