package orchestrator

import(
 "context"
 "testing"
 "github.com/divibisoul/Orquestrador-/neural"
 "github.com/divibisoul/Orquestrador-/prefrontal"
 "github.com/divibisoul/Orquestrador-/protocol"
 "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestSubmitMeshRoundTrip(t *testing.T){n,_:=neural.New(2,.1);c,_:=prefrontal.New(.1,4);e,err:=New(n,c,supergpu.New(nil));if err!=nil{t.Fatal(err)};if err:=e.Register("tool.echo@1.0.0",func(_ context.Context,m protocol.Message)(protocol.Result,error){return protocol.Result{TraceID:m.TraceID,Source:"N07.tool",Target:m.Source,Status:"ok",Metadata:map[string]string{"tool":"echo"},Payload:append([]float64(nil),m.Payload...)},nil});err!=nil{t.Fatal(err)};env:=protocol.MeshEnvelope{ContractVersion:protocol.MeshContractVersion,Operation:"tool.echo",Payload:map[string]any{"name":"echo","values":[]any{2.0,3.0}},CorrelationID:"trace-mesh-1",Source:"N01",Target:"N07"};out,err:=e.SubmitMesh(context.Background(),env);if err!=nil{t.Fatal(err)};if out.ContractVersion!=protocol.MeshContractVersion||out.CorrelationID!="trace-mesh-1"||out.Source!="N07.tool"||out.Target!="N01"{t.Fatalf("unexpected output=%+v",out)};if got:=out.Payload["values"].([]any);len(got)!=2||got[0].(float64)!=2||got[1].(float64)!=3{t.Fatalf("payload=%+v",out.Payload)}}

func TestSubmitMeshRejectsUnsupportedContract(t *testing.T){n,_:=neural.New(1,.1);c,_:=prefrontal.New(.1,2);e,_:=New(n,c,supergpu.New(nil));_,err:=e.SubmitMesh(context.Background(),protocol.MeshEnvelope{ContractVersion:"other/1",Operation:"x",CorrelationID:"c",Source:"N01",Target:"N07"});if err==nil{t.Fatal("unsupported mesh contract accepted")}}
