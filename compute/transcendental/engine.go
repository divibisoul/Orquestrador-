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

// Estimate supplies a concrete conservative cost model to the PFC. It does
// not claim that a named GPU is physically present; availability belongs to
// the provider/device executor.
func (e *Engine) Estimate(ctx context.Context, w trinity.Workload) (trinity.CostEstimate, error) {
	if e == nil {
		return trinity.CostEstimate{}, errors.New("nil compute engine")
	}
	if ctx == nil {
		return trinity.CostEstimate{}, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return trinity.CostEstimate{}, err
	}
	matrix := w.MatrixSize
	if matrix < 1 {
		matrix = 1
	}
	batch := w.BatchSize
	if batch < 1 {
		batch = 1
	}
	memory := w.MemoryNeeded
	if memory <= 0 {
		memory = float64(matrix*matrix) * 4 / (1024 * 1024)
	}
	latency := (float64(matrix*matrix)/10000 + float64(batch)*5) / e.cfg.EfficiencyFactor
	return trinity.CostEstimate{
		LatencyMS:   latency,
		Memory:      memory,
		ComputeCost: latency / 1000,
		Confidence:  0.8,
	}, nil
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
