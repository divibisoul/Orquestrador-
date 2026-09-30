package prefrontal

import (
	"context"
	"errors"
	"math"
)

// NeuralSignalProvider is the minimal neural-network boundary required by the
// Prefrontal Neocortex. The concrete implementation remains owned by package neural.
type NeuralSignalProvider interface {
	Forward(context.Context, []float64) ([]float64, error)
}

// Neocortex is the canonical executive boundary: neural signal -> candidate ->
// safety inhibition -> committed decision. It never bypasses Cortex policy.
type Neocortex struct {
	cortex *Cortex
	neural NeuralSignalProvider
}

func NewNeocortex(c *Cortex, n NeuralSignalProvider) (*Neocortex, error) {
	if c == nil {
		return nil, errors.New("cortex is required")
	}
	if n == nil {
		return nil, errors.New("neural signal provider is required")
	}
	return &Neocortex{cortex: c, neural: n}, nil
}

func (n *Neocortex) Evaluate(ctx context.Context, id string, input []float64, risk, cost, urgency, impact float64) (Candidate, error) {
	if ctx == nil {
		return Candidate{}, errors.New("context is nil")
	}
	if id == "" {
		return Candidate{}, errors.New("candidate id is required")
	}
	signal, err := n.neural.Forward(ctx, input)
	if err != nil {
		return Candidate{}, err
	}
	utility := 0.0
	mean := 0.0
	for _, value := range signal {
		utility += 1 / (1 + math.Exp(-value))
		mean += value
	}
	variance := 0.0
	if len(signal) > 0 {
		mean /= float64(len(signal))
		utility /= float64(len(signal))
		for _, value := range signal {
			delta := value - mean
			variance += delta * delta
		}
		variance /= float64(len(signal))
	}
	stddev := math.Sqrt(math.Max(0, variance))
	uncertainty := stddev / (1 + stddev)
	candidate := Candidate{ID: id, Utility: utility, Risk: risk, Cost: cost, Urgency: urgency, Impact: impact, Uncertainty: uncertainty, Context: map[string]any{"neural_dimensions": len(signal), "signal_mean": mean, "signal_stddev": stddev, "uncertainty_semantics": "bounded-dispersion-proxy-not-probability"}}
	if err := n.cortex.ValidateAction(candidate); err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}

func (n *Neocortex) Commit(candidate Candidate, reason string) (Decision, error) {
	if n == nil || n.cortex == nil {
		return Decision{}, errors.New("neocortex unavailable")
	}
	return n.cortex.Commit(candidate, reason)
}

func (n *Neocortex) Health() map[string]any {
	if n == nil || n.cortex == nil {
		return map[string]any{"status": "degraded", "error": "neocortex unavailable"}
	}
	h := n.cortex.Health()
	h["module"] = "PrefrontalNeocortex"
	h["neural_binding"] = n.neural != nil
	h["role"] = "planning-decision-inhibition-global-synthesis"
	return h
}
