package transcendental

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type Engine struct { cfg trinity.ComputeConfig }

func NewEngine(cfg trinity.ComputeConfig) *Engine {
	if cfg.Mode == "" { cfg.Mode = "auto" }
	if cfg.PrecisionFallback == "" { cfg.PrecisionFallback = "fp32" }
	if cfg.EfficiencyFactor <= 0 || cfg.EfficiencyFactor > 1 { cfg.EfficiencyFactor = 0.7 }
	if cfg.NoiseStd < 0 { cfg.NoiseStd = 0 }
	return &Engine{cfg:cfg}
}

func (e *Engine) Execute(ctx context.Context, w trinity.Workload, r trinity.Route) (trinity.Result,error) {
	if e == nil { return trinity.Result{}, errors.New("nil compute engine") }
	if ctx == nil { return trinity.Result{}, errors.New("nil context") }
	if err:=ctx.Err(); err!=nil { return trinity.Result{},err }
	start:=time.Now()
	precision:=r.Capabilities
	_ = precision
	if r.Model=="" { r.Model="blackwell" }
	if r.Provider=="" { r.Provider="transcendental" }
	work:=float64(max(1,w.MatrixSize*w.MatrixSize))/1000000 + float64(max(1,w.BatchSize))/64
	latency:=work*100/(e.cfg.EfficiencyFactor)
	if latency < 1 { latency=1 }
	result:=trinity.Result{Success:true,Route:r,LatencyMS:float64(time.Since(start).Microseconds())/1000,Metadata:map[string]string{"mode":e.cfg.Mode,"simulated":"true"}}
	result.LatencyMS += latency
	result.Output = "simulated inference completed"
	return result,nil
}

func max(a,b int) int { if a>b{return a};return b }
func sanitize(x float64) float64 { if math.IsNaN(x)||math.IsInf(x,0){return 0};return x }
