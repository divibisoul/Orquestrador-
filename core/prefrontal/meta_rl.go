package prefrontal

import (
	"math/rand"
	"sync"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type MetaPolicy struct {
	mu sync.RWMutex
	epsilon float64
	values map[string]float64
	alpha float64
}

func NewMetaPolicy(epsilon float64) *MetaPolicy {
	if epsilon < 0 { epsilon = 0 }
	if epsilon > 1 { epsilon = 1 }
	return &MetaPolicy{epsilon: epsilon, alpha: 0.1, values: map[string]float64{"precision": 0.5, "parallelism": 0.5}}
}

func (p *MetaPolicy) Choose() trinity.Strategy {
	if p == nil { return trinity.Strategy{Name: "balanced", Parallelism: 1, Precision: "fp32"} }
	p.mu.RLock(); eps, precision, parallel := p.epsilon, p.values["precision"], p.values["parallelism"]; p.mu.RUnlock()
	if rand.Float64() < eps {
		if rand.Intn(2) == 0 { return trinity.Strategy{Name: "precision", Parallelism: 1, Precision: "fp64"} }
		return trinity.Strategy{Name: "parallel", Parallelism: 4, Precision: "fp32"}
	}
	if precision >= parallel { return trinity.Strategy{Name: "precision", Parallelism: 1, Precision: "fp64"} }
	return trinity.Strategy{Name: "parallel", Parallelism: 4, Precision: "fp32"}
}

func (p *MetaPolicy) Update(strategy trinity.Strategy, reward float64) {
	if p == nil { return }
	if reward < -1 { reward = -1 }; if reward > 1 { reward = 1 }
	key := "parallelism"; if strategy.Name == "precision" { key = "precision" }
	p.mu.Lock(); p.values[key] += p.alpha*(reward-p.values[key]); p.mu.Unlock()
}
