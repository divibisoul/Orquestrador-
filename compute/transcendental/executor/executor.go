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
	Engine   *core.Engine
	Catalog  []models.PerformanceModel
	Mode     string
	HistorySize int
	mu sync.RWMutex
	history []core.Metrics
}

var _ interfaces.ComputeBackend = (*SimulatedExecutor)(nil)
var _ interfaces.MetricsProvider = (*SimulatedExecutor)(nil)
var _ interfaces.CostEstimator = (*SimulatedExecutor)(nil)

func New(engine *core.Engine, catalog []models.PerformanceModel, mode string) (*SimulatedExecutor,error) {
	if engine==nil { return nil, errors.New("engine is required") }
	if err:=engine.Config.Validate(); err!=nil { return nil,err }
	if len(catalog)==0 { return nil, errors.New("model catalog is empty") }
	if mode=="" { mode=engine.Config.Mode }
	return &SimulatedExecutor{Engine:engine,Catalog:catalog,Mode:mode,HistorySize:256},nil
}

func (e *SimulatedExecutor) Estimate(ctx context.Context, wl core.Workload) (core.CostEstimate,error) {
	if !e.Engine.Config.Enabled { return core.CostEstimate{}, errors.New("transcendental compute engine is disabled") }
	if ctx==nil { return core.CostEstimate{},errors.New("context is nil") }
	sel,err:=selection.Select(wl,e.Catalog,e.Mode,e.Engine.Config.PrecisionFallback); if err!=nil{return core.CostEstimate{},err}
	return e.Engine.Estimate(ctx,wl,sel.Model,sel.EffectivePrecision,sel.Penalty)
}

func (e *SimulatedExecutor) EstimateCost(ctx context.Context, plan core.Plan) (core.CostEstimate,error) {
	if err:=ctxErr(ctx);err!=nil{return core.CostEstimate{},err}
	if len(plan.Workloads)==0{return core.CostEstimate{},errors.New("plan contains no workloads")}
	var total time.Duration; var weightedPF,mem,energy,confidence float64
	for _,wl:=range plan.Workloads { ce,err:=e.Estimate(ctx,wl); if err!=nil{return core.CostEstimate{},err}; total+=ce.EstimatedTime; weightedPF+=ce.EstimatedPFLOPS; mem+=ce.EstimatedMemoryGB; energy+=ce.EstimatedEnergy; confidence+=ce.Confidence }
	first,err:=e.Estimate(ctx,plan.Workloads[0]);if err!=nil{return core.CostEstimate{},err}
	return core.CostEstimate{EstimatedTime:total,EstimatedPFLOPS:weightedPF/float64(len(plan.Workloads)),EstimatedMemoryGB:mem,EstimatedEnergy:energy,Architecture: "plan:"+first.Architecture,Confidence:confidence/float64(len(plan.Workloads))},nil
}

func (e *SimulatedExecutor) Execute(ctx context.Context, wl core.Workload) (core.Result,error) {
	if !e.Engine.Config.Enabled { err:=errors.New("transcendental compute engine is disabled"); return core.Result{WorkloadID:wl.ID,Error:err},err }
	if err:=ctxErr(ctx);err!=nil{return core.Result{WorkloadID:wl.ID,Error:err},err}
	sel,err:=selection.Select(wl,e.Catalog,e.Mode,e.Engine.Config.PrecisionFallback); if err!=nil{return core.Result{WorkloadID:wl.ID,Error:err},err}
	metrics,err:=e.Engine.Metrics(wl,sel.Model,sel.EffectivePrecision);if err!=nil{return core.Result{WorkloadID:wl.ID,Error:err},err}
	metrics.Timestamp=time.Now().UTC()
	e.push(metrics)
	result:=core.Result{WorkloadID:wl.ID,Metrics:metrics,Data:[]byte("simulated-result")}
	return result,nil
}

func (e *SimulatedExecutor) GetLastMetrics() core.Metrics { e.mu.RLock();defer e.mu.RUnlock();if len(e.history)==0{return core.Metrics{}};return e.history[len(e.history)-1] }
func (e *SimulatedExecutor) GetMetricsHistory(limit int) []core.Metrics { e.mu.RLock();defer e.mu.RUnlock();if limit<=0||limit>len(e.history){limit=len(e.history)};out:=make([]core.Metrics,limit);copy(out,e.history[len(e.history)-limit:]);return out }
func (e *SimulatedExecutor) push(m core.Metrics) { e.mu.Lock();defer e.mu.Unlock(); if e.HistorySize<1{e.HistorySize=256}; e.history=append(e.history,m);if len(e.history)>e.HistorySize{e.history=e.history[len(e.history)-e.HistorySize:]} }
func ctxErr(ctx context.Context) error { if ctx==nil{return errors.New("context is nil")};select{case <-ctx.Done():return ctx.Err();default:return nil} }
