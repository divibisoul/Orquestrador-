package trinity

import "context"

type Workload struct {
	ID           string
	Kind         string
	Payload      string
	MatrixSize   int
	BatchSize    int
	MemoryNeeded float64
	Precision    string
	Risk         float64
	Metadata     map[string]string
}

type CostEstimate struct {
	LatencyMS   float64
	Memory      float64
	ComputeCost float64
	Confidence  float64
}

type Strategy struct {
	Name        string
	Parallelism int
	Precision   string
	Priority    int
}

type Decision struct {
	Approved       bool
	ConflictScore  float64
	Strategy       Strategy
	Estimate       CostEstimate
	Reason         string
}

type Route struct {
	Target       string
	Model        string
	Provider     string
	Score        float64
	Fallback     string
	Capabilities []string
}

type Result struct {
	Output     string
	Route      Route
	LatencyMS  float64
	Success    bool
	Error      string
	Metadata   map[string]string
}

type Feedback struct {
	WorkloadID string
	Route      Route
	Success    bool
	LatencyMS  float64
	Quality    float64
	Error      string
}

type Prefrontal interface {
	Evaluate(context.Context, Workload, ComputeEngine) (Decision, error)
	GateWorkingMemory(context.Context, Workload, Result) error
}

type NeuralFabric interface {
	Route(context.Context, Strategy, Workload) (Route, error)
	Feedback(context.Context, Feedback) error
}

type ComputeEngine interface {
	Execute(context.Context, Workload, Route) (Result, error)
}

type CostEstimator interface {
	Estimate(context.Context, Workload) (CostEstimate, error)
}

type TrinityConfig struct {
	PFCEnabled     bool
	FabricEnabled  bool
	ComputeEnabled bool
	RiskThreshold  float64
	FallbackMode   string
	Prefrontal     PrefrontalConfig
	Fabric         FabricConfig
	Compute        ComputeConfig
}

type PrefrontalConfig struct {
	WorkingMemoryLimit int
	MetaRLEpsilon      float64
	ConflictSensitivity float64
}

type FabricConfig struct {
	DecisionTreeDepth int
	LearningRate      float64
	FeedbackDiscount  float64
}

type ComputeConfig struct {
	Mode             string
	PrecisionFallback string
	EfficiencyFactor float64
	UseSparsity      bool
	NoiseStd         float64
}

type TrinityOrchestrator struct {
	PFC     Prefrontal
	Fabric  NeuralFabric
	Compute ComputeEngine
	Config  TrinityConfig
}
