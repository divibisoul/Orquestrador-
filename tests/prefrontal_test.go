package tests
import("testing";"github.com/divibisoul/Orquestrador-/core/prefrontal")
func TestCortex(t *testing.T){c:=prefrontal.New();c.FuseSignals([]prefrontal.Signal{{Name:"queue",Value:10}});if c.ReadContext().Values["queue"]!=10{t.Fatal("state fusion failed")};p:=c.GeneratePlan("x",[]string{"a","b"});if c.Decide([]prefrontal.Plan{p},10).ID!=p.ID{t.Fatal("decision failed")}}
