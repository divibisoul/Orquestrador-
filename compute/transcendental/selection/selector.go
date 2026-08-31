package selection

import (
	"errors"
	"fmt"
	"strings"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
)

type Selection struct {
	Model              models.PerformanceModel
	EffectivePrecision core.Precision
	FallbackUsed       bool
	Penalty             float64
}

func Select(wl core.Workload, catalog []models.PerformanceModel, mode string, strategy string, fallback core.Precision) (Selection, error) {
	if len(catalog) == 0 {
		return Selection{}, errors.New("no performance models available")
	}
	if err := wl.Validate(); err != nil {
		return Selection{}, err
	}
	if fallback == "" {
		fallback = core.FP32
	}
	strategy = strings.ToLower(strings.TrimSpace(strategy))
	if strategy == "" {
		strategy = "auto"
	}
	if strategy != "auto" && strategy != "precision_first" && strategy != "latency_first" && strategy != "memory_first" {
		return Selection{}, fmt.Errorf("unsupported strategy: %s", strategy)
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "auto" {
		mode = ""
	}

	candidates := make([]Selection, 0, len(catalog))
	for _, model := range catalog {
		if model == nil || model.GetMemoryCapacityGB() < wl.MemoryNeeded || model.GetBandwidthGBs() <= 0 {
			continue
		}
		if mode != "" && !matchesMode(model.Name(), mode) {
			continue
		}
		precision := wl.Precision
		fallbackUsed := false
		penalty := 0.0
		if !model.Supports(precision) {
			if !model.Supports(fallback) {
				continue
			}
			precision = fallback
			fallbackUsed = true
			penalty = 0.25
		}
		if model.GetPFLOPS(precision) <= 0 {
			continue
		}
		candidates = append(candidates, Selection{Model: model, EffectivePrecision: precision, FallbackUsed: fallbackUsed, Penalty: penalty})
	}
	if len(candidates) == 0 {
		return Selection{}, errors.New("no compatible model satisfies mode, memory, and precision requirements")
	}

	if strategy == "auto" && mode == "" {
		if preferred, ok := selectSmallWorkloadTrillium(wl, candidates); ok {
			return preferred, nil
		}
	}

	best := candidates[0]
	bestScore := score(wl, best, strategy)
	for _, candidate := range candidates[1:] {
		s := score(wl, candidate, strategy)
		if s < bestScore || (s == bestScore && candidate.Model.Name() < best.Model.Name()) {
			best, bestScore = candidate, s
		}
	}
	return best, nil
}

func selectSmallWorkloadTrillium(wl core.Workload, candidates []Selection) (Selection, bool) {
	if wl.MatrixSize >= 1024 {
		return Selection{}, false
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate.Model.Name()), "trillium") && !candidate.FallbackUsed {
			return candidate, true
		}
	}
	return Selection{}, false
}

func matchesMode(name, mode string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "blackwell":
		return strings.Contains(n, "blackwell")
	case "vera_rubin":
		return strings.Contains(n, "rubin")
	case "cdna5":
		return strings.Contains(n, "cdna5")
	case "trillium":
		return strings.Contains(n, "trillium")
	case "atlas":
		return strings.Contains(n, "atlas 950")
	default:
		return false
	}
}

func score(wl core.Workload, s Selection, strategy string) float64 {
	estimate := s.Model.EstimateTime(core.Workload{
		ID:           wl.ID,
		Operation:    wl.Operation,
		Precision:    s.EffectivePrecision,
		MatrixSize:   wl.MatrixSize,
		BatchSize:    wl.BatchSize,
		DataBytes:    wl.DataBytes,
		MemoryNeeded: wl.MemoryNeeded,
		Priority:     wl.Priority,
		Metadata:     wl.Metadata,
	})
	base := estimate.Seconds()
	mem := float64(max64(wl.MemoryNeeded, 0))
	penalized := base * (1 + s.Penalty)

	switch strings.ToLower(strategy) {
	case "latency_first":
		return penalized + mem*1e-6
	case "memory_first":
		return mem + base*1e-3*(1+s.Penalty)
	case "precision_first":
		if s.FallbackUsed {
			return penalized + 0.1 + mem*1e-6
		}
		return penalized - 0.1 + mem*1e-6
	}

	if wl.MatrixSize > 4096 && (s.Model.GetBandwidthGBs() > 10000 || s.Model.GetMemoryCapacityGB() > 200) {
		return penalized*0.8 + mem*1e-6
	}
	return penalized + mem*1e-6
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
