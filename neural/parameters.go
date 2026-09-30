package neural

import "time"

// Parameters exposes the active neural configuration in the canonical field names.
// Canonical source is this file; callers must not infer a second implementation.
// consumed by the N02->N07 bridge. It is descriptive runtime state, not a claim of
// model quality or external inference capability.
func (n *Network) Parameters() map[string]any {
	if n == nil {
		return map[string]any{"status": "unavailable"}
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	layers := make([]map[string]any, len(n.config.Layers))
	for i, layer := range n.config.Layers {
		layers[i] = map[string]any{
			"activation":   layer.Activation,
			"dropout_rate": layer.DropoutRate,
		}
	}
	updated := any(nil)
	if !n.stats.LastUpdate.IsZero() {
		updated = n.stats.LastUpdate.UTC().Format(time.RFC3339Nano)
	}
	return map[string]any{
		"size":           n.size,
		"learning_rate":  n.learningRate,
		"optimizer":      n.config.Optimizer,
		"regularization": n.config.Regularization,
		"gradient_clip":  n.config.GradientClip,
		"heads":          n.config.Heads,
		"batch_cache":    n.config.BatchCache,
		"layers":         layers,
		"learning_steps": n.stats.LearningSteps,
		"last_update":    updated,
	}
}
