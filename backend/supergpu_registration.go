package backend

import "github.com/divibisoul/Orquestrador-/orchestrator"

func (s *Server) ensureSuperGPUOperations() {
	if s == nil || s.Engine == nil {
		return
	}
	_ = orchestrator.RegisterSuperGPUOperations(s.Engine)
}
