package orchestrator

import (
  "context"
  "testing"
  "github.com/divibisoul/Orquestrador-/learning"
  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/protocol"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func newEngineForTest(t *testing.T) *Engine {
  t.Helper()
  n, err := neural.New(8, 0.05); if err != nil { t.Fatal(err) }
  c, err := prefrontal.New(0.10, 32); if err != nil { t.Fatal(err) }
  e, err := New(n, c, supergpu.New(nil)); if err != nil { t.Fatal(err) }
  return e
}

func TestCanonicalNeuralParametersOperation(t *testing.T) {
  e := newEngineForTest(t)
  m := protocol.NewMessage("N02","N07","command","neural.parameters@1.0.0",nil)
  r, err := e.Submit(context.Background(),m)
  if err != nil { t.Fatal(err) }
  if r.Status != "ok" || r.Metadata["parameters"] == "" { t.Fatalf("unexpected result: %+v",r) }
}

func TestLearningMachineBecomesCanonicalNeuralLearnPath(t *testing.T) {
  e := newEngineForTest(t)
  machine, err := learning.New(e.neural, e.cortex, nil); if err != nil { t.Fatal(err) }
  if err := e.SetLearningMachine(machine); err != nil { t.Fatal(err) }
  if !containsString(e.Operations(),"learning.feedback@1.0.0") { t.Fatal(e.Operations()) }
  m := protocol.NewMessage("N06","N07","command","neural.learn@1.0.0",[]float64{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})
  r, err := e.Submit(context.Background(),m)
  if err != nil { t.Fatal(err) }
  if r.Status != "ok" { t.Fatalf("unexpected result: %+v",r) }
  if machine.Snapshot().Supervised != 1 { t.Fatalf("machine did not observe learning: %+v",machine.Snapshot()) }
}
func containsString(values []string,want string) bool { for _,v:=range values{if v==want{return true}};return false }
