package backend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

type Config struct {
	AppToken               string
	CORSOrigins            []string
	MaxRequestBytes        int64
	MaxUploadBytes         int64
	SupabaseURL            string
	SupabaseServiceKey     string
	SupabaseRunsTable      string
	SupabaseArtifactsTable string
	Web3StorageURL         string
	Web3StorageToken       string
	IPFSGatewayURL         string
	RequestTimeout         time.Duration
}

func DefaultConfig() Config {
	maxRequest := int64(2 << 20)
	maxUpload := int64(100 << 20)
	return Config{
		AppToken:               strings.TrimSpace(getenv("N07_APP_TOKEN")),
		CORSOrigins:            splitCSV(getenv("N07_CORS_ORIGINS")),
		MaxRequestBytes:        envInt64("N07_MAX_REQUEST_BYTES", maxRequest),
		MaxUploadBytes:         envInt64("N07_MAX_UPLOAD_BYTES", maxUpload),
		SupabaseURL:             strings.TrimRight(strings.TrimSpace(getenv("SUPABASE_URL")), "/"),
		SupabaseServiceKey:      strings.TrimSpace(getenv("SUPABASE_SERVICE_ROLE_KEY")),
		SupabaseRunsTable:       envString("SUPABASE_RUNS_TABLE", "n07_runs"),
		SupabaseArtifactsTable:  envString("SUPABASE_ARTIFACTS_TABLE", "n07_artifacts"),
		Web3StorageURL:          strings.TrimRight(envString("WEB3_STORAGE_API_URL", "https://api.web3.storage"), "/"),
		Web3StorageToken:        strings.TrimSpace(getenv("WEB3_STORAGE_TOKEN")),
		IPFSGatewayURL:          strings.TrimRight(envString("N07_IPFS_GATEWAY_URL", "https://dweb.link/ipfs"), "/"),
		RequestTimeout:          envDuration("N07_BACKEND_TIMEOUT", 30*time.Second),
	}
}

type Server struct {
	Engine  *orchestrator.Engine
	Config  Config
	Store   *SupabaseStore
	Storage *Web3Storage
}

func New(engine *orchestrator.Engine, cfg Config) *Server {
	s := &Server{
		Engine:  engine,
		Config:  cfg,
		Store:   NewSupabaseStore(cfg),
		Storage: NewWeb3Storage(cfg),
	}
	if engine != nil {
		s.ensureSuperGPUOperations()
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/capabilities", s.capabilities)
	mux.HandleFunc("/v1/health", s.health)
	mux.HandleFunc("/v1/execute", s.execute)
	mux.HandleFunc("/v1/intent", s.intent)
	mux.HandleFunc("/v1/storage/upload", s.upload)
	mux.HandleFunc("/v1/storage/status/", s.storageStatus)
	mux.HandleFunc("/v1/storage/object/", s.storageObject)
	return withCORS(s.Config.CORSOrigins, withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.authorize(r); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
			return
		}
		mux.ServeHTTP(w, r)
	})))
}

func (s *Server) authorize(r *http.Request) error {
	if r.Method == http.MethodOptions {
		return nil
	}
	if strings.TrimSpace(s.Config.AppToken) == "" {
		return errors.New("N07_APP_TOKEN is not configured")
	}
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(h) < 7 || !strings.EqualFold(h[:7], "bearer ") {
		return errors.New("Bearer authentication required")
	}
	if !secureEqual(strings.TrimSpace(h[7:]), s.Config.AppToken) {
		return errors.New("invalid application token")
	}
	return nil
}

type executeRequest struct {
	Operation     string            `json:"operation"`
	Payload       []float64         `json:"payload"`
	Metadata      map[string]string `json:"metadata"`
	CorrelationID string            `json:"correlationId"`
}

func (s *Server) execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"})
		return
	}
	var req executeRequest
	if err := decodeJSON(r, s.Config.MaxRequestBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Operation) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "operation is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.Config.RequestTimeout)
	defer cancel()
	result, err := s.Engine.Execute(ctx, req.Operation, req.Payload, req.Metadata)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, result)
		return
	}
	_ = s.Store.RecordRun(r.Context(), map[string]any{
		"trace_id":       result.TraceID,
		"correlation_id": result.CorrelationID,
		"source":         result.Source,
		"operation":      req.Operation,
		"status":         result.Status,
		"payload":        req.Payload,
		"result":         result.Payload,
		"error":          result.Error,
		"metadata":       req.Metadata,
	})
	writeJSON(w, http.StatusOK, result)
}

type intentRequest struct {
	Tool          string         `json:"tool"`
	Input         map[string]any `json:"input"`
	CorrelationID string         `json:"correlationId"`
}

func (s *Server) intent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"})
		return
	}
	var req intentRequest
	if err := decodeJSON(r, s.Config.MaxRequestBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	toolName := strings.TrimSpace(req.Tool)
	if toolName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "tool is required"})
		return
	}
	values, metadata, err := mapIntent(toolName, req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.Config.RequestTimeout)
	defer cancel()
	result, err := s.Engine.Execute(ctx, operationForTool(toolName), values, metadata)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, result)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tool":          toolName,
		"operation":     operationForTool(toolName),
		"traceId":       result.TraceID,
		"correlationId": result.CorrelationID,
		"status":        result.Status,
		"payload":       result.Payload,
		"metadata":      result.Metadata,
	})
}

func operationForTool(tool string) string {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "neural.forward":
		return "neural.forward@1.0.0"
	case "neural.learn":
		return "neural.learn@1.0.0"
	case "compute.execute":
		return "compute.execute@1.0.0"
	case "cognitive.execute":
		return "cognitive.execute@1.0.0"
	case "supergpu.execute":
		return "supergpu.execute@1.0.0"
	case "supergpu.parallel":
		return "supergpu.parallel@1.0.0"
	case "supergpu.describe":
		return "supergpu.describe@1.0.0"
	case "supergpu.memory":
		return "supergpu.memory@1.0.0"
	default:
		return tool
	}
}

