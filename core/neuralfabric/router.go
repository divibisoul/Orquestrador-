package neuralfabric

import (
	"context"
	"errors"
	"github.com/divibisoul/Orquestrador-/core/trinity"
)

func route(ctx context.Context, tree *DecisionTree, strategy trinity.Strategy, w trinity.Workload) (trinity.Route, error) {
	if ctx == nil { return trinity.Route{}, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return trinity.Route{}, err }
	if w.ID == "" { return trinity.Route{}, errors.New("workload id required") }
	model := tree.Select(w)
	if strategy.Precision == "" { strategy.Precision = "fp32" }
	return trinity.Route{Target:"local", Model:model, Provider:"transcendental", Score:0.5, Fallback:"local", Capabilities:[]string{"inference", strategy.Precision}}, nil
}
