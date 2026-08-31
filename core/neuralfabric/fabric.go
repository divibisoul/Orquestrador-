package neuralfabric

import (
	"context"
	"errors"
	"sync"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type Vector []float64
type State struct {
	Goal        string
	Features    Vector
	Constraints map[string]float64
}
type Prediction struct {
	Value      float64
	Confidence float64
}
type Route struct {
	NodeID     string
	DeviceID   string
	Precision  string
	BatchSize  int
	Score      float64
	Confidence float64
}
type Experience struct {
	State  State
	Action Route
	Reward float64
	Done   bool
}
type Encoder interface {
	EncodeState(context.Context, State) (Vector, error)
	EncodeTask(context.Context, string) (Vector, error)
	Normalize(Vector) Vector
}
type Predictor interface {
	Latency(State, Route) Prediction
	Cost(State, Route) Prediction
	Energy(State, Route) Prediction
	Quality(State, Route) Prediction
	Failure(State, Route) Prediction
}
type Router interface {
	Route(context.Context, State, []Route) (Route, error)
}
type Learner interface {
	Observe(Experience)
	Reward(Experience) float64
	Update(context.Context) error
	Save(context.Context) error
	Load(context.Context) error
}

// Fabric is the small deterministic adapter used by Trinity. Legacy contracts above remain intact.
type Fabric struct {
	mu   sync.RWMutex
	cfg  trinity.FabricConfig
	tree *DecisionTree
}

func NewFabric(cfg trinity.FabricConfig) *Fabric {
	if cfg.DecisionTreeDepth < 1 {
		cfg.DecisionTreeDepth = 6
	}
	if cfg.LearningRate <= 0 {
		cfg.LearningRate = 0.01
	}
	if cfg.FeedbackDiscount <= 0 || cfg.FeedbackDiscount > 1 {
		cfg.FeedbackDiscount = 0.9
	}
	return &Fabric{cfg: cfg, tree: NewDecisionTree(cfg.DecisionTreeDepth)}
}

func (f *Fabric) Route(ctx context.Context, strategy trinity.Strategy, w trinity.Workload) (trinity.Route, error) {
	if f == nil {
		return trinity.Route{}, errors.New("nil neural fabric")
	}
	if ctx == nil {
		return trinity.Route{}, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return trinity.Route{}, err
	}
	if w.ID == "" {
		return trinity.Route{}, errors.New("workload id required")
	}

	f.mu.RLock()
	tree := f.tree
	f.mu.RUnlock()
	if tree == nil {
		return trinity.Route{}, errors.New("neural fabric decision tree unavailable")
	}
	return route(ctx, tree, strategy, w)
}
