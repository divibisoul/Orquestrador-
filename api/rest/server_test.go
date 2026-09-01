package rest

import (
	"encoding/json"
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
	if r.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", r.Code, r.Body.String())
	}
}

func TestMetrics(t *testing.T) {
	s := NewServer(nil, nil, nil)
	s.Metrics.Record("test", 5, false)
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", r.Code, r.Body.String())
	}
	var snapshot map[string]struct {
		Tasks            uint64  `json:"tasks"`
		Failures         uint64  `json:"failures"`
		TotalLatencyMS   float64 `json:"total_latency_ms"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	metric, ok := snapshot["test"]
	if !ok || metric.Tasks != 1 || metric.Failures != 0 || metric.TotalLatencyMS != 5 {
		t.Fatalf("metrics=%+v", snapshot)
	}
}

func TestPlanValidation(t *testing.T) {
	s := NewServer(nil, nil, nil)
	r := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/plan", nil)
	s.Handler().ServeHTTP(r, req)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", r.Code)
	}
}
