package grf

import (
	"testing"
)

func TestParticipantRegistryContainsTenRequestedContracts(t *testing.T) {
	r,err:=ParticipantRegistry()
	if err!=nil{t.Fatal(err)}
	if len(r)!=10{t.Fatalf("participants=%d",len(r))}
}

func TestBlockedParticipantNeverClaimsActiveExecution(t *testing.T) {
	r,err:=ParticipantRegistry()
	if err!=nil{t.Fatal(err)}
	p:=r["belel-protocol"]
	if p.EpistemicState()!=BLOCKED{t.Fatal("blocked source must remain blocked")}
	in:=State{ID:"test",EpistemicState:PRESERVED,Payload:map[string]any{"x":1}}
	ctx:=Context{TraceID:"t",CorrelationID:"c",SequenceIndex:1}
	out,prov,ev:=p.Ingest(in,ctx)
	if out.EpistemicState!=BLOCKED || ev.State!=BLOCKED || prov.Stage!="PARTICIPANT_INGEST_BLOCKED"{
		t.Fatalf("unexpected blocked participant state: out=%#v prov=%#v ev=%#v",out,prov,ev)
	}
}
