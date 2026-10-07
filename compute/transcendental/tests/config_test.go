package tests

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

func TestConfigRejectsInvalidFallbackPrecision(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.PrecisionFallback = core.Precision("invalid")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid fallback precision to be rejected")
	}
}

func TestConfigCapsSimulationParallelUnits(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.MaxParallelUnits = 8192
	if got := cfg.EffectiveParallelUnits(); got != core.MaxSimulationParallelUnits {
		t.Fatalf("expected hard cap %d, got %d", core.MaxSimulationParallelUnits, got)
	}
}

func TestConfigRejectsZeroParallelUnits(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.MaxParallelUnits = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected zero parallel units to be rejected")
	}
}
