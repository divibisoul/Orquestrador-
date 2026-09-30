package prefrontal

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

type Candidate struct {
	ID          string
	Score       float64
	Cost        float64
	Risk        float64
	Utility     float64
	Uncertainty float64
	Urgency     float64
	Impact      float64
	Context     map[string]any
}
type Decision struct {
	ID            string
	Score         float64
	Reason        string
	Justification string
	Committed     bool
	At            time.Time
	Outcome       string
}
type Policy struct {
	CostWeight        float64
	RiskWeight        float64
	UtilityWeight     float64
	UncertaintyWeight float64
	UrgencyWeight     float64
	ImpactWeight      float64
	Epsilon           float64
	Horizon           time.Duration
	MaxRisk           float64
	MaxUncertainty    float64
	MinUtility        float64
}
type Cortex struct {
	mu               sync.RWMutex
	decisions        []Decision
	decisionHistory  []Decision
	threshold        float64
	capacity         int
	policy           Policy
	inhibited        uint64
	inhibitionChecks uint64
	evaluated        uint64
	commits          uint64
	decisionNanos    uint64
	decisionCount    uint64
	evaluationNanos  uint64
	commitNanos      uint64
	lastDecision     time.Time
	workingMemory    map[string]workingMemoryEntry
	workingMemoryArchive []workingMemoryEntry
	taskFrames       []TaskFrame
	currentTask      string
}

func New(threshold float64, capacity int) (*Cortex, error) {
	if threshold < 0 || threshold > 1 || math.IsNaN(threshold) || math.IsInf(threshold, 0) {
		return nil, errors.New("threshold must be finite and between 0 and 1")
	}
	if capacity < 1 {
		return nil, errors.New("capacity must be positive")
	}
	policy := Policy{CostWeight: 1, RiskWeight: 1, UtilityWeight: 1, UncertaintyWeight: .25, UrgencyWeight: .1, ImpactWeight: .2, Epsilon: 0, Horizon: 15 * time.Minute, MaxRisk: .8, MaxUncertainty: .8, MinUtility: 0}
	return &Cortex{threshold: threshold, capacity: capacity, policy: policy, workingMemory: make(map[string]workingMemoryEntry)}, nil
}
func valid(v Candidate) error {
	if v.ID == "" {
		return errors.New("candidate id is required")
	}
	for _, x := range []float64{v.Cost, v.Risk, v.Utility, v.Score, v.Uncertainty, v.Urgency, v.Impact} {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return errors.New("candidate contains non-finite value")
		}
	}
	if v.Cost < 0 || v.Risk < 0 || v.Utility < 0 || v.Uncertainty < 0 || v.Urgency < 0 || v.Impact < 0 {
		return errors.New("cost, risk, utility, uncertainty, urgency and impact cannot be negative")
	}
	if v.Risk > 1 || v.Uncertainty > 1 {
		return errors.New("risk and uncertainty must be <= 1")
	}
	return nil
}
func boundedPositive(v float64) float64 {
	if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v / (1 + v)
}

