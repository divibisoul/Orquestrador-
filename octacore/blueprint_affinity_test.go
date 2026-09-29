package octacore

import "testing"

func TestBlueprintAffinityExposedThroughProcessor(t *testing.T) {
	processor := &Processor{}
	matches, err := processor.ResolveBlueprint("audio multimodal perception", 3)
	if err != nil {
		t.Fatalf("ResolveBlueprint failed: %v", err)
	}
	if len(matches) == 0 || matches[0].Family != "N1_PERCEPTION" {
		t.Fatalf("expected N1_PERCEPTION primary, got %#v", matches)
	}

	plan, err := processor.ComposeBlueprint("ethics governance memory trace", 4)
	if err != nil {
		t.Fatalf("ComposeBlueprint failed: %v", err)
	}
	if plan.Primary.EntryID == "" || len(plan.Preserved) == 0 {
		t.Fatalf("expected non-empty additive plan: %#v", plan)
	}
}
