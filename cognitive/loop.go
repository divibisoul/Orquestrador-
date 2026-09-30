package cognitive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/prefrontal"
)

type Executor interface {
	Execute(context.Context, string, map[string]any, string) (map[string]any, string, error)
}

type LocalExecutor func(context.Context, string, map[string]any, string) (map[string]any, string, error)

type MeshExecutor struct {
	peers *mesh.PeerClient
	local LocalExecutor
}

func NewMeshExecutor(peers *mesh.PeerClient, local ...LocalExecutor) (*MeshExecutor, error) {
	if peers == nil {
		return nil, errors.New("mesh peer client is required")
	}
	var localExec LocalExecutor
	if len(local) > 0 {
		localExec = local[0]
	}
	return &MeshExecutor{peers: peers, local: localExec}, nil
}

func (e *MeshExecutor) Execute(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
	capability = strings.TrimSpace(capability)
	correlation = strings.TrimSpace(correlation)
	if ctx == nil {
		return nil, "", errors.New("context is nil")
	}
	if capability == "" {
		return nil, "", errors.New("capability is required")
	}
	if correlation == "" {
		return nil, "", errors.New("correlation is required")
	}
	if e.local != nil && isGeminiCapability(capability) {
		return e.local(ctx, capability, payload, correlation)
	}
	return e.peers.CallBestDynamic(ctx, capability, payload, correlation)
}

func isGeminiCapability(capability string) bool {
	switch strings.TrimSpace(capability) {
	case "gemini.text.generate", "gemini.multimodal.generate", "gemini.audio.transcribe", "gemini.audio.analyze", "gemini.speech.synthesize":
		return true
	default:
		return false
	}
}

type Critic struct {
	cortex *prefrontal.Cortex
	sara   *backend.SARAProxy
	cfg    Config
}

func NewCritic(c *prefrontal.Cortex, s *backend.SARAProxy, cfg Config) (*Critic, error) {
	if c == nil {
		return nil, errors.New("prefrontal cortex is required")
	}
	return &Critic{cortex: c, sara: s, cfg: cfg}, nil
}

func (c *Critic) Check(ctx context.Context, g Goal) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if strings.TrimSpace(g.ID) == "" || strings.TrimSpace(g.Objective) == "" {
		return errors.New("goal id and objective are required")
	}
	candidate := prefrontal.Candidate{
		ID: g.ID, Risk: g.Risk, Cost: g.Cost, Utility: 1 - g.Risk,
		Urgency: g.Urgency, Impact: g.Impact, Uncertainty: 0,
		Score: 1 - g.Risk,
	}
	if err := c.cortex.ValidateAction(candidate); err != nil {
		return fmt.Errorf("LOCAL_CRITIQUE_BLOCKED: %w", err)
	}
	if g.Irreversible && c.cfg.RequireSaraForPolicy {
		if c.sara == nil || !c.sara.Configured() {
			return errors.New("SARA_POLICY_UNAVAILABLE")
		}
		if _, err := c.sara.Audit(ctx, g.Objective, g.CorrelationID); err != nil {
			return fmt.Errorf("SARA_POLICY_AUDIT_FAILED: %w", err)
		}
	}
	return nil
}

type Loop struct {
	cfg      Config
	goals    *GoalStore
	memory   *WorkingMemory
	executor Executor
	critic   *Critic
	observer func(Observation)
	store    *backend.SupabaseStore
}

func New(cfg Config, goals *GoalStore, memory *WorkingMemory, executor Executor, critic *Critic, store *backend.SupabaseStore) (*Loop, error) {
	if !cfg.Enabled {
		return nil, errors.New("COGNITIVE_LOOP_DISABLED")
	}
	if goals == nil || memory == nil || executor == nil || critic == nil {
		return nil, errors.New("goal,memory,executor and critic are required")
	}
	return &Loop{cfg: cfg, goals: goals, memory: memory, executor: executor, critic: critic, store: store}, nil
}

