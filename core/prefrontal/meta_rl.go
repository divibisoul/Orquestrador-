package prefrontal

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type deterministicRNG struct {
	state uint64
}

func newDeterministicRNG(seed int64) *deterministicRNG {
	digest := sha256.Sum256([]byte(time.Unix(0, seed).UTC().Format(time.RFC3339Nano)))
	s := binary.LittleEndian.Uint64(digest[:8])
	if s == 0 {
		s = 0x9e3779b97f4a7c15
	}
	return &deterministicRNG{state: s}
}

func (r *deterministicRNG) next() uint64 {
	x := r.state
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	r.state = x
	return x * 2685821657736338717
}

func (r *deterministicRNG) float64() float64 {
	return float64(r.next()>>11) * (1.0 / 9007199254740992.0)
}

func (r *deterministicRNG) bit() bool {
	return r.next()&1 == 0
}

type MetaPolicy struct {
	mu      sync.RWMutex
	rngMu   sync.Mutex
	rng     *deterministicRNG
	epsilon float64
	values  map[string]float64
	alpha   float64
}

func NewMetaPolicy(epsilon float64) *MetaPolicy {
	return NewMetaPolicyWithSeed(epsilon, time.Now().UnixNano())
}

func NewMetaPolicyWithSeed(epsilon float64, seed int64) *MetaPolicy {
	if epsilon < 0 {
		epsilon = 0
	}
	if epsilon > 1 {
		epsilon = 1
	}
	return &MetaPolicy{rng: newDeterministicRNG(seed), epsilon: epsilon, alpha: 0.1, values: map[string]float64{"precision": 0.5, "parallelism": 0.5}}
}

func (p *MetaPolicy) Choose() trinity.Strategy {
	if p == nil {
		return trinity.Strategy{Name: "balanced", Parallelism: 1, Precision: "fp32"}
	}
	p.mu.RLock()
	eps, precision, parallel := p.epsilon, p.values["precision"], p.values["parallelism"]
	p.mu.RUnlock()
	p.rngMu.Lock()
	explore := p.rng != nil && p.rng.float64() < eps
	explorePrecision := p.rng != nil && p.rng.bit()
	p.rngMu.Unlock()
	if explore {
		if explorePrecision {
			return trinity.Strategy{Name: "precision", Parallelism: 1, Precision: "fp64"}
		}
		return trinity.Strategy{Name: "parallel", Parallelism: 4, Precision: "fp32"}
	}
	if precision >= parallel {
		return trinity.Strategy{Name: "precision", Parallelism: 1, Precision: "fp64"}
	}
	return trinity.Strategy{Name: "parallel", Parallelism: 4, Precision: "fp32"}
}

func (p *MetaPolicy) Update(strategy trinity.Strategy, reward float64) {
	if p == nil {
		return
	}
	if reward < -1 {
		reward = -1
	}
	if reward > 1 {
		reward = 1
	}
	key := "parallelism"
	if strategy.Name == "precision" {
		key = "precision"
	}
	p.mu.Lock()
	if p.values == nil {
		p.values = map[string]float64{"precision": 0.5, "parallelism": 0.5}
	}
	p.values[key] += p.alpha * (reward - p.values[key])
	p.mu.Unlock()
}
