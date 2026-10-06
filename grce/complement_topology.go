package grce

import (
	"fmt"
	"github.com/divibisoul/Orquestrador-/grf"
)

type ComplementPlacement struct {
	ID                 string
	GRCEStages         []string
	PrimaryIntegration string
	Epistemic          grf.EpistemicState
}

func SOUL28ComplementPlacements() []ComplementPlacement {
	participants := NewExternalParticipants()
	return []ComplementPlacement{
		{ID: BijuxDAGRuntimeID, GRCEStages: []string{"0-12"}, PrimaryIntegration: "validated DAG execution, replay and provenance", Epistemic: participants[BijuxDAGRuntimeID].EpistemicState()},
		{ID: OuroLoopID, GRCEStages: []string{"7", "9"}, PrimaryIntegration: "bounded remediation and verification", Epistemic: participants[OuroLoopID].EpistemicState()},
		{ID: RecursID, GRCEStages: []string{"3", "11"}, PrimaryIntegration: "failure localization and experiential memory", Epistemic: participants[RecursID].EpistemicState()},
	}
}

func ValidateComplementTopology() error {
	placements := SOUL28ComplementPlacements()
	if len(placements) != 3 {
		return fmt.Errorf("SOUL28_COMPLEMENT_COUNT:%d", len(placements))
	}
	for _, p := range placements {
		if p.ID == "" || len(p.GRCEStages) == 0 || p.PrimaryIntegration == "" {
			return fmt.Errorf("SOUL28_COMPLEMENT_INCOMPLETE:%s", p.ID)
		}
	}
	return nil
}
