package models

import (
	"time"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type CDNA5 struct{ Efficiency float64 }
func (m CDNA5) Name() string { return "AMD Instinct MI400 / CDNA5" }
func (m CDNA5) GetBandwidthGBs() float64 { return 23000 }
func (m CDNA5) GetMemoryCapacityGB() int64 { return 432 }
func (m CDNA5) GetMaxParallelUnits() int { return 8192 }
func (m CDNA5) Supports(p core.Precision) bool { return p==core.FP4||p==core.FP8||p==core.FP16||p==core.BF16 }
func (m CDNA5) GetPFLOPS(p core.Precision) float64 { return map[core.Precision]float64{core.FP4:40,core.FP8:20,core.FP16:7.5,core.BF16:7.5}[p] }
func (m CDNA5) EstimateTime(wl core.Workload) time.Duration { return estimateRoofline(wl,m,eff(m.Efficiency)) }
