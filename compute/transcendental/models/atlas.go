package models

import (
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"math"
	"time"
)

// Atlas 950 is modeled as a scale-out super-node. The figures are preliminary/reference
// estimates and must be replaced by official published specifications when available.
type Atlas950 struct {
	Efficiency  float64
	ScaleFactor int
}

func (m Atlas950) Name() string               { return "Huawei Atlas 950 (scale-out reference)" }
func (m Atlas950) GetBandwidthGBs() float64   { return 16000 * float64(maxInt(m.ScaleFactor, 1)) }
func (m Atlas950) GetMemoryCapacityGB() int64 { return 128 * int64(maxInt(m.ScaleFactor, 1)) }
func (m Atlas950) GetMaxParallelUnits() int   { return 1024 * maxInt(m.ScaleFactor, 1) }
func (m Atlas950) Supports(p core.Precision) bool {
	return p == core.FP4 || p == core.FP8 || p == core.FP16 || p == core.BF16 || p == core.FP32 || p == core.FP64 || p == core.INT8
}
func (m Atlas950) GetPFLOPS(p core.Precision) float64 {
	base := map[core.Precision]float64{core.FP4: 20, core.FP8: 10, core.FP16: 5, core.BF16: 5, core.FP32: 0.1, core.FP64: 0.05, core.INT8: 20}[p]
	return base * float64(maxInt(m.ScaleFactor, 1))
}
func (m Atlas950) EstimateTime(wl core.Workload) time.Duration {
	return estimateRoofline(wl, m, atlasEff(m.Efficiency))
}
func atlasEff(v float64) float64 {
	if v <= 0 || v > 1 {
		return .7
	}
	return math.Min(v, 1)
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
