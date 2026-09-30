package cognitive

import (
	"context"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/prefrontal"
)

func TestWorkingMemoryRetainsEvidenceAndEvictsLowRelevance(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WorkingMemoryItems = 2
	cfg.WorkingMemoryTTL = time.Hour
	memory := NewWorkingMemory(cfg)

	if err := memory.Put("low", map[string]any{"value": 1}, 0.1); err != nil {
		t.Fatal(err)
	}
	if err := memory.Put("high", map[string]any{"value": 2}, 0.9); err != nil {
		t.Fatal(err)
	}
	if err := memory.Put("higher", map[string]any{"value": 3}, 1.0); err != nil {
		t.Fatal(err)
	}

	if _, ok := memory.Get("low"); ok {
		t.Fatal("low relevance entry should have been evicted")
	}
	if _, ok := memory.Get("high"); !ok {
		t.Fatal("high relevance evidence should remain")
	}
}

func TestGoalStoreRejectsExpiredGoalsAndPreservesCorrelation(t *testing.T) {
	store := NewGoalStore()
	goal := Goal{
		ID: "goal-1", Objective: "recover cognitive path", CorrelationID: "corr-1",
		Capabilities: []string{"ai.infer"}, ExpiresAt: time.Now().UTC().Add(time.Minute),
	}
	if err := store.Put(goal); err != nil {
		t.Fatal(err)
	}
	got, ok := store.Get(goal.ID)
	if !ok || got.CorrelationID != "corr-1" {
		t.Fatalf("goal was not retained: %#v", got)
	}

	expired := goal
	expired.ID = "goal-expired"
	expired.ExpiresAt = time.Now().UTC().Add(-time.Second)
	if err := store.Put(expired); err == nil {
		t.Fatal("expired goals must be rejected")
	}
}

func TestCriticUsesExistingPrefrontalAndBlocksIrreversibleWithoutSARA(t *testing.T) {
	cortex, err := prefrontal.New(0.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	critic, err := NewCritic(cortex, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	reversible := Goal{
		ID: "rev", Objective: "analyze", CorrelationID: "corr-rev",
		Capabilities: []string{"ai.infer"}, Risk: 0.01, Cost: 0.01,
		Urgency: 0.1, Impact: 0.1, ExpiresAt: time.Now().UTC().Add(time.Minute),
	}
	if err := critic.Check(ctx, reversible); err != nil {
		t.Fatalf("reversible goal unexpectedly blocked: %v", err)
	}

	irreversible := reversible
	irreversible.ID = "irr"
	irreversible.Irreversible = true
	if err := critic.Check(ctx, irreversible); err == nil {
		t.Fatal("irreversible goal must require configured SARA policy audit")
	}
}

func TestMeshExecutorRequiresPeerContract(t *testing.T) {
	if _, err := NewMeshExecutor(nil); err == nil {
		t.Fatal("MeshExecutor must require a real Mesh peer contract")
	}
}

func TestEnsureGeminiPolicyPreservesGoalPolicy(t *testing.T) {
	goal := Goal{ID: "gemini-goal", Cost: 0.2, Risk: 0.1, Urgency: 0.4, Impact: 0.7}
	step := Step{Capability: "gemini.text.generate", Payload: map[string]any{}}
	ensureGeminiPolicy(&step, goal)
	if _, ok := step.Payload["candidate_json"]; !ok {
		t.Fatal("Gemini policy was not attached to the cognitive step")
	}
}
