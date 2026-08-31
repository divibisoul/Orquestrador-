package prefrontal

import "github.com/divibisoul/Orquestrador-/core/trinity"

type ConflictMonitor struct{ sensitivity float64 }

func NewConflictMonitor(sensitivity float64) *ConflictMonitor {
	if sensitivity <= 0 { sensitivity = 0.5 }
	if sensitivity > 1 { sensitivity = 1 }
	return &ConflictMonitor{sensitivity: sensitivity}
}

func (m *ConflictMonitor) Score(w trinity.Workload, e trinity.CostEstimate, history float64) float64 {
	ambiguity := 0.0
	if w.ID == "" || w.Kind == "" { ambiguity += 0.5 }
	if w.MatrixSize <= 0 && w.BatchSize <= 0 && w.MemoryNeeded <= 0 { ambiguity += 0.25 }
	if e.Confidence < 1 { ambiguity += (1 - e.Confidence) * 0.25 }
	if history < 0 { history = 0 }
	if history > 1 { history = 1 }
	costRisk := e.ComputeCost / (1 + e.ComputeCost)
	score := m.sensitivity*(0.45*costRisk+0.35*history+0.20*ambiguity)
	if score > 1 { return 1 }
	return score
}
