package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental"
	"github.com/divibisoul/Orquestrador-/core/neuralfabric"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/core/trinity"
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

	// The legacy state store remains part of the legacy path. It is not used
	// as evidence of durable execution; durable recovery is a separate gate.
	store := state.NewStore()
	version := store.Put("workflow", []byte("completed"))
	fmt.Printf("[state][legacy] workflow version=%d\n", version)

	if trinityEnabled() {
		if err := runTrinity(ctx, goal); err != nil {
			fmt.Printf("[trinity] activation failed: %v\n", err)
			return
		}
	}
	fmt.Println("orchestrator runtime path completed")
}

func trinityEnabled() bool {
	v, err := strconv.ParseBool(os.Getenv("TRINITY_ENABLED"))
	return err == nil && v
}

func runTrinity(ctx context.Context, goal string) error {
	trace := os.Getenv("TRINITY_TRACE_ID")
	if trace == "" {
		trace = fmt.Sprintf("trinity-%d", time.Now().UnixNano())
	}
	ctx = trinity.WithTraceID(ctx, trace)

	cfg := trinity.TrinityConfig{
		PFCEnabled: true, FabricEnabled: true, ComputeEnabled: true,
		RiskThreshold: 0.75, FallbackMode: "retry",
		Prefrontal: trinity.PrefrontalConfig{WorkingMemoryLimit: 16, MetaRLEpsilon: 0.1, ConflictSensitivity: 0.5},
		Fabric: trinity.FabricConfig{DecisionTreeDepth: 6, LearningRate: 0.01, FeedbackDiscount: 0.9},
		Compute: trinity.ComputeConfig{Mode: "auto", PrecisionFallback: "fp32", EfficiencyFactor: 0.7},
	}

	compute := transcendental.NewEngine(cfg.Compute)
	pfc := prefrontal.NewPrefrontal(cfg.Prefrontal, compute)
	fabric := neuralfabric.NewFabric(cfg.Fabric)
	tri := &trinity.TrinityOrchestrator{PFC: pfc, Fabric: fabric, Compute: compute, Config: cfg}

	result, err := tri.ExecuteTask(ctx, trinity.Task{ID: "activation-demo", Kind: "text", Payload: goal, BatchSize: 1, Precision: "fp32"})
	if err != nil {
		return fmt.Errorf("trace=%s: %w", trace, err)
	}
	fmt.Printf("[trinity][trace=%s] target=%s model=%s backend=%s latency_ms=%.3f\n", trace, result.Route.Target, result.Route.Model, result.Metadata["backend"], result.LatencyMS)
	return nil
}
