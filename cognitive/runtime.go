package cognitive

import (
	"errors"

	"github.com/divibisoul/Orquestrador-/prefrontal"
)

func Build(cfg Config, cortex *prefrontal.Cortex, peers MeshPeer, sara PolicyAuditor, store RunStore, local ...LocalExecutor) (*Loop, error) {
	if !cfg.Enabled {
		return nil, errors.New("COGNITIVE_LOOP_DISABLED")
	}
	executor, err := NewMeshExecutor(peers, local...)
	if err != nil {
		return nil, err
	}
	critic, err := NewCritic(cortex, sara, cfg)
	if err != nil {
		return nil, err
	}
	return New(cfg, NewGoalStore(), NewWorkingMemory(cfg), executor, critic, store)
}
