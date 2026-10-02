package mesh

import (
  "testing"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestSuperpowersMeshClareiraAgentPreflight(t *testing.T) {
  a := NewSuperpowersMeshClareiraAgent()
  if err := a.Preflight(supergpu.ExecutionEvent{Operation:"x",CorrelationID:"c"}); err != nil { t.Fatal(err) }
  if a.Describe()["id"] != "superpowers.mesh-clareira" { t.Fatal("wrong agent id") }
}
