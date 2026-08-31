package orchestrator

import (
	"context"
	"errors"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

// TrinityRuntime wraps the legacy Engine without changing its execution path.
// Trinity is opt-in: when it is nil or disabled, callers continue using Engine.
type TrinityRuntime struct {
	Legacy  *Engine
	Trinity *trinity.TrinityOrchestrator
}

func NewTrinityRuntime(legacy *Engine, tri *trinity.TrinityOrchestrator) *TrinityRuntime {
	if legacy == nil { legacy = NewEngine(1) }
	return &TrinityRuntime{Legacy: legacy, Trinity: tri}
}

func (r *TrinityRuntime) ExecuteTask(ctx context.Context, task interface{}) (trinity.Result, error) {
	if r == nil || r.Legacy == nil { return trinity.Result{}, errors.New("nil runtime") }
	if r.Trinity == nil || !r.Trinity.Config.PFCEnabled || !r.Trinity.Config.FabricEnabled || !r.Trinity.Config.ComputeEnabled {
		return trinity.Result{}, trinity.ErrDisabled
	}
	return r.Trinity.ExecuteTask(ctx, task)
}
