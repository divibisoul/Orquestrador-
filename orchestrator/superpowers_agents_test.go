package orchestrator

import (
  "context"
  "testing"
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
  e := NewTestEngine()
  if err := RegisterSuperpowersAgentOperations(e); err != nil { t.Fatal(err) }
  m := NewTestMessage("N01","N07","superpowers.agent.route@1.0.0")
  m.Metadata["agent_id"] = "superpowers.mesh-clareira"
  m.Metadata["operation"] = "octacore.batch"
  _, err := e.Submit(context.Background(), m)
  if err == nil { t.Fatal("expected fail-closed rejection") }
}
