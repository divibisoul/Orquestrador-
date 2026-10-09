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
	if len(plan.FederatedSources) != 22 {
		t.Fatalf("providers=%d want=22 (16 primary + 6 complementary)", len(plan.FederatedSources))
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


func TestResolveCapabilityUpgradeIncludesComplementaryAffinity(t *testing.T) {
	plan, err := ResolveCapabilityUpgrade(context.Background(), "N07", "parallel-fanout")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range plan.DirectAugmenters {
		if p.ID == "octos" {
			found = true
			if p.Evidence != "PROJECTED" {
				t.Fatalf("complementary provider evidence=%q, want PROJECTED", p.Evidence)
			}
			break
		}
	}
	if !found {
		t.Fatal("octos complementary provider was not resolved for N07 parallel-fanout")
	}
	if !plan.RequiresExplicitAdapter || !plan.NoFakeRuntimeSuccess {
		t.Fatal("complementary capability resolution must retain explicit-adapter and no-fake-success gates")
	}
}

func TestResolveCapabilityUpgradeKeepsNativeOwnershipWhenNoAffinityMatches(t *testing.T) {
	plan, err := ResolveCapabilityUpgrade(context.Background(), "N05", "CUDA-compiler")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.DirectAugmenters) != 0 {
		t.Fatalf("unexpected direct N05 augmenter for CUDA-compiler: %#v", plan.DirectAugmenters)
	}
	if plan.NativeAuthority != "N05" {
		t.Fatalf("native authority changed to %q", plan.NativeAuthority)
	}
}
