package grf

import (
	"errors"
	"fmt"
)

type ParticipantDescriptor struct {
	ID string
	Source string
	Role string
	State EpistemicState
	Capabilities []Capability
}

type BoundaryParticipant struct {
	desc ParticipantDescriptor
	invariants InvariantSet
}

func NewBoundaryParticipant(desc ParticipantDescriptor) (*BoundaryParticipant,error) {
	if desc.ID=="" || desc.Source=="" || desc.Role=="" {
		return nil,errors.New("participant descriptor incomplete")
	}
	if err:=ValidateInvariantSet(CanonicalInvariantSet());err!=nil{return nil,err}
	return &BoundaryParticipant{desc:desc,invariants:CanonicalInvariantSet()},nil
}

func (p *BoundaryParticipant) Ingest(input State, c Context) (State, Provenance, Evidence) {
	parent, _ := input.Hash()
	stage := "PARTICIPANT_INGEST"
	if p.desc.State==BLOCKED { stage="PARTICIPANT_INGEST_BLOCKED" }
	payload := map[string]any{
		"participant": p.desc.ID,
		"source": p.desc.Source,
		"role": p.desc.Role,
		"input": input.Payload,
		"external_execution": false,
		"epistemic_state": string(p.desc.State),
	}
	out := State{
		ID: input.ID + ":" + p.desc.ID,
		EpistemicState: p.desc.State,
		Payload: payload,
		ParentHash: parent,
	}
	prov, err := SealProvenance(parent,input,out,c.SequenceIndex,stage,"")
	if err!=nil {
		return input, Provenance{ParentHash:parent,SequenceIndex:c.SequenceIndex,Stage:stage}, Evidence{
			ID:"evidence:participant:"+p.desc.ID,FailureID:"participant-ingest",State:PRESERVED,SequenceIndex:c.SequenceIndex,
		}
	}
	evState:=p.desc.State
	ev:=Evidence{
		ID:"evidence:participant:"+p.desc.ID,
		FailureID:"participant-source",
		State:evState,
		Hash:prov.OutputHash,
		InputHash:prov.InputHash,
		OutputHash:prov.OutputHash,
		SequenceIndex:c.SequenceIndex,
		Payload:map[string]any{"source":p.desc.Source,"role":p.desc.Role,"external_execution":false},
	}
	return out,prov,ev
}

func (p *BoundaryParticipant) EpistemicState() EpistemicState { return p.desc.State }
func (p *BoundaryParticipant) Invariants() []string {
	ids:=InvariantIDs(p.invariants)
	return append([]string(nil),ids...)
}
func (p *BoundaryParticipant) Capabilities() []Capability { return append([]Capability(nil),p.desc.Capabilities...) }
func (p *BoundaryParticipant) FailuresAbsorbed() []Failure {
	state:=PRESERVED
	if p.desc.State==BLOCKED { state=BLOCKED }
	return []Failure{{ID:"participant-boundary:"+p.desc.ID,Source:p.desc.Source,Description:"absence or non-proof of external runtime execution is preserved at the GRF boundary",State:state}}
}
func (p *BoundaryParticipant) Provenance() []Provenance {
	prov, _ := SealProvenance("", map[string]any{"source":p.desc.Source}, map[string]any{"state":p.desc.State},1,"PARTICIPANT_REGISTRATION","")
	return []Provenance{prov}
}

func ParticipantDescriptors() []ParticipantDescriptor {
	return []ParticipantDescriptor{
		{ID:"autogenesis",Source:"DVampire/Autogenesis",Role:"self-evolution",State:PROJECTED},
		{ID:"cognifold",Source:"OpenNerve/CogniFold",Role:"neocortex",State:BLOCKED},
		{ID:"belel-protocol",Source:"TTOPM/belel-protocol",Role:"nervo-vago",State:BLOCKED},
		{ID:"opensinn-bus",Source:"OpenSIN-AI/OpenSIN-Neural-Bus",Role:"vagus-bus",State:BLOCKED},
		{ID:"octos",Source:"lispking/octos",Role:"octacore reinforcement",State:PROJECTED},
		{ID:"hora-graph-core",Source:"Vivien83/hora-graph-core",Role:"hortacore reinforcement",State:PROJECTED},
		{ID:"mycelium",Source:"mycelium-io/mycelium",Role:"clareira workspace",State:PROJECTED},
		{ID:"prime-agent",Source:"PrimeIntellect-ai/prime-agent",Role:"recursive neocortex/clareira",State:PROJECTED},
		{ID:"cuda-oxide",Source:"SuperInstance/cuda-oxide",Role:"supergpu reinforcement",State:PROJECTED},
		{ID:"functional-graph-agi",Source:"kexi-bq/functional-graph-agi",Role:"supergpu-agi",State:BLOCKED},
	}
}

func ParticipantRegistry() (map[string]*BoundaryParticipant,error) {
	out:=map[string]*BoundaryParticipant{}
	for _,desc:=range ParticipantDescriptors(){
		p,err:=NewBoundaryParticipant(desc)
		if err!=nil{return nil,fmt.Errorf("%s: %w",desc.ID,err)}
		out[desc.ID]=p
	}
	return out,nil
}
