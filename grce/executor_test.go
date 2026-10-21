package grce

import (
	"context"
	"testing"
)

type invalidParticipant struct{}

func (invalidParticipant) ExecuteHook(context.Context, Hook, string, string, string) (HookResult, error) {
	return HookResult{EvidenceState: StateBlocked}, nil
}

func TestExecutorDoesNotPromoteNonRealEvidence(t *testing.T) {
	result := New(invalidParticipant{}).Execute(context.Background(), "x", "corr")
	if result.State != StateBlocked {
		t.Fatalf("expected BLOCKED, got %s", result.State)
	}
}
