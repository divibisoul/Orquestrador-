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

type Executor interface {
	Execute(context.Context, trinity.Workload, trinity.Route) (trinity.Result, error)
}

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

// WithExecutor installs a concrete hardware/provider runtime. A nil executor
// restores the real deterministic CPU backend.
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

// CPUExecutor performs a real deterministic CPU operation. It is deliberately
// not advertised as GPU execution or LLM inference.
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

	metadata := map[string]string{
		"backend":        "cpu",
		"execution":      "deterministic",
		"hardware_model": r.Model,
		"digest":         hex.EncodeToString(digest[:]),
	}
	if trace := trinity.TraceID(ctx); trace != "" {
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