func (l *Loop) Enabled() bool {
	return l != nil && l.cfg.Enabled
}

func (l *Loop) Run(ctx context.Context, g Goal) ([]Observation, error) {
	if !l.Enabled() {
		return nil, errors.New("COGNITIVE_LOOP_DISABLED")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	normalizeGoal(&g, l.cfg.GoalTTL)
	if len(g.Capabilities) == 0 {
		return nil, errors.New("goal requires capabilities")
	}
	if err := l.goals.Put(g); err != nil {
		return nil, err
	}
	if err := l.critic.Check(ctx, g); err != nil {
		return nil, err
	}
	out := make([]Observation, 0, len(g.Capabilities))
	for i, capability := range g.Capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" {
			return out, errors.New("empty capability")
		}
		step := Step{
			ID: fmt.Sprintf("%s-step-%d", g.ID, i+1), GoalID: g.ID,
			Capability: capability, Payload: cloneMap(g.Input),
			CorrelationID: g.CorrelationID,
		}
		ensureGeminiPolicy(&step, g)
		obs := l.executeStep(ctx, step)
		out = append(out, obs)
		if err := l.recordStep(ctx, step, obs); err != nil && obs.OK {
			obs.OK = false
			obs.Error = "COGNITIVE_OBSERVATION_PERSIST_FAILED: " + err.Error()
			out[len(out)-1] = obs
		}
		if !obs.OK {
			return out, fmt.Errorf("goal step %s failed: %s", step.ID, obs.Error)
		}
	}
	return out, nil
}

func (l *Loop) executeStep(ctx context.Context, step Step) Observation {
	started := time.Now()
	value, peer, err := l.executor.Execute(ctx, step.Capability, step.Payload, step.CorrelationID)
	obs := Observation{
		GoalID: gID(step), StepID: step.ID, Capability: step.Capability,
		OK: err == nil, Peer: peer, LatencyMS: time.Since(started).Milliseconds(),
		CorrelationID: step.CorrelationID, At: time.Now().UTC(),
	}
	if err != nil {
		obs.Error = err.Error()
	}
	if err == nil {
		_ = l.memory.Put(step.ID, map[string]any{"output": value, "observation": obs}, 1)
	} else {
		_ = l.memory.Put(step.ID, map[string]any{"observation": obs}, 1)
	}
	if l.observer != nil {
		l.observer(obs)
	}
	return obs
}

func (l *Loop) recordStep(ctx context.Context, step Step, obs Observation) error {
	if l.store == nil || !l.store.Configured() {
		return nil
	}
	status := "error"
	if obs.OK {
		status = "ok"
	}
	return l.store.RecordRun(ctx, map[string]any{
		"trace_id": obs.CorrelationID, "correlation_id": obs.CorrelationID,
		"source": "N07.cognitive", "status": status,
		"metadata": map[string]any{
			"goal_id": obs.GoalID, "step_id": obs.StepID, "capability": obs.Capability,
			"peer": obs.Peer, "latency_ms": obs.LatencyMS, "error": obs.Error,
		},
	})
}

func normalizeGoal(g *Goal, ttl time.Duration) {
	now := time.Now().UTC()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	if strings.TrimSpace(g.CorrelationID) == "" {
		g.CorrelationID = fmt.Sprintf("goal-%d", now.UnixNano())
	}
	if g.ExpiresAt.IsZero() {
		g.ExpiresAt = now.Add(ttl)
	}
}

func ensureGeminiPolicy(step *Step, g Goal) {
	if !isGeminiCapability(step.Capability) {
		return
	}
	if _, ok := step.Payload["candidate_json"]; ok {
		return
	}
	policy := map[string]any{
		"id": g.ID, "cost": g.Cost, "risk": g.Risk,
		"urgency": g.Urgency, "impact": g.Impact,
	}
	raw, err := json.Marshal(policy)
	if err == nil {
		step.Payload["candidate_json"] = string(raw)
	}
}

func gID(step Step) string { return step.GoalID }
