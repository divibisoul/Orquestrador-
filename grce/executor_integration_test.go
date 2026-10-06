//go:build integration

package grce

import (
	"context"
	"os"
	"testing"

	"github.com/divibisoul/Orquestrador-/backend"
)

func TestGRCEAgainstRealSARA(t *testing.T) {
	serviceURL := os.Getenv("SARA_SERVICE_URL")
	token := os.Getenv("SARA_SERVICE_TOKEN")
	if serviceURL == "" || token == "" {
		t.Fatal("real SARA runtime is required; set SARA_SERVICE_URL and SARA_SERVICE_TOKEN")
	}

	cfg := backend.DefaultConfig()
	cfg.SARAServiceURL = serviceURL
	cfg.SARAServiceToken = token
	proxy := backend.NewSARAProxy(cfg)
	if !proxy.Configured() {
		t.Fatal("SARA proxy did not accept the real service configuration")
	}

	participant := NewCompositeParticipant(
		NewSARAParticipant(proxy),
		ConfiguredExternalParticipants()...,
	)
	cycle := New(participant).Execute(context.Background(),
		"preserve evidence and transform the detected divergence without deleting history",
		"grce-real-sara")
	if cycle.State != StateReal {
		t.Fatalf("GRCE was not REAL: %+v", cycle)
	}
	if cycle.FinalOutputHash == "" {
		t.Fatal("final output hash missing")
	}
	if len(cycle.Stages) != 6 {
		t.Fatalf("expected six real hooks, got %d", len(cycle.Stages))
	}
	if len(cycle.Providers) != 3 {
		t.Fatalf("expected three external provider evidences, got %d", len(cycle.Providers))
	}
	for _, provider := range cycle.Providers {
		if provider.State != StateReal {
			t.Fatalf("provider %s did not execute as REAL: %+v", provider.Provider, provider)
		}
	}
}
