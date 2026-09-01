package orchestrator

import("context";"testing";"github.com/divibisoul/Orquestrador-/neural";"github.com/divibisoul/Orquestrador-/prefrontal";"github.com/divibisoul/Orquestrador-/supergpu")
func TestEngineRuntime(t *testing.T){n,_:=neural.New(2,.1);c,_:=prefrontal.New(.1,4);g:=supergpu.New(nil);g.Discover();e,err:=New(n,c,g);if err!=nil{t.Fatal(err)};r,err:=e.Execute(context.Background(),"compute.execute",[]float64{2,3},map[string]string{"operation":"square"});if err!=nil{t.Fatal(err)};if len(r.Payload)!=2||r.Payload[0]!=4{t.Fatal("orchestration result incorrect")};if e.Status()!="ready"{t.Fatal("engine not ready")};if e.Health()["nucleus"]!="N07"{t.Fatal("wrong nucleus")};if err=e.Shutdown(context.Background());err!=nil{t.Fatal(err)}}
