package executor

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/interfaces"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/selection"
)

type SimulatedExecutor struct {
	Engine      *core.Engine
	Catalog     []models.PerformanceModel
	Mode        string
	HistorySize int
	mu          sync.RWMutex
	history     []core.Metrics
}

var _ interfaces.ComputeBackend = (*SimulatedExecutor)(nil)
var _ interfaces.MetricsProvider = (*SimulatedExecutor)(nil)
var _ interfaces.CostEstimator = (*SimulatedExecutor)(nil)

func New(engine *core.Engine, catalog []models.PerformanceModel, mode string) (*SimulatedExecutor, error) {
	if engine == nil {
		return nil, errors.New("engine is required")
	}
	if err := engine.Config.Validate(); err != nil {
		return nil, err
	}
	if len(catalog) == 0 {
		return nil, errors.New("model catalog is empty")
	}
	if mode == "" {
		mode = engine.Config.Mode
	}
	return &SimulatedExecutor{Engine: engine, Catalog: catalog, Mode: mode, HistorySize: 256}, nil
}

func (e *SimulatedExecutor) Estimate(ctx context.Context, wl core.Workload) (core.CostEstimate, error) {
	if e == nil || e.Engine == nil {
		return core.CostEstimate{}, errors.New("executor is not initialized")
	}
	return e.estimateWithStrategy(ctx, wl, "auto")
}

func (e *SimulatedExecutor) estimateWithStrategy(ctx context.Context, wl core.Workload, strategy string) (core.CostEstimate, error) {
	if e == nil || e.Engine == nil {
		return core.CostEstimate{}, errors.New("executor is not initialized")
	}
	if !e.Engine.Config.Enabled {
		return core.CostEstimate{}, errors.New("transcendental compute engine is disabled")
	}
	if err := ctxErr(ctx); err != nil {
		return core.CostEstimate{}, err
	}
	if err := wl.Validate(); err != nil {
		return core.CostEstimate{}, err
	}
	sel, err := selection.Select(wl, e.Catalog, e.Mode, strategy, e.Engine.Config.PrecisionFallback)
	if err != nil {
		return core.CostEstimate{}, err
	}
	return e.Engine.Estimate(ctx, wl, sel.Model, sel.EffectivePrecision, sel.Penalty)
}

func (e *SimulatedExecutor) EstimateCost(ctx context.Context, plan core.Plan) (core.CostEstimate, error) {
	if e == nil || e.Engine == nil {
		return core.CostEstimate{}, errors.New("executor is not initialized")
	}
	if err := ctxErr(ctx); err != nil {
		return core.CostEstimate{}, err
	}
	if err := plan.Validate(); err != nil {
		return core.CostEstimate{}, err
	}
	var total time.Duration
	var weightedPF, mem, energy, confidence float64
	architecture := ""
	for _, wl := range plan.Workloads {
		ce, err := e.estimateWithStrategy(ctx, wl, plan.Strategy)
		if err != nil {
			return core.CostEstimate{}, err
		}
		total += ce.EstimatedTime
		weightedPF += ce.EstimatedPFLOPS
		mem += ce.EstimatedMemoryGB
		energy += ce.EstimatedEnergy
		confidence += ce.Confidence
		if architecture == "" {
			architecture = ce.Architecture
		} else if architecture != ce.Architecture {
			architecture = "mixed"
		}
	}
	count := float64(len(plan.Workloads))
	return core.CostEstimate{
		EstimatedTime:     total,
		EstimatedPFLOPS:   weightedPF / count,
		EstimatedMemoryGB: mem,
		EstimatedEnergy:   energy,
		Architecture:      architecture,
		Confidence:        confidence / count,
	}, nil
}

func (e *SimulatedExecutor) Execute(ctx context.Context, wl core.Workload) (core.Result, error) {
	if e == nil || e.Engine == nil {
		err := errors.New("executor is not initialized")
		return core.Result{WorkloadID: wl.ID, Error: err}, err
	}
	if !e.Engine.Config.Enabled {
		err := errors.New("transcendental compute engine is disabled")
		return core.Result{WorkloadID: wl.ID, Error: err}, err
	}
	if err := ctxErr(ctx); err != nil {
		return core.Result{WorkloadID: wl.ID, Error: err}, err
	}
	if err := wl.Validate(); err != nil {
		return core.Result{WorkloadID: wl.ID, Error: err}, err
	}
	sel, err := selection.Select(wl, e.Catalog, e.Mode, "auto", e.Engine.Config.PrecisionFallback)
	if err != nil {
		return core.Result{WorkloadID: wl.ID, Error: err}, err
	}
	metrics, err := e.Engine.Metrics(wl, sel.Model, sel.EffectivePrecision)
	if err != nil {
		return core.Result{WorkloadID: wl.ID, Error: err}, err
	}
	e.push(metrics)
	return core.Result{WorkloadID: wl.ID, Metrics: metrics, Data: []byte("simulated-result")}, nil
}

// ExecutePlan runs independent simulated workloads in parallel using a bounded worker pool.
// The configured value can describe a larger hardware model, but runtime workers are capped at 1000.
// Individual execution errors are retained in each Result instead of being discarded.
func (e *SimulatedExecutor) ExecutePlan(ctx context.Context, workloads []core.Workload) []core.Result {
	results := make([]core.Result, len(workloads))
	if len(workloads) == 0 {
		return results
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if e == nil || e.Engine == nil {
		err := errors.New("executor is not initialized")
		for i, workload := range workloads {
			results[i] = core.Result{WorkloadID: workload.ID, Error: err}
		}
		return results
	}
	workers := e.Engine.Config.EffectiveParallelUnits()
	if workers > len(workloads) {
		workers = len(workloads)
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				result, err := e.Execute(ctx, workloads[idx])
				if err != nil {
					result.Error = err
				}
				results[idx] = result
			}
		}()
	}
	for i := range workloads {
		select {
		case <-ctx.Done():
			for j := i; j < len(workloads); j++ {
				results[j] = core.Result{WorkloadID: workloads[j].ID, Error: ctx.Err()}
			}
			i = len(workloads)
		case jobs <- i:
		}
		if i >= len(workloads)-1 || ctx.Err() != nil {
			break
		}
	}
	close(jobs)
	wg.Wait()
	return results
}

func (e *SimulatedExecutor) GetLastMetrics() core.Metrics {
	if e == nil {
		return core.Metrics{}
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.history) == 0 {
		return core.Metrics{}
	}
	return e.history[len(e.history)-1]
}

func (e *SimulatedExecutor) GetMetricsHistory(limit int) []core.Metrics {
	if e == nil {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if limit <= 0 || limit > len(e.history) {
		limit = len(e.history)
	}
	out := make([]core.Metrics, limit)
	copy(out, e.history[len(e.history)-limit:])
	return out
}

func (e *SimulatedExecutor) push(m core.Metrics) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.HistorySize < 1 {
		e.HistorySize = 256
	}
	e.history = append(e.history, m)
	if len(e.history) > e.HistorySize {
		e.history = e.history[len(e.history)-e.HistorySize:]
	}
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
