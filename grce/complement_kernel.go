package grce

import "github.com/divibisoul/Orquestrador-/grf"

type ComplementRole string

const (
	ComplementExecutionKernel    ComplementRole = "execution-kernel"
	ComplementSelfHealing        ComplementRole = "self-healing"
	ComplementExperientialMemory ComplementRole = "experiential-memory"
)

type ComplementBinding struct {
	ID         string
	Role       ComplementRole
	GRCEStages []string
	Source     string
	Revision   string
	State      grf.EpistemicState
}

func SOUL28Bindings() []ComplementBinding {
	p := NewExternalParticipants()
	return []ComplementBinding{
		{ID: BijuxDAGRuntimeID, Role: ComplementExecutionKernel, GRCEStages: []string{"0-12"}, Source: p[BijuxDAGRuntimeID].Config.Source, Revision: p[BijuxDAGRuntimeID].Config.Revision, State: p[BijuxDAGRuntimeID].EpistemicState()},
		{ID: OuroLoopID, Role: ComplementSelfHealing, GRCEStages: []string{"7", "9"}, Source: p[OuroLoopID].Config.Source, Revision: p[OuroLoopID].Config.Revision, State: p[OuroLoopID].EpistemicState()},
		{ID: RecursID, Role: ComplementExperientialMemory, GRCEStages: []string{"3", "11"}, Source: p[RecursID].Config.Source, Revision: p[RecursID].Config.Revision, State: p[RecursID].EpistemicState()},
	}
}
