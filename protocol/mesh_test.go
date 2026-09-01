package protocol

import "testing"

func TestMeshRoundTrip(t *testing.T){m:=MeshEnvelope{ContractVersion:"1.2",Operation:"execute",Payload:map[string]any{"x":1},CorrelationID:"corr-1",Source:"N07",Target:"N01"};b,err:=EncodeMesh(m);if err!=nil{t.Fatal(err)};d,err:=DecodeMesh(b);if err!=nil{t.Fatal(err)};if d.CorrelationID!=m.CorrelationID||d.Operation!=m.Operation{t.Fatal("mesh round-trip failed")}}
func TestMeshRejectsInvalidEnvelope(t *testing.T){cases:=[]MeshEnvelope{{},{ContractVersion:"1.2"},{ContractVersion:"1.2",Operation:"execute",CorrelationID:"c",Source:"N07"},{ContractVersion:"1.2",Operation:"execute",CorrelationID:"c",Source:"N07",Target:"N01"}};for i,m:=range cases{if i==len(cases)-1{if err:=m.Validate();err!=nil{t.Fatal(err)};continue};if _,err:=EncodeMesh(m);err==nil{t.Fatalf("case %d accepted",i)}}}
