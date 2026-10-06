package grce

import "github.com/divibisoul/Orquestrador-/grf"

type BijuxDAGParticipant struct{ *ExternalParticipant }
type OuroLoopParticipant struct{ *ExternalParticipant }
type RecursParticipant struct{ *ExternalParticipant }

func NewSOUL28Participants() map[string]grf.GoldenRuleParticipant {
	p := NewExternalParticipants()
	return map[string]grf.GoldenRuleParticipant{
		BijuxDAGRuntimeID: &BijuxDAGParticipant{ExternalParticipant: p[BijuxDAGRuntimeID]},
		OuroLoopID:        &OuroLoopParticipant{ExternalParticipant: p[OuroLoopID]},
		RecursID:          &RecursParticipant{ExternalParticipant: p[RecursID]},
	}
}
