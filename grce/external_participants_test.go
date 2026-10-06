package grce

import (
	"context"
	"testing"
)

func TestConfiguredExternalParticipantsExposeDistinctGoldenRuleHooks(t *testing.T) {
	participants := ConfiguredExternalParticipants()
	if len(participants) != 3 {
		t.Fatalf("expected three external GoldenRule participants, got %d", len(participants))
	}

	expected := map[string]Hook{
		"bijux-core": HookCharacterize,
		"ouro-loop":  HookValidate,
		"recuris":    HookTrace,
	}

	for _, participant := range participants {
		named, ok := participant.(interface{ ID() string })
		if !ok {
			t.Fatalf("participant %T does not expose its provider identity", participant)
		}
		staged, ok := participant.(interface{ Hook() Hook })
		if !ok {
			t.Fatalf("participant %T does not expose its designated hook", participant)
		}
		want, ok := expected[named.ID()]
		if !ok {
			t.Fatalf("unexpected provider %q", named.ID())
		}
		if staged.Hook() != want {
			t.Fatalf("provider %q assigned to %q, want %q", named.ID(), staged.Hook(), want)
		}
		if _, ok := participant.(GoldenRuleParticipant); !ok {
			t.Fatalf("provider %q does not implement GoldenRuleParticipant", named.ID())
		}
	}

	primary := invalidParticipant{}
	composite := NewCompositeParticipant(primary, participants...)
	if _, err := composite.ExecuteHook(context.Background(), HookDetect, "x", "corr", "cycle"); err != nil {
		t.Fatalf("primary participant unexpectedly failed: %v", err)
	}
}
