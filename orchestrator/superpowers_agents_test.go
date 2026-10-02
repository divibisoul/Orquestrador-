package orchestrator

import (
  "context"
  "testing"

  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestSuperpowersAgentBindingsCoverRequestedCrossfronts(t *testing.T) {
  want := map[string][]string{
    "superpowers.sara-trinity": {"ARA","ETR","ITR","RGO","MMD"},
    "superpowers.cortex-orbital-supergpu": {"prefrontal-neocortex","orbital-reasoning","SuperGPU"},
    "superpowers.mesh-clareira": {"SOUL-Mesh","Projeto-Clareira"},
    "superpowers.octacore": {"Octacore","Mesh","SuperGPU"},
  }
  got := SuperpowersAgentBindings()
  if len(got) != len(want) { t.Fatalf("bindings=%d want=%d", len(got), len(want)) }
  for _, b := range got {
    expected, ok := want[b.ID]
    if !ok { t.Fatalf("unexpected binding %s", b.ID) }
    if len(expected) != len(b.Targets) { t.Fatalf("%s targets=%v want=%v", b.ID, b.Targets, expected) }
  }
}

func TestSuperpowersAgentRouteRejectsUnboundOperation(t *testing.T) {
  n, err := neural.New(2, .05)
  if err != nil { t.Fatal(err) }
  p, err := prefrontal.New(.1, 8)
  if err != nil { t.Fatal(err) }
  g := supergpu.New(nil)
  e, err := New(n, p, g)
  if err != nil { t.Fatal(err) }
  if err := RegisterSuperpowersAgentOperations(e); err != nil { t.Fatal(err) }
  m := NewMessage("N01", "N07", "command", SuperpowersAgentRouteOperation, nil)
  m.Metadata["agent_id"] = "superpowers.mesh-clareira"
  m.Metadata["operation"] = "octacore.batch"
  _, err = e.Submit(context.Background(), m)
  if err == nil { t.Fatal("expected fail-closed rejection") }
}
