package octacore

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestOctaCoreHTTPSubmitUsesExistingSuperGPU(t *testing.T) {
	old := os.Getenv("N07_APP_TOKEN")
	t.Cleanup(func() { _ = os.Setenv("N07_APP_TOKEN", old) })
	_ = os.Setenv("N07_APP_TOKEN", "octa-test-token")
	runtime := supergpu.New(nil)
	runtime.Discover()
	processor, err := NewProcessorWithRuntime(DefaultSchedulerConfig(), nil, runtime)
	if err != nil {
		t.Fatal(err)
	}

	job := OctaCoreJob{
		JobID: "http-job-1", CorrelationID: "http-corr-1", Kind: KindCustom,
		Source: G7, Target: "G7", BackendPrefs: []Backend{BackendInProcess},
		Payload:  map[string]any{"operation": "identity", "values": []float64{1, 2, 3}},
		Priority: 90, TTLMS: 5000,
	}
	body, _ := json.Marshal(job)
	req := httptest.NewRequest(http.MethodPost, "/v1/octacore/submit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer octa-test-token")
	req.Header.Set("X-Correlation-ID", "http-corr-1")
	rec := httptest.NewRecorder()
	HTTPHandler(processor).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d body=%s", rec.Code, rec.Body.String())
	}
	var result OctaCoreResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.BackendUsed != string(BackendInProcess) {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.CorrelationID != job.CorrelationID {
		t.Fatalf("correlation lost: %s", result.CorrelationID)
	}
}

func TestOctaCoreHTTPRejectsCorrelationMismatch(t *testing.T) {
	old := os.Getenv("N07_APP_TOKEN")
	t.Cleanup(func() { _ = os.Setenv("N07_APP_TOKEN", old) })
	_ = os.Setenv("N07_APP_TOKEN", "octa-test-token")
	processor, err := NewProcessorWithRuntime(DefaultSchedulerConfig(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(OctaCoreJob{JobID: "job", CorrelationID: "job-corr", Kind: KindCustom, Source: G7, Target: "G7", BackendPrefs: []Backend{BackendInProcess}, Payload: map[string]any{"operation": "identity", "values": []float64{1}}, Priority: 1, TTLMS: 1000})
	req := httptest.NewRequest(http.MethodPost, "/v1/octacore/submit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer octa-test-token")
	req.Header.Set("X-Correlation-ID", "wrong-corr")
	rec := httptest.NewRecorder()
	HTTPHandler(processor).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