func (c *Cortex) score(v Candidate) float64 {
	weights := c.policy.CostWeight + c.policy.RiskWeight + c.policy.UtilityWeight + c.policy.UncertaintyWeight + c.policy.UrgencyWeight + c.policy.ImpactWeight
	if weights <= 0 || math.IsNaN(weights) || math.IsInf(weights, 0) {
		return -math.MaxFloat64
	}
	utility := boundedPositive(v.Utility)
	cost := boundedPositive(v.Cost)
	urgency := boundedPositive(v.Urgency)
	impact := boundedPositive(v.Impact)
	risk := boundedPositive(v.Risk)
	uncertainty := boundedPositive(v.Uncertainty)
	raw := c.policy.UtilityWeight*utility -
		c.policy.CostWeight*cost -
		c.policy.RiskWeight*risk -
		c.policy.UncertaintyWeight*uncertainty +
		c.policy.UrgencyWeight*urgency +
		c.policy.ImpactWeight*impact
	return raw / weights
}
func (c *Cortex) Evaluate(candidates []Candidate) (Candidate, error) {
	start := time.Now()
	best := Candidate{}
	bestScore := -math.MaxFloat64
	found := false
	for _, v := range candidates {
		c.mu.Lock()
		c.evaluated++
		c.mu.Unlock()
		if err := valid(v); err != nil {
			continue
		}
		s := c.score(v)
		if s > bestScore {
			best, bestScore, found = v, s, true
		}
	}
	duration := time.Since(start)
	c.mu.Lock()
	ns := uint64(duration.Nanoseconds())
	c.decisionNanos += ns
	c.decisionCount++
	c.evaluationNanos += ns
	c.mu.Unlock()
	if !found {
		return Candidate{}, errors.New("no valid candidates")
	}
	if bestScore < c.threshold {
		return Candidate{}, errors.New("no candidate exceeds decision threshold")
	}
	best.Score = bestScore
	return best, nil
}
func dominated(a, b Candidate) bool {
	return b.Cost <= a.Cost && b.Risk <= a.Risk && b.Uncertainty <= a.Uncertainty && b.Utility >= a.Utility && (b.Cost < a.Cost || b.Risk < a.Risk || b.Uncertainty < a.Uncertainty || b.Utility > a.Utility)
}
func (c *Cortex) Plan(candidates []Candidate) ([]Candidate, error) {
	if len(candidates) == 0 {
		return nil, errors.New("no candidates")
	}
	out := make([]Candidate, 0, len(candidates))
	for _, v := range candidates {
		if valid(v) == nil {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("no valid candidates")
	}
	pareto := out[:0]
	for i, a := range out {
		dom := false
		for j, b := range out {
			if i != j && dominated(a, b) {
				dom = true
				break
			}
		}
		if !dom {
			a.Score = c.score(a)
			pareto = append(pareto, a)
		}
	}
	sort.SliceStable(pareto, func(i, j int) bool {
		if pareto[i].Score == pareto[j].Score {
			return pareto[i].ID < pareto[j].ID
		}
		return pareto[i].Score > pareto[j].Score
	})
	if len(pareto) > c.capacity {
		pareto = pareto[:c.capacity]
	}
	return pareto, nil
}
func (c *Cortex) Prioritize(candidates []Candidate) ([]Candidate, error) {
	out, err := c.Plan(candidates)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			if out[i].Urgency == out[j].Urgency {
				return out[i].ID < out[j].ID
			}
			return out[i].Urgency > out[j].Urgency
		}
		return out[i].Score > out[j].Score
	})
	return out, nil
}
func (c *Cortex) Inhibit(candidate Candidate) bool {
	blocked := valid(candidate) != nil ||
		candidate.Risk > c.policy.MaxRisk ||
		candidate.Uncertainty > c.policy.MaxUncertainty ||
		candidate.Utility < c.policy.MinUtility
	c.mu.Lock()
	c.inhibitionChecks++
	if blocked {
		c.inhibited++
	}
	c.mu.Unlock()
	return blocked
}
func (c *Cortex) Select(candidates []Candidate) (Candidate, error) {
	planned, err := c.Prioritize(candidates)
	if err != nil {
		return Candidate{}, err
	}
	for _, v := range planned {
		if !c.Inhibit(v) && c.score(v) >= c.threshold {
			return v, nil
		}
	}
	return Candidate{}, errors.New("all candidates inhibited or below threshold")
}
func (c *Cortex) ValidateAction(candidate Candidate) error {
	if err := valid(candidate); err != nil {
		return err
	}
	if c.Inhibit(candidate) {
		return errors.New("action inhibited by risk policy")
	}
	if c.score(candidate) < c.threshold {
		return errors.New("action below decision threshold")
	}
	return nil
}
func (c *Cortex) Commit(candidate Candidate, reason string) (Decision, error) {
	start := time.Now()
	if err := c.ValidateAction(candidate); err != nil {
		return Decision{}, err
	}
	if reason == "" {
		return Decision{}, errors.New("decision reason is required")
	}
	d := Decision{ID: candidate.ID, Score: c.score(candidate), Reason: reason, Justification: "policy=weighted-risk-utility;validated=true", Committed: true, At: time.Now().UTC(), Outcome: "pending"}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.decisions = append(c.decisions, d)
	c.decisionHistory = append(c.decisionHistory, d)
	c.lastDecision = time.Now().UTC()
	c.commits++
	ns := uint64(time.Since(start).Nanoseconds())
	c.decisionNanos += ns
	c.decisionCount++
	c.commitNanos += ns
	return d, nil
}
func (c *Cortex) Recall(limit int) []Decision {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if limit <= 0 || limit > len(c.decisionHistory) {
		limit = len(c.decisionHistory)
	}
	out := make([]Decision, limit)
	copy(out, c.decisionHistory[len(c.decisionHistory)-limit:])
	return out
}

func (c *Cortex) HistorySize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.decisionHistory)
}
func (c *Cortex) Health() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	avgDecision := 0.0
	if c.decisionCount > 0 {
		avgDecision = float64(c.decisionNanos) / float64(c.decisionCount) / 1e6
	}
	avgEvaluate := 0.0
	if c.evaluated > 0 {
		avgEvaluate = float64(c.evaluationNanos) / float64(c.evaluated) / 1e6
	}
	avgCommit := 0.0
	if c.commits > 0 {
		avgCommit = float64(c.commitNanos) / float64(c.commits) / 1e6
	}
	inhibitionRate := 0.0
	if c.inhibitionChecks > 0 {
		inhibitionRate = float64(c.inhibited) / float64(c.inhibitionChecks)
	}
	return map[string]any{
		"status": "ready",
		"threshold": c.threshold,
		"capacity": c.capacity,
		"decisions": len(c.decisions),
		"decision_history_size": len(c.decisionHistory),
		"evaluated": c.evaluated,
		"commits": c.commits,
		"inhibited": c.inhibited,
		"inhibition_checks": c.inhibitionChecks,
		"inhibition_rate": inhibitionRate,
		"avg_decision_ms": avgDecision,
		"avg_evaluate_ms": avgEvaluate,
		"avg_commit_ms": avgCommit,
		"working_memory_size": len(c.workingMemory),
		"working_memory_archive_size": len(c.workingMemoryArchive),
		"current_task": c.currentTask,
		"task_switches": len(c.taskFrames),
		"last_decision": c.lastDecision,
	}
}
