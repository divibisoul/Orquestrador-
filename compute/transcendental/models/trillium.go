package models

import (
	"time"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type Trillium struct{ Efficiency float64 }
func (m Trillium) Name() string { return "Google TPU v6e Trillium" }
func (m Trillium) GetBandwidthGBs() float64 { return 1640 }
func (m Trillium) GetMemoryCapacityGB() int64 { return 32 }
func (m Trillium) GetMaxParallelUnits() int { return 256 }
func (m Trillium) Supports(p core.Precision) bool { return p==core.INT8||p==core.BF16 }
func (m Trillium) GetPFLOPS(p core.Precision) float64 { return map[core.Precision]float64{core.INT8:0.001836,core.BF16:0.918}[p] }
func (m Trillium) EstimateTime(wl core.Workload) time.Duration { return estimateRoofline(wl,m,eff(m.Efficiency)) }
