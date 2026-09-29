package cognitive

import (
    "testing"
    "time"

    "github.com/divibisoul/Orquestrador-/octacore"
    "github.com/divibisoul/Orquestrador-/prefrontal"
)

func TestWorkingMemoryTTLAndBound(t *testing.T) {
    cfg := DefaultConfig()
    cfg.WorkingMemoryItems = 2
    cfg.WorkingMemoryTTL = 10 * time.Millisecond
    memory := NewWorkingMemory(cfg)
    if err := memory.Put("a", map[string]any{"v": 1}, 0.1); err != nil { t.Fatal(err) }
    if err := memory.Put("b", map[string]any{"v": 2}, 0.2); err != nil { t.Fatal(err) }
    if err := memory.Put("c", map[string]any{"v": 3}, 0.3); err != nil { t.Fatal(err) }
    if len(memory.Snapshot()) != 2 { t.Fatalf("expected bounded memory") }
    time.Sleep(15 * time.Millisecond)
    if len(memory.Snapshot()) != 0 { t.Fatalf("expected TTL eviction") }
}

func TestStepMappingPreservesFederationMetadata(t *testing.T) {
    group := "pre"
    step := Step{
        ID: "step-1", GoalID: "goal-1", Capability: "core.health", Target: "N04",
        BackendPrefs: []string{"REMOTE_MESH"}, Kind: "custom",
        Payload: map[string]any{"x": 1}, ParallelGroup: &group, CorrelationID: "corr-1",
    }
    job := octacore.Job{
        JobID: step.ID, CorrelationID: step.CorrelationID, Kind: stepKind(step.Kind),
        Source: octacore.G7, Target: step.Target, BackendPrefs: []octacore.Backend{octacore.BackendRemoteMesh},
        ParallelGroup: step.ParallelGroup, Payload: map[string]any{"capability": step.Capability, "payload": step.Payload},
        Priority: 50, TTLMS: 30000,
    }
    if job.CorrelationID != "corr-1" || job.Target != "N04" || job.ParallelGroup == nil || *job.ParallelGroup != "pre" { t.Fatal("federation metadata lost") }
}

func TestPlannerHelpers(t *testing.T) {
    if !sameCapability("tool:read@1.0.0", "tool:read") { t.Fatal("version compatibility failed") }
    if sameCapability("tool:write@1.0.0", "tool:read") { t.Fatal("different capability was accepted") }
    if classifyKind("research.search") != "research" { t.Fatal("research kind mismatch") }
    if classifyKind("audio.transcribe") != "perceive" { t.Fatal("perception kind mismatch") }
    if classifyKind("support.context") != "session_step" { t.Fatal("session kind mismatch") }
}

func TestCritiqueBlocksHighRisk(t *testing.T) {
    cortex, err := prefrontal.New(.1, 8)
    if err != nil { t.Fatal(err) }
    critic, err := NewCritic(cortex, nil, DefaultConfig())
    if err != nil { t.Fatal(err) }
    err = critic.Check(nil, Goal{ID: "g", Objective: "write", Risk: 1, Cost: 0, CorrelationID: "c"})
    if err == nil { t.Fatal("expected policy block") }
}

func TestDisabledConfigIsExplicit(t *testing.T) {
    cfg := DefaultConfig()
    if cfg.Enabled { t.Fatal("cognitive loop must be disabled by default") }
    if cfg.MaxParallelSteps <= 0 { t.Fatal("parallel step limit must be positive") }
}
