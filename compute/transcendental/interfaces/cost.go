package interfaces

import (
	"context"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type CostEstimator interface {
	EstimateCost(ctx context.Context, plan core.Plan) (core.CostEstimate, error)
}
