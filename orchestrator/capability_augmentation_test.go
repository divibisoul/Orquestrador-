package orchestrator

import (
	"context"
	"testing"
)

func TestCapabilityAugmentationManifestCovers25RepositoryFabric(t *testing.T) {
	m, err := loadCapabilityAugmentationManifest()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.NativeComponents) != 9 {
		t.Fatalf("components=%d want=9", len(m.NativeComponents))
	}
	if len(m.Providers) != 25 {
		t.Fatalf("providers=%d want=25", len(m.Providers))
	}
}

func TestCapabilityAugmentationSelectsFunctionalProvider(t *testing.T) {
	plan, err := ResolveCapabilityAugmentation(context.Background(), "N03", "audio.transcribe")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, id := range plan.Augmenters {
		if id == "whisper" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("whisper missing from N03 audio.transcribe augmentation")
	}
	if plan.EvidenceState != "PROJECTED" {
		t.Fatalf("evidence=%s want=PROJECTED", plan.EvidenceState)
	}
}

func TestCapabilityAugmentationRejectsUnknownNativeCapability(t *testing.T) {
	plan, err := ResolveCapabilityAugmentation(context.Background(), "N03", "not.real")
	if err == nil {
		t.Fatal("expected rejection")
	}
	if plan.EvidenceState != "BLOCKED" {
		t.Fatalf("evidence=%s want=BLOCKED", plan.EvidenceState)
	}
}
