package prefrontal

import (
	"context"
	"errors"
	"sync"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type Prefrontal struct {
	mu sync.RWMutex
	cfg trinity.PrefrontalConfig
	estimator trinity.CostEstimator
	memory *WorkingMemory
	monitor *ConflictMonitor
	policy *MetaPolicy
}

func NewPrefrontal(cfg trinity.PrefrontalConfig, estimator trinity.CostEstimator) *Prefrontal {
	if cfg.WorkingMemoryLimit < 1 { cfg.WorkingMemoryLimit = 16 }
	if cfg.MetaRLEpsilon < 0 || cfg.MetaRLEpsilon > 1 { cfg.MetaRLEpsilon = 0.1 }
	if cfg.ConflictSensitivity <= 0 || cfg.ConflictSensitivity > 1 { cfg.ConflictSensitivity = 0.5 }
	return &Prefrontal{cfg: cfg, estimator: estimator, memory: NewWorkingMemory(cfg.WorkingMemoryLimit), monitor: NewConflictMonitor(cfg.ConflictSensitivity), policy: NewMetaPolicy(cfg.MetaRLEpsilon)}
}

func (p *Prefrontal) Evaluate(ctx context.Context, w trinity.Workload, compute trinity.ComputeEngine) (trinity.Decision, error) {
	if p == nil { return trinity.Decision{}, errors.New("nil prefrontal") }
	if ctx == nil { return trinity.Decision{}, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return trinity.Decision{}, err }
	if w.ID == "" || w.Kind == "" { return trinity.Decision{}, errors.New("invalid workload") }
	var est trinity.CostEstimate
	var err error
	p.mu.RLock(); estimator := p.estimator; monitor := p.monitor; policy := p.policy; p.mu.RUnlock()
	if estimator != nil { est, err = estimator.Estimate(ctx, w) }
	if err != nil || estimator == nil { est = conservativeEstimate(w) }
	strategy := policy.Choose()
	if w.Precision != "" { strategy.Precision = w.Precision }
	decision := trinity.Decision{Strategy: strategy, Estimate: est}
	decision.ConflictScore = monitor.Score(w, est, 0)
	return approve(decision, 0.75), nil
}

func conservativeEstimate(w trinity.Workload) trinity.CostEstimate {
	memory := w.MemoryNeeded; if memory <= 0 { memory = float64(max(1, w.MatrixSize*w.MatrixSize))*4/1024/1024 }
	latency := float64(max(1, w.BatchSize))*10 + float64(max(1, w.MatrixSize))/256
	return trinity.CostEstimate{LatencyMS: latency, Memory: memory, ComputeCost: latency/100, Confidence: 0.5}
}

func max(a,b int) int { if a > b { return a }; return b }

func (p *Prefrontal) GateWorkingMemory(ctx context.Context, w trinity.Workload, r trinity.Result) error {
	if p == nil { return errors.New("nil prefrontal") }
	if ctx == nil { return errors.New("nil context") }
	if err := ctx.Err(); err != nil { return err }
	score := 0.0; if r.Success { score = 1 }; if r.LatencyMS > 0 { score += 1/(1+r.LatencyMS/1000) }
	p.memory.Put(w, score)
	return nil
}

func (p *Prefrontal) Learn(strategy trinity.Strategy, reward float64) { if p != nil { p.policy.Update(strategy, reward) } }
func (p *Prefrontal) WorkingMemory() []trinity.Workload { if p == nil { return nil }; return p.memory.Snapshot() }
