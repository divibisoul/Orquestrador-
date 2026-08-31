package interfaces

import (
	"context"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
)

type ComputeBackend interface {
	Execute(ctx context.Context, wl core.Workload) (core.Result, error)
	Estimate(ctx context.Context, wl core.Workload) (core.CostEstimate, error)
}
