package grf

import (
	"encoding/json"
	"strings"
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

func TestGRFProvenanceSerializesCanonicalFieldNames(t *testing.T) {
	raw, err := json.Marshal(Provenance{
		ParentHash:"parent", InputHash:"input", OutputHash:"output", SequenceIndex:7, Stage:"TEST",
	})
	if err != nil { t.Fatal(err) }
	s := string(raw)
	for _, forbidden := range []string{"ParentHash","InputHash","OutputHash","SequenceIndex"} {
		if strings.Contains(s, forbidden) { t.Fatalf("non-canonical provenance field leaked: %s", forbidden) }
	}
	for _, required := range []string{"parent_hash","input_hash","output_hash","sequence_index"} {
		if !strings.Contains(s, required) { t.Fatalf("missing canonical provenance field: %s", required) }
	}
}
