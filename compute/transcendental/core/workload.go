package core

import "time"

type Precision string

const (
	FP4 Precision = "fp4"
	FP8 Precision = "fp8"
	FP16 Precision = "fp16"
	BF16 Precision = "bf16"
	FP32 Precision = "fp32"
	FP64 Precision = "fp64"
	INT8 Precision = "int8"
)

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
