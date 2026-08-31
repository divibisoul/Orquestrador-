package transcendental

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type Executor interface {
	Execute(context.Context, trinity.Workload, trinity.Route) (trinity.Result, error)
}

type Engine struct {
	mu       sync.RWMutex
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
	if cfg.NoiseStd < 0 || math.IsNaN(cfg.NoiseStd) || math.IsInf(cfg.NoiseStd, 0) {
		cfg.NoiseStd = 0
	}
	return &Engine{cfg: cfg, executor: CPUExecutor{}}
}

func (e *Engine) WithExecutor(executor Executor) *Engine {
	if e == nil {
		return e
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if executor == nil {
		e.executor = CPUExecutor{}
	} else {
		e.executor = executor
	}
	return e
}

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
	if w.MatrixSize < 0 || w.BatchSize < 0 || w.MemoryNeeded < 0 {
		return trinity.CostEstimate{}, errors.New("invalid workload resource values")
	}
	matrix := w.MatrixSize
	if matrix < 1 {
		matrix = 1
	}
	batch := w.BatchSize
	if batch < 1 {
		batch = 1
	}
	e.mu.RLock()
	efficiency := e.cfg.EfficiencyFactor
	e.mu.RUnlock()
	memory := w.MemoryNeeded
	if memory <= 0 {
		memory = float64(matrix) * float64(matrix) * 4 / (1024 * 1024)
	}
	latency := (float64(matrix)*float64(matrix)/10000 + float64(batch)*5) / efficiency
	return trinity.CostEstimate{LatencyMS: latency, Memory: memory, ComputeCost: latency / 1000, Confidence: 0.8}, nil
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
	if strings.TrimSpace(w.ID) == "" || strings.TrimSpace(w.Kind) == "" {
		return trinity.Result{}, errors.New("invalid workload")
	}
	if w.MatrixSize < 0 || w.BatchSize < 0 || w.MemoryNeeded < 0 {
		return trinity.Result{}, errors.New("invalid workload resource values")
	}
	e.mu.RLock()
	executor := e.executor
	precisionFallback := e.cfg.PrecisionFallback
	e.mu.RUnlock()
	if r.Model == "" {
		r.Model = "blackwell"
	}
	if r.Provider == "" {
		r.Provider = "transcendental"
	}
	if executor == nil {
		return trinity.Result{}, errors.New("compute executor unavailable")
	}
	if strings.TrimSpace(w.Precision) == "" {
		w.Precision = precisionFallback
	}
	return executor.Execute(ctx, w, r)
}

type CPUExecutor struct{}

func (CPUExecutor) Execute(ctx context.Context, w trinity.Workload, r trinity.Route) (trinity.Result, error) {
	if ctx == nil {
		return trinity.Result{}, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return trinity.Result{}, err
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
	input := []byte(fmt.Sprintf("%s|%s|%d|%d|%s", payload, w.Kind, w.MatrixSize, w.BatchSize, w.Precision))
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
	return trinity.Result{Output: hex.EncodeToString(digest[:]), Route: r, LatencyMS: float64(time.Since(start).Microseconds()) / 1000, Success: true, Metadata: metadata}, nil
}
