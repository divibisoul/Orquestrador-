package core

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Precision string

const (
	FP4  Precision = "fp4"
	FP8  Precision = "fp8"
	FP16 Precision = "fp16"
	BF16 Precision = "bf16"
	FP32 Precision = "fp32"
	FP64 Precision = "fp64"
	INT8 Precision = "int8"
)

func (p Precision) Valid() bool {
	switch p {
	case FP4, FP8, FP16, BF16, FP32, FP64, INT8:
		return true
	default:
		return false
	}
}

type Workload struct {
	ID           string
	Operation    string
	Precision    Precision
	MatrixSize   int
	BatchSize    int
	DataBytes    int64
	MemoryNeeded int64
	Priority     int
	Metadata     map[string]string
}

type Metrics struct {
	Architecture     string
	EffectivePFLOPS  float64
	BandwidthUsedGBs float64
	LatencyMs        float64
	MemoryUsedGB     float64
	Efficiency       float64
	Timestamp        time.Time
}

type CostEstimate struct {
	EstimatedTime     time.Duration
	EstimatedPFLOPS   float64
	EstimatedMemoryGB float64
	EstimatedEnergy   float64
	Architecture      string
	Confidence        float64
}

type Result struct {
	WorkloadID string
	Metrics    Metrics
	Data       []byte
	Error      error
}

type Plan struct {
	Workloads []Workload
	Strategy  string
}

func (w Workload) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return errors.New("workload ID is required")
	}
	if strings.TrimSpace(w.Operation) == "" {
		return errors.New("workload operation is required")
	}
	if !w.Precision.Valid() {
		return fmt.Errorf("unsupported precision: %q", w.Precision)
	}
	if w.MatrixSize < 1 {
		return errors.New("matrix size must be >= 1")
	}
	if w.BatchSize < 1 {
		return errors.New("batch size must be >= 1")
	}
	if w.DataBytes < 0 || w.MemoryNeeded < 0 {
		return errors.New("data and memory requirements must be non-negative")
	}
	return nil
}

func (p Plan) Validate() error {
	if len(p.Workloads) == 0 {
		return errors.New("plan contains no workloads")
	}
	strategy := strings.ToLower(strings.TrimSpace(p.Strategy))
	if strategy != "" && strategy != "auto" && strategy != "precision_first" && strategy != "latency_first" && strategy != "memory_first" {
		return fmt.Errorf("unsupported strategy: %q", p.Strategy)
	}
	for i, w := range p.Workloads {
		if err := w.Validate(); err != nil {
			return fmt.Errorf("invalid workload at index %d: %w", i, err)
		}
	}
	return nil
}
