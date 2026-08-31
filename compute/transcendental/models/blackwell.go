package models

import (
	"math"
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type BlackwellB200 struct{ Efficiency float64; Sparse bool }

func (m BlackwellB200) Name() string { return "NVIDIA Blackwell B200" }
func (m BlackwellB200) GetBandwidthGBs() float64 { return 8000 }
func (m BlackwellB200) GetMemoryCapacityGB() int64 { return 192 }
func (m BlackwellB200) GetMaxParallelUnits() int { return 8192 }
func (m BlackwellB200) Supports(p core.Precision) bool { switch p { case core.FP4,core.FP8,core.FP16,core.BF16,core.FP32,core.FP64: return true }; return false }
func (m BlackwellB200) GetPFLOPS(p core.Precision) float64 { base:=map[core.Precision]float64{core.FP4:9,core.FP8:4.5,core.FP16:2.25,core.BF16:2.25,core.FP32:0.080,core.FP64:0.040}[p]; if m.Sparse && base>0 { return base*2 }; return base }
func (m BlackwellB200) EstimateTime(wl core.Workload) time.Duration { return estimateRoofline(wl,m,eff(m.Efficiency)) }

func estimateRoofline(wl core.Workload,m PerformanceModel,efficiency float64) time.Duration {
	if efficiency<=0 || efficiency>1 { efficiency=.7 }
	units:=wl.MatrixSize; if units<1 { units=1 }
	batch:=wl.BatchSize; if batch<1 { batch=1 }
	flops:=2*float64(units)*float64(units)*float64(batch)
	pflops:=m.GetPFLOPS(wl.Precision)
	if pflops<=0 { return 0 }
	compute:=flops/(pflops*1e15)
	memory:=float64(max64(wl.DataBytes,0))/(m.GetBandwidthGBs()*1e9)
	seconds:=math.Max(compute,memory)/efficiency
	return time.Duration(seconds*float64(time.Second))
}
func eff(v float64) float64 { if v==0 { return .7 }; return v }
func max64(a,b int64) int64 { if a>b{return a};return b }
