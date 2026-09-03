package backend

import "github.com/divibisoul/Orquestrador-/orchestrator"

func NewUnified(engine *orchestrator.Engine, cfg Config) *Server {
	return New(engine, cfg)
}
