package core

import (
	"context"
	"errors"
	"math"
	"time"
)

type PerformanceModel interface {
	Name() string
	EstimateTime(wl Workload) time.Duration
	GetPFLOPS(p Precision) float64
	GetBandwidthGBs() float64
	GetMemoryCapacityGB() int64
	GetMaxParallelUnits() int
	Supports(p Precision) bool
}

type Engine struct{ Config Config }

func NewEngine(cfg Config) (*Engine, error) {
	if cfg.Mode == "" { cfg.Mode = "auto" }
	if cfg.PrecisionFallback == "" { cfg.PrecisionFallback = FP32 }
	if cfg.MaxParallelUnits == 0 { cfg.MaxParallelUnits = MaxSimulationParallelUnits }
	if cfg.Simulation.EfficiencyFactor == 0 { cfg.Simulation.EfficiencyFactor = 0.7 }
	if err := cfg.Validate(); err != nil { return nil, err }
	return &Engine{Config: cfg}, nil
}

func (e *Engine) Estimate(ctx context.Context, wl Workload, model PerformanceModel, effectivePrecision Precision, confidencePenalty float64) (CostEstimate, error) {
	if err := validateContext(ctx); err != nil { return CostEstimate{}, err }
	if model == nil { return CostEstimate{}, errors.New("performance model is required") }
	if wl.ID == "" { return CostEstimate{}, errors.New("workload ID is required") }
	if wl.MatrixSize < 1 { return CostEstimate{}, errors.New("matrix size must be >= 1") }
	if wl.BatchSize < 1 { return CostEstimate{}, errors.New("batch size must be >= 1") }
	if wl.DataBytes < 0 || wl.MemoryNeeded < 0 { return CostEstimate{}, errors.New("data and memory requirements must be non-negative") }
	if effectivePrecision == "" { effectivePrecision = wl.Precision }
	if !model.Supports(effectivePrecision) { return CostEstimate{}, errors.New("model does not support requested precision") }

	t := model.EstimateTime(Workload{ID: wl.ID, Operation: wl.Operation, Precision: effectivePrecision, MatrixSize: wl.MatrixSize, BatchSize: wl.BatchSize, DataBytes: wl.DataBytes, MemoryNeeded: wl.MemoryNeeded, Priority: wl.Priority, Metadata: wl.Metadata})
	if t <= 0 { return CostEstimate{}, errors.New("model returned non-positive estimate") }
	confidence := 0.90 - math.Max(0, math.Min(0.50, confidencePenalty))
	return CostEstimate{EstimatedTime: t, EstimatedPFLOPS: model.GetPFLOPS(effectivePrecision), EstimatedMemoryGB: float64(max64(wl.MemoryNeeded,0)), EstimatedEnergy: t.Seconds() * math.Max(model.GetPFLOPS(effectivePrecision), 0), Architecture: model.Name(), Confidence: confidence}, nil
}

func (e *Engine) Metrics(wl Workload, model PerformanceModel, effectivePrecision Precision) (Metrics, error) {
	if model == nil { return Metrics{}, errors.New("performance model is required") }
	if effectivePrecision == "" { effectivePrecision = wl.Precision }
	if !model.Supports(effectivePrecision) { return Metrics{}, errors.New("model does not support requested precision") }
	t := model.EstimateTime(Workload{ID: wl.ID, Operation: wl.Operation, Precision: effectivePrecision, MatrixSize: wl.MatrixSize, BatchSize: wl.BatchSize, DataBytes: wl.DataBytes, MemoryNeeded: wl.MemoryNeeded, Priority: wl.Priority, Metadata: wl.Metadata})
	if t <= 0 { return Metrics{}, errors.New("model returned non-positive estimate") }
	seconds := t.Seconds()
	bandwidth := float64(max64(wl.DataBytes,0)) / seconds / 1e9
	return Metrics{Architecture:model.Name(), EffectivePFLOPS:model.GetPFLOPS(effectivePrecision), BandwidthUsedGBs:bandwidth, LatencyMs:t.Seconds()*1000, MemoryUsedGB:float64(max64(wl.MemoryNeeded,0)), Efficiency:e.Config.Simulation.EfficiencyFactor, Timestamp:time.Now().UTC()}, nil
}

func validateContext(ctx context.Context) error { if ctx == nil { return errors.New("context is nil") }; select { case <-ctx.Done(): return ctx.Err(); default: return nil } }
func max64(a,b int64) int64 { if a>b { return a }; return b }