func mapIntent(tool string, input map[string]any) ([]float64, map[string]string, error) {
	metadata := map[string]string{}
	if input == nil {
		input = map[string]any{}
	}
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "neural.forward":
		return numberArray(input["values"])
	case "neural.learn":
		in, err := numberArrayOnly(input["input"])
		if err != nil {
			return nil, nil, err
		}
		target, err := numberArrayOnly(input["target"])
		if err != nil {
			return nil, nil, err
		}
		return append(in, target...), metadata, nil
	case "compute.execute", "cognitive.execute", "supergpu.execute":
		values, err := numberArrayOnly(input["values"])
		if err != nil {
			return nil, nil, err
		}
		if op, ok := input["operation"].(string); ok && strings.TrimSpace(op) != "" {
			metadata["operation"] = strings.TrimSpace(op)
		}
		if device, ok := input["device"].(string); ok && strings.TrimSpace(device) != "" {
			metadata["device"] = strings.TrimSpace(device)
		}
		return values, metadata, nil
	case "supergpu.parallel":
		batch, err := json.Marshal(input["inputs"])
		if err != nil {
			return nil, nil, errors.New("inputs must be an array of numeric arrays")
		}
		var inputs [][]float64
		if err := json.Unmarshal(batch, &inputs); err != nil || len(inputs) == 0 {
			return nil, nil, errors.New("inputs must be an array of numeric arrays")
		}
		metadata["inputs_json"] = string(batch)
		if op, ok := input["operation"].(string); ok && strings.TrimSpace(op) != "" {
			metadata["operation"] = strings.TrimSpace(op)
		}
		if device, ok := input["device"].(string); ok && strings.TrimSpace(device) != "" {
			metadata["device"] = strings.TrimSpace(device)
		}
		if workers, ok := input["workers"].(float64); ok {
			if workers < 1 || workers > 128 || workers != float64(int(workers)) {
				return nil, nil, errors.New("workers must be an integer between 1 and 128")
			}
			metadata["workers"] = strconv.Itoa(int(workers))
		}
		return []float64{1}, metadata, nil
	default:
		return nil, nil, fmt.Errorf("unsupported tool: %s", tool)
	}
}

func numberArray(v any) ([]float64, map[string]string, error) {
	values, err := numberArrayOnly(v)
	return values, map[string]string{}, err
}

func numberArrayOnly(v any) ([]float64, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New("values must be an array of numbers")
	}
	var values []float64
	if err := json.Unmarshal(encoded, &values); err != nil {
		return nil, errors.New("values must be an array of numbers")
	}
	if len(values) == 0 {
		return nil, errors.New("values must be a non-empty array of numbers")
	}
	return values, nil
}

func (s *Server) capabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET required"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"nucleus": "N07", "operations": s.Engine.Operations(), "storage": map[string]any{"configured": s.Storage.Configured(), "api": "web3.storage-compatible"}, "supabase": map[string]any{"configured": s.Store.Configured()}})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET required"})
		return
	}
	status := s.Engine.Health()
	status["backend"] = map[string]any{"supabase_configured": s.Store.Configured(), "storage_configured": s.Storage.Configured()}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "POST required"})
		return
	}
	if !s.Storage.Configured() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Web3 Storage is not configured"})
		return
	}
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid multipart content type"})
			return
		}
		if err := r.ParseMultipartForm(s.Config.MaxUploadBytes); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "multipart field 'file' is required"})
			return
		}
		defer file.Close()
		_ = params
		cid, size, err := s.Storage.Upload(r.Context(), file, header.Filename)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		_ = s.Store.RecordArtifact(r.Context(), map[string]any{"cid": cid, "filename": header.Filename, "size_bytes": size})
		writeJSON(w, http.StatusOK, map[string]any{"cid": cid, "filename": header.Filename, "size": size, "gateway": s.Storage.ObjectURL(cid)})
		return
	}
	limited := io.LimitReader(r.Body, s.Config.MaxUploadBytes+1)
	cid, size, err := s.Storage.Upload(r.Context(), limited, r.Header.Get("X-Filename"))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	if size > s.Config.MaxUploadBytes {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "upload exceeds configured maximum"})
		return
	}
	_ = s.Store.RecordArtifact(r.Context(), map[string]any{"cid": cid, "filename": r.Header.Get("X-Filename"), "size_bytes": size})
	writeJSON(w, http.StatusOK, map[string]any{"cid": cid, "size": size, "gateway": s.Storage.ObjectURL(cid)})
}

func (s *Server) storageStatus(w http.ResponseWriter, r *http.Request) {
	cid := strings.TrimPrefix(r.URL.Path, "/v1/storage/status/")
	if cid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "cid is required"})
		return
	}
	status, err := s.Storage.Status(r.Context(), cid)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) storageObject(w http.ResponseWriter, r *http.Request) {
	cid := strings.TrimPrefix(r.URL.Path, "/v1/storage/object/")
	if cid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "cid is required"})
		return
	}
	http.Redirect(w, r, s.Storage.ObjectURL(cid), http.StatusFound)
}

func decodeJSON(r *http.Request, limit int64, out any) error {
	if limit <= 0 {
		limit = 2 << 20
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, limit))
	if err := decoder.Decode(out); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain exactly one JSON value")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if strings.TrimSpace(id) == "" {
			id = randomID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func withCORS(origins []string, next http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Filename")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func secureEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func randomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
