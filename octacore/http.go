package octacore

import (
    "encoding/json"
    "errors"
    "net/http"
    "os"
    "strings"
    "time"
)

const maxOctaHTTPBody = 2 << 20

func HTTPHandler(processor *Processor) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/v1/octacore/health", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet { writeOctaJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET required"}); return }
        if err := authorizeOctaHTTP(r); err != nil { writeOctaJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()}); return }
        writeOctaJSON(w, http.StatusOK, processor.Health())
    })
    mux.HandleFunc("/v1/octacore/inventory", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet { writeOctaJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET required"}); return }
        if err := authorizeOctaHTTP(r); err != nil { writeOctaJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()}); return }
        writeOctaJSON(w, http.StatusOK, map[string]any{"slots": processor.Inventory(), "supergpu_connected": processor.SuperGPUConnected()})
    })
    mux.HandleFunc("/v1/octacore/submit", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost { writeOctaJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"}); return }
        if err := authorizeOctaHTTP(r); err != nil { writeOctaJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()}); return }
        var job OctaCoreJob
        decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxOctaHTTPBody))
        if err := decoder.Decode(&job); err != nil { writeOctaJSON(w, http.StatusBadRequest, map[string]any{"error": "INVALID_OCTACORE_JSON", "detail": err.Error()}); return }
        if header := strings.TrimSpace(r.Header.Get("X-Correlation-ID")); header != "" && header != job.CorrelationID { writeOctaJSON(w, http.StatusBadRequest, map[string]any{"error": "CORRELATION_MISMATCH"}); return }
        if err := job.Validate(); err != nil { writeOctaJSON(w, http.StatusBadRequest, map[string]any{"error": "INVALID_OCTACORE_JOB", "detail": err.Error()}); return }
        result := processor.Submit(r.Context(), job)
        status := http.StatusOK
        if !result.OK { status = http.StatusBadGateway }
        writeOctaJSON(w, status, result)
    })
    mux.HandleFunc("/v1/octacore/batch", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost { writeOctaJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"}); return }
        if err := authorizeOctaHTTP(r); err != nil { writeOctaJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()}); return }
        var jobs []OctaCoreJob
        decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxOctaHTTPBody))
        if err := decoder.Decode(&jobs); err != nil { writeOctaJSON(w, http.StatusBadRequest, map[string]any{"error": "INVALID_OCTACORE_BATCH_JSON", "detail": err.Error()}); return }
        if len(jobs) == 0 { writeOctaJSON(w, http.StatusBadRequest, map[string]any{"error": "OCTACORE_BATCH_EMPTY"}); return }
        for i, job := range jobs { if err := job.Validate(); err != nil { writeOctaJSON(w, http.StatusBadRequest, map[string]any{"error": "INVALID_OCTACORE_JOB", "index": i, "detail": err.Error()}); return } }
        writeOctaJSON(w, http.StatusOK, processor.Batch(r.Context(), jobs))
    })
    return mux
}

func authorizeOctaHTTP(r *http.Request) error {
    expected := strings.TrimSpace(os.Getenv("N07_APP_TOKEN"))
    if expected == "" { return errors.New("N07_APP_TOKEN is not configured") }
    auth := strings.TrimSpace(r.Header.Get("Authorization"))
    if len(auth) < 7 || !strings.EqualFold(auth[:7], "bearer ") { return errors.New("Bearer authentication required") }
    if strings.TrimSpace(auth[7:]) != expected { return errors.New("invalid application token") }
    return nil
}

func writeOctaJSON(w http.ResponseWriter, status int, value any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(value)
}

func octaHTTPTimeout(timeout time.Duration) time.Duration {
    if timeout <= 0 { return 15 * time.Second }
    return timeout
}
