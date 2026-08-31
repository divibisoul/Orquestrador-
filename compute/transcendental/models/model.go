package models

import (
	"time"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type PerformanceModel = core.PerformanceModel

type Catalog struct {
	Blackwell BlackwellB200
	VeraRubin VeraRubin
	CDNA5 CDNA5
	Trillium Trillium
	Atlas Atlas950
}

func DefaultCatalog(cfg core.Config) []PerformanceModel {
	return []PerformanceModel{
		BlackwellB200{Efficiency:cfg.Simulation.EfficiencyFactor, Sparse:cfg.Simulation.UseSparsity},
		VeraRubin{Efficiency:cfg.Simulation.EfficiencyFactor},
		CDNA5{Efficiency:cfg.Simulation.EfficiencyFactor},
		Trillium{Efficiency:cfg.Simulation.EfficiencyFactor},
		Atlas950{Efficiency:cfg.Simulation.EfficiencyFactor, ScaleFactor:1},
	}
}

var _ PerformanceModel = BlackwellB200{}
var _ PerformanceModel = VeraRubin{}
var _ PerformanceModel = CDNA5{}
var _ PerformanceModel = Trillium{}
var _ PerformanceModel = Atlas950{}

var _ = time.Second
