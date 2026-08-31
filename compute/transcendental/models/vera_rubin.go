package models

import (
	"time"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

// Vera Rubin figures are preliminary 2026 marketing/reference figures and must be updated when official specifications are published.
type VeraRubin struct{ Efficiency float64 }
func (m VeraRubin) Name() string { return "NVIDIA Vera Rubin (preliminary 2026)" }
func (m VeraRubin) GetBandwidthGBs() float64 { return 22000 }
func (m VeraRubin) GetMemoryCapacityGB() int64 { return 288 }
func (m VeraRubin) GetMaxParallelUnits() int { return 8192 }
func (m VeraRubin) Supports(p core.Precision) bool { return p==core.FP4||p==core.FP8||p==core.FP16||p==core.BF16 }
func (m VeraRubin) GetPFLOPS(p core.Precision) float64 { return map[core.Precision]float64{core.FP4:50,core.FP8:25,core.FP16:11,core.BF16:11}[p] }
func (m VeraRubin) EstimateTime(wl core.Workload) time.Duration { return estimateRoofline(wl,m,eff(m.Efficiency)) }
