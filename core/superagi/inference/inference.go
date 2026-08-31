package inference

import (
	"context"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

// Engine is a thin adapter over the canonical SuperAGI runtime.
// It deliberately does not duplicate provider or memory implementations.
type Engine struct { Runtime *superagi.Runtime }

func New(r *superagi.Runtime) *Engine {
	if r == nil { r = superagi.NewRuntime() }
	return &Engine{Runtime:r}
}

func (e *Engine) Generate(ctx context.Context, model, input string) (string, error) {
	if e == nil || e.Runtime == nil { return "", errors.New("inference runtime unavailable") }
	model = strings.TrimSpace(model)
	if model == "" { return "", errors.New("model required") }
	return e.Runtime.Inference(ctx, model, input)
}

func (e *Engine) Batch(ctx context.Context, model string, inputs []string) ([]string, error) {
	if e == nil || e.Runtime == nil { return nil, errors.New("inference runtime unavailable") }
	return e.Runtime.BatchInference(ctx, strings.TrimSpace(model), inputs)
}

func (e *Engine) EstimateCost(tokens int) map[string]float64 {
	if e == nil || e.Runtime == nil { return map[string]float64{} }
	return e.Runtime.EstimateCost(tokens)
}

func (e *Engine) Drift(reference, current []float64) float64 {
	if e == nil || e.Runtime == nil { return 0 }
	return e.Runtime.MonitorDrift(reference, current)
}
