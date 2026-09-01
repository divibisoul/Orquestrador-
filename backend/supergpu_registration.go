package backend

import (
	"sync"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

var registerSuperGPUOnce sync.Once

func (s *Server) ensureSuperGPUOperations() {
	registerSuperGPUOnce.Do(func() {
		_ = orchestrator.RegisterSuperGPUOperations(s.Engine)
	})
}
