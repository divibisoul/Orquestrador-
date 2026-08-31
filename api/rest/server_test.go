package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/mesh"
)

func TestHealth(t *testing.T) {
	s := NewServer(orchestrator.NewEngine(1), prefrontal.New(), mesh.NewRegistry())
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusOK { t.Fatalf("status=%d body=%s", r.Code, r.Body.String()) }
}

func TestPlanValidation(t *testing.T) {
	s := NewServer(nil, nil, nil)
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/plan", nil)
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusBadRequest { t.Fatalf("status=%d", r.Code) }
}
