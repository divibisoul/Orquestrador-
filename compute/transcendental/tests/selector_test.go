package tests

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/selection"
)

func workload(id string, precision core.Precision, matrix int, memory int64) core.Workload {
	return core.Workload{ID: id, Operation: "matmul", Precision: precision, MatrixSize: matrix, BatchSize: 1, DataBytes: 1 << 20, MemoryNeeded: memory}
}

func TestSelectorSmallPrefersTrillium(t *testing.T) {
	cfg := core.DefaultConfig()
	s, err := selection.Select(workload("small", core.BF16, 512, 1), models.DefaultCatalog(cfg), "auto", "auto", cfg.PrecisionFallback)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model.Name() != "Google TPU v6e Trillium" {
		t.Fatalf("got %s", s.Model.Name())
	}
}

func TestSelectorLargeRequiresEnoughMemory(t *testing.T) {
	cfg := core.DefaultConfig()
	s, err := selection.Select(workload("large", core.FP8, 8192, 300), models.DefaultCatalog(cfg), "auto", "auto", cfg.PrecisionFallback)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model.GetMemoryCapacityGB() < 300 {
		t.Fatalf("selected insufficient memory model %s", s.Model.Name())
	}
}

func TestSelectorFP64RestrictsFamily(t *testing.T) {
	cfg := core.DefaultConfig()
	s, err := selection.Select(workload("fp64", core.FP64, 2048, 1), models.DefaultCatalog(cfg), "auto", "auto", cfg.PrecisionFallback)
	if err != nil {
		t.Fatal(err)
	}
	if s.EffectivePrecision != core.FP64 {
		t.Fatalf("precision fallback unexpected: %s", s.EffectivePrecision)
	}
	if s.Model.Name() != "NVIDIA Blackwell B200" && s.Model.Name() != "NVIDIA Vera Rubin (preliminary 2026)" {
		t.Fatalf("FP64 selected unsupported family: %s", s.Model.Name())
	}
}

func TestSelectorExplicitMode(t *testing.T) {
	cfg := core.DefaultConfig()
	s, err := selection.Select(workload("explicit", core.FP8, 2048, 1), models.DefaultCatalog(cfg), "blackwell", "auto", cfg.PrecisionFallback)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model.Name() != "NVIDIA Blackwell B200" {
		t.Fatalf("got %s", s.Model.Name())
	}
}

func TestSelectorPrecisionFirstAvoidsFallback(t *testing.T) {
	cfg := core.DefaultConfig()
	s, err := selection.Select(workload("precision", core.BF16, 4096, 1), models.DefaultCatalog(cfg), "auto", "precision_first", cfg.PrecisionFallback)
	if err != nil {
		t.Fatal(err)
	}
	if s.FallbackUsed {
		t.Fatal("precision_first selected a fallback precision")
	}
}

func TestSelectorMemoryFirstUsesCapacityRatio(t *testing.T) {
	cfg := core.DefaultConfig()
	s, err := selection.Select(workload("memory", core.FP8, 4096, 250), models.DefaultCatalog(cfg), "auto", "memory_first", cfg.PrecisionFallback)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model.GetMemoryCapacityGB() < 250 {
		t.Fatalf("selected insufficient memory model %s", s.Model.Name())
	}
}

func TestSelectorRejectsUnsupportedStrategy(t *testing.T) {
	cfg := core.DefaultConfig()
	if _, err := selection.Select(workload("bad", core.FP16, 1024, 1), models.DefaultCatalog(cfg), "auto", "unknown", cfg.PrecisionFallback); err == nil {
		t.Fatal("expected unsupported strategy error")
	}
}

func TestSelectorRejectsUnsupportedFallback(t *testing.T) {
	cfg := core.DefaultConfig()
	if _, err := selection.Select(workload("bad-fallback", core.FP16, 1024, 1), models.DefaultCatalog(cfg), "auto", "auto", core.Precision("bogus")); err == nil {
		t.Fatal("expected unsupported fallback precision error")
	}
}

func TestSelectorRejectsInsufficientMemory(t *testing.T) {
	cfg := core.DefaultConfig()
	if _, err := selection.Select(workload("memory", core.FP16, 1024, 10000), models.DefaultCatalog(cfg), "auto", "auto", cfg.PrecisionFallback); err == nil {
		t.Fatal("expected no compatible model")
	}
}
