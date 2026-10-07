package orchestrator

import (
	"context"
	"testing"
)

func TestResolveCapabilityUpgradeExposesAllProviders(t *testing.T) {
	plan, err := ResolveCapabilityUpgrade(context.Background(), "N07", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.FederatedSources) != 16 {
		t.Fatalf("providers=%d want=16", len(plan.FederatedSources))
	}
	if !plan.RequiresExplicitAdapter || !plan.NoFakeRuntimeSuccess {
		t.Fatal("activation policy weakened")
	}
}

func TestResolveCapabilityUpgradeUsesDirectFunctionalAffinity(t *testing.T) {
	plan, err := ResolveCapabilityUpgrade(context.Background(), "N03", "speech-to-text")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range plan.DirectAugmenters {
		if p.ID == "whisper" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("whisper direct affinity missing for N03 speech-to-text")
	}
}
