package orchestrator

import "testing"

func TestN07ResidentAgentBound(t *testing.T) {
	a := N07ResidentAgent()
	if a.ID != "N07.resident" || a.Nucleus != "N07" || a.Lifecycle != "BOUND" {
		t.Fatalf("unexpected resident agent: %#v", a)
	}
	if a.ProviderCount != 25 {
		t.Fatalf("provider count=%d want=25", a.ProviderCount)
	}
	if len(a.Capabilities) < 3 {
		t.Fatalf("capabilities=%d want>=3", len(a.Capabilities))
	}
}
