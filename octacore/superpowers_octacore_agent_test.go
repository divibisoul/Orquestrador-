package octacore

import "testing"

func TestSuperpowersOctacoreAgentPreflight(t *testing.T) {
  a := NewSuperpowersOctacoreAgent()
  if err := a.Preflight(OpFusionExecute); err != nil { t.Fatal(err) }
  if a.Describe()["id"] != "superpowers.octacore" { t.Fatal("wrong agent id") }
}
