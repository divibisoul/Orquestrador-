package orchestrator

import "testing"

func TestIntegrationEngineeringAgentSet(t *testing.T) {
	if len(IntegrationEngineers()) != 6 {
		t.Fatal("integration engineering agent count mismatch")
	}
}

func TestN07IdentityExposesAllEngineeringAgents(t *testing.T) {
  ids := map[string]bool{}
  for _, a := range N07Agents() { ids[a.ID] = true }
  for _, id := range []string{
    "soul.engineer.fabric", "soul.engineer.provenance", "soul.engineer.adapter",
    "soul.engineer.runtime", "soul.engineer.verification", "soul.engineer.resilience",
  } { if !ids[id] { t.Fatalf("missing engineering agent %s", id) } }
}
