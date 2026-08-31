package models

import (
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type PerformanceModel interface {
	Name() string
	EstimateTime(wl core.Workload) time.Duration
	GetPFLOPS(p core.Precision) float64
	GetBandwidthGBs() float64
	GetMemoryCapacityGB() int64
	GetMaxParallelUnits() int
	Supports(p core.Precision) bool
}
