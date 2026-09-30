package blueprint

import "testing"

func TestManifestIsNonDestructiveAndSingleOwner(t *testing.T) {
	if err := ValidateManifest(); err != nil {
		t.Fatalf("manifest validation failed: %v", err)
	}
	m := DefaultManifest()
	if !m.NonDestructive {
		t.Fatal("manifest must preserve non-destructive policy")
	}
	if len(m.Entries) != 8 {
		t.Fatalf("expected 8 blueprint families, got %d", len(m.Entries))
	}
}

func TestResolveFunctionByAffinity(t *testing.T) {
	matches, err := Resolve("voice multimodal perception audio transcription", 3)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if len(matches) == 0 || matches[0].Family != "N1_PERCEPTION" {
		t.Fatalf("expected perception as primary, got %#v", matches)
	}
}

func TestResolveEthicsPreservesSARAAndN01Complement(t *testing.T) {
	match, err := ResolveFamily("N4_ETHICS")
	if err != nil {
		t.Fatalf("ResolveFamily failed: %v", err)
	}
	owners := map[string]bool{}
	for _, affinity := range match.Affinities {
		owners[affinity.Owner] = true
	}
	if !owners["SARA"] || !owners["N01"] {
		t.Fatalf("expected SARA canonical ethics plus N01 complementary guard, got %#v", owners)
	}
}

func TestLearningDoesNotClaimTwentyParadigms(t *testing.T) {
	manifest := DefaultManifest()
	for _, entry := range manifest.Entries {
		if entry.Family == "N6_LEARNING" {
			if entry.CoverageNote == "" {
				t.Fatal("learning coverage boundary must be explicit")
			}
			return
		}
	}
	t.Fatal("learning family missing")
}

func TestComposePreservesExistingAffinities(t *testing.T) {
	plan, err := Compose("ethics privacy governance memory traceability", 4)
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if plan.Primary.EntryID == "" || len(plan.Preserved) == 0 {
		t.Fatalf("expected composed plan with preserved affinities: %#v", plan)
	}
}
