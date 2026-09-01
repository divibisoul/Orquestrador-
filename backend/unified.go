package backend

import "github.com/divibisoul/Orquestrador-/orchestrator"

func NewUnified(engine *orchestrator.Engine, cfg Config) *Server {
	s := New(engine, cfg)
	_ = orchestrator.RegisterSuperGPUOperations(engine)
	return s
}
