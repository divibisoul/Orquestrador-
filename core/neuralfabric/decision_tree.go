package neuralfabric

import (
	"sync"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type DecisionTree struct {
	mu      sync.RWMutex
	depth   int
	weights map[string]float64
}

func NewDecisionTree(depth int) *DecisionTree {
	if depth < 1 {
		depth = 6
	}
	return &DecisionTree{
		depth: depth,
		weights: map[string]float64{
			"blackwell":  0.5,
			"vera_rubin": 0.5,
			"mi400":      0.5,
			"atlas":      0.5,
			"trillium":   0.5,
		},
	}
}

// Select preserves the explicit workload heuristics as the primary rule.
// Learned weights can override them only after a material observed advantage,
// preventing an untrained fabric from randomly changing routes.
func (t *DecisionTree) Select(w trinity.Workload) string {
	if t == nil {
		return "blackwell"
	}
	candidate := "blackwell"
	switch {
	case w.Precision == "fp64":
		candidate = "blackwell"
	case w.MemoryNeeded > 400:
		candidate = "mi400"
	case w.MatrixSize > 4096:
		candidate = "vera_rubin"
	case w.MatrixSize > 0 && w.MatrixSize < 1024:
		candidate = "trillium"
	case w.BatchSize > 64:
		candidate = "blackwell"
	}

	t.mu.RLock()
	best := candidate
	bestScore := t.weights[candidate]
	for name, score := range t.weights {
		if name == candidate {
			continue
		}
		if score > bestScore+0.10 {
			best = name
			bestScore = score
		}
	}
	t.mu.RUnlock()
	return best
}

func (t *DecisionTree) Update(route string, reward, lr float64) {
	if t == nil || route == "" {
		return
	}
	if lr <= 0 {
		lr = 0.01
	}
	if reward < -1 {
		reward = -1
	}
	if reward > 1 {
		reward = 1
	}

	t.mu.Lock()
	if _, ok := t.weights[route]; !ok {
		t.weights[route] = 0.5
	}
	t.weights[route] += lr * (reward - t.weights[route])
	t.mu.Unlock()
}
