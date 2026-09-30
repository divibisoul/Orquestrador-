package prefrontal

import (
	"errors"
	"math"
	"sort"
	"strings"
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
	LearningWeight    float64
	Epsilon           float64
	Horizon           time.Duration
}
type OutcomeStats struct {
	Attempts    uint64
	Successes   uint64
	RewardSum   float64
	Confidence  float64
	UpdatedAt   time.Time
}
type Cortex struct {
	mu                 sync.RWMutex
	decisions          []Decision
	outcomes           map[string]OutcomeStats
	threshold          float64
	capacity           int
	policy             Policy
	inhibited          uint64
	evaluated          uint64
	learningObserved   uint64
	decisionNanos      uint64
	lastDecision       time.Time
}

func New(threshold float64, capacity int) (*Cortex, error) {
	if threshold < 0 || threshold > 1 || math.IsNaN(threshold) || math.IsInf(threshold, 0) {
		return nil, errors.New("threshold must be finite and between 0 and 1")
	}
	if capacity < 1 {
		return nil, errors.New("capacity must be positive")
	}
	return &Cortex{
		threshold: threshold,
		capacity: capacity,
		outcomes:  make(map[string]OutcomeStats),
		policy: Policy{
			CostWeight:        1,
			RiskWeight:        1,
			UtilityWeight:     1,
			UncertaintyWeight: .25,
			UrgencyWeight:     .1,
			ImpactWeight:      .2,
			LearningWeight:    .15,
			Epsilon:           0,
			Horizon:           15 * time.Minute,
		},
	}, nil
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
	if v.Cost < 0 || v.Risk < 0 || v.Uncertainty < 0 || v.Urgency < 0 || v.Impact < 0 {
		return errors.New("cost, risk, uncertainty, urgency and impact cannot be negative")
	}
	if v.Risk > 1 || v.Uncertainty > 1 {
		return errors.New("risk and uncertainty must be <= 1")
	}
	return nil
}
func candidateCapability(v Candidate) string {
	if v.Context == nil {
		return ""
	}
	value, ok := v.Context["capability"].(string)
	return strings.TrimSpace(value)
}
func (c *Cortex) learnedCapabilityRewardLocked(capability string) float64 {
	if capability == "" {
		return 0
	}
	stats, ok := c.outcomes[capability]
	if !ok || stats.Attempts == 0 {
		return 0
	}
	denom := stats.Confidence
	if denom <= 0 {
		denom = float64(stats.Attempts)
	}
	if denom <= 0 {
		return 0
	}
	reward := stats.RewardSum / denom
	if reward < -1 {
		return -1
	}
	if reward > 1 {
		return 1
	}
	return reward
}
func (c *Cortex) score(v Candidate) float64 {
	c.mu.RLock()
	learningSignal := c.learnedCapabilityRewardLocked(candidateCapability(v))
	policy := c.policy
	c.mu.RUnlock()
	return policy.UtilityWeight*v.Utility -
		policy.CostWeight*v.Cost -
		policy.RiskWeight*v.Risk -
		policy.UncertaintyWeight*v.Uncertainty +
		policy.UrgencyWeight*v.Urgency +
		policy.ImpactWeight*v.Impact +
		policy.LearningWeight*learningSignal
}
func (c *Cortex) ObserveOutcome(capability string, reward, confidence float64) error {
	capability = strings.TrimSpace(capability)
	if capability == "" {
		return errors.New("learning capability is required")
	}
	if math.IsNaN(reward) || math.IsInf(reward, 0) || reward < -1 || reward > 1 {
		return errors.New("learning reward must be finite and within [-1,1]")
	}
	if math.IsNaN(confidence) || math.IsInf(confidence, 0) || confidence < 0 || confidence > 1 {
		return errors.New("learning confidence must be finite and within [0,1]")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	stats := c.outcomes[capability]
	stats.Attempts++
	if reward > 0 {
		stats.Successes++
	}
	stats.RewardSum += reward * confidence
	stats.Confidence += confidence
	stats.UpdatedAt = time.Now().UTC()
	c.outcomes[capability] = stats
	c.learningObserved++
	return nil
}
func (c *Cortex) Outcome(capability string) (OutcomeStats, bool) {
	capability = strings.TrimSpace(capability)
	c.mu.RLock()
	defer c.mu.RUnlock()
	stats, ok := c.outcomes[capability]
	return stats, ok
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
	c.mu.Lock()
	c.decisionNanos += uint64(time.Since(start).Nanoseconds())
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
	blocked := valid(candidate) != nil || candidate.Risk > candidate.Utility || candidate.Risk >= 1
	c.mu.Lock()
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
	d := Decision{ID: candidate.ID, Score: c.score(candidate), Reason: reason, Justification: "policy=weighted-risk-utility+learned-outcomes;validated=true", Committed: true, At: time.Now().UTC(), Outcome: "pending"}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.decisions = append(c.decisions, d)
	if len(c.decisions) > c.capacity {
		c.decisions = c.decisions[len(c.decisions)-c.capacity:]
	}
	c.lastDecision = time.Now().UTC()
	c.decisionNanos += uint64(time.Since(start).Nanoseconds())
	return d, nil
}
func (c *Cortex) Recall(limit int) []Decision {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if limit <= 0 || limit > len(c.decisions) {
		limit = len(c.decisions)
	}
	out := make([]Decision, limit)
	copy(out, c.decisions[len(c.decisions)-limit:])
	return out
}
func (c *Cortex) Health() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	avg := 0.0
	if c.evaluated > 0 {
		avg = float64(c.decisionNanos) / float64(c.evaluated) / 1e6
	}
	learningObservations := c.learningObserved
	return map[string]any{
		"status":                 "ready",
		"threshold":              c.threshold,
		"capacity":               c.capacity,
		"decisions":              len(c.decisions),
		"evaluated":              c.evaluated,
		"inhibited":              c.inhibited,
		"inhibition_rate": func() float64 {
			if c.evaluated == 0 {
				return 0
			}
			return float64(c.inhibited) / float64(c.evaluated)
		}(),
		"learning_observations":  learningObservations,
		"learned_capabilities":   len(c.outcomes),
		"learning_weight":        c.policy.LearningWeight,
		"avg_decision_ms":        avg,
		"last_decision":          c.lastDecision,
	}
}
