package orchestrator

import "testing"

func TestIntegrationEngineeringAgentSet(t *testing.T) {
	if len(IntegrationEngineers()) != 5 {
		t.Fatal("integration engineering agent count mismatch")
	}
}
