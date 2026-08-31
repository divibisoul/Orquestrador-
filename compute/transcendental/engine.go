package transcendental

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

// Executor is the concrete compute boundary. Hardware-specific backends can
// implement this interface without changing Trinity contracts.
type Executor interface {
	Execute(context.Context, trinity.Workload, trinity.Route) (trinity.Result, error)
}

// Engine is the provider-neutral compute engine. In auto mode it selects a
// real local CPU backend unless an external executor is supplied.
type Engine struct {
	cfg      trinity.ComputeConfig
	executor Executor
}

func NewEngine(cfg trinity.ComputeConfig) *Engine {
	if cfg.Mode == "" {
		cfg.Mode = "auto"
	}
	if cfg.PrecisionFallback == "" {
		cfg.PrecisionFallback = "fp32"
	}
	if cfg.EfficiencyFactor <= 0 || cfg.EfficiencyFactor > 1 {
		cfg.EfficiencyFactor = 0.7
	}
	if cfg.NoiseStd < 0 {
		cfg.NoiseStd = 0
	}
	return &Engine{cfg: cfg, executor: CPUExecutor{}}
}

// WithExecutor replaces the local backend with a real provider/runtime.
// Passing nil restores the local CPU backend.
func (e *Engine) WithExecutor(executor Executor) *Engine {
	if e == nil {
		return e
	}
	if executor == nil {
		e.executor = CPUExecutor{}
	} else {
		e.executor = executor
	}
	return e
}

func (e *Engine) Execute(ctx context.Context, w trinity.Workload, r trinity.Route) (trinity.Result, error) {
	if e == nil {
		return trinity.Result{}, errors.New("nil compute engine")
	}
	if ctx == nil {
		return trinity.Result{}, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return trinity.Result{}, err
	}
	if w.ID == "" || w.Kind == "" {
		return trinity.Result{}, errors.New("invalid workload")
	}
	if r.Model == "" {
		r.Model = "blackwell"
	}
	if r.Provider == "" {
		r.Provider = "transcendental"
	}
	if r.Precision == "" {
		r.Precision = e.cfg.PrecisionFallback
	}
	if e.executor == nil {
		return trinity.Result{}, errors.New("compute executor unavailable")
	}
	return e.executor.Execute(ctx, w, r)
}

// CPUExecutor performs a real deterministic CPU operation over the workload
// payload. It is intentionally not presented as GPU/model inference.
type CPUExecutor struct{}

func (CPUExecutor) Execute(ctx context.Context, w trinity.Workload, r trinity.Route) (trinity.Result, error) {
	if ctx == nil {
		return trinity.Result{}, errors.New("nil context")
	}
	start := time.Now()
	payload := strings.TrimSpace(w.Payload)
	if payload == "" {
		payload = w.ID
	}
	iterations := w.BatchSize
	if iterations < 1 {
		iterations = 1
	}
	if iterations > 4096 {
		iterations = 4096
	}

	var digest [32]byte
	input := []byte(fmt.Sprintf("%s|%s|%d|%d|%s", payload, w.Kind, w.MatrixSize, w.BatchSize, r.Precision))
	for i := 0; i < iterations; i++ {
		if err := ctx.Err(); err != nil {
			return trinity.Result{}, err
		}
		h := sha256.Sum256(input)
		digest = h
		input = h[:]
	}

	trace := ""
	if v, ok := ctx.Value(traceContextKey{}).(string); ok {
		trace = strings.TrimSpace(v)
	}
	metadata := map[string]string{
		"backend":      "cpu",
		"execution":    "deterministic",
		"hardware_model": r.Model,
		"digest":       hex.EncodeToString(digest[:]),
	}
	if trace != "" {
		metadata["trace_id"] = trace
	}
	return trinity.Result{
		Output:    hex.EncodeToString(digest[:]),
		Route:     r,
		LatencyMS: float64(time.Since(start).Microseconds()) / 1000,
		Success:   true,
		Metadata:  metadata,
	}, nil
}

type traceContextKey struct{}

// WithTraceID makes the trace available to the compute boundary without
// coupling the compute package to the orchestrator package.
func WithTraceID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, traceContextKey{}, strings.TrimSpace(id))
}
