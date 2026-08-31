package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/mesh"
)

type request struct { Goal string `json:"goal"` }

type Server struct {
	Engine *orchestrator.Engine
	PFC    *prefrontal.Cortex
	Mesh   *mesh.Registry
}

func NewServer(engine *orchestrator.Engine, pfc *prefrontal.Cortex, registry *mesh.Registry) *Server {
	if engine == nil { engine = orchestrator.NewEngine(4) }
	if pfc == nil { pfc = prefrontal.New() }
	if registry == nil { registry = mesh.NewRegistry() }
	return &Server{Engine:engine, PFC:pfc, Mesh:registry}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/v1/plan", s.plan)
	mux.HandleFunc("/v1/nodes", s.nodes)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	if s == nil { http.Error(w, "rest server unavailable", http.StatusServiceUnavailable); return }
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status":"ok","nodes":len(s.Mesh.Snapshot())})
}

func (s *Server) plan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(http.StatusMethodNotAllowed); return }
	var q request
	if json.NewDecoder(r.Body).Decode(&q) != nil || strings.TrimSpace(q.Goal) == "" { http.Error(w,"goal required",http.StatusBadRequest); return }
	plan := s.PFC.GeneratePlan(strings.TrimSpace(q.Goal), []string{"evaluate","execute"})
	decision := s.PFC.Decide([]prefrontal.Plan{plan}, float64(len(plan.Steps)))
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(decision)
}

func (s *Server) nodes(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(s.Mesh.Snapshot())
}
