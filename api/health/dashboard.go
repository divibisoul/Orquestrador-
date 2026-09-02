package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	protocolVersion = "soul-mesh/1"
	contractVersion = "1.1.0"
	maxBodyBytes    = 1 << 20
	historyLimit    = 10
)

type Dashboard struct {
	Client *http.Client
	Peers  map[string]string

	historyMu sync.Mutex
	history   map[string][]healthSample
}

type probeResult struct {
	Nucleus        string `json:"nucleus"`
	Status         string `json:"status"`
	HTTP           int    `json:"http,omitempty"`
	LatencyMs      int64  `json:"latencyMs"`
	CorrelationID  string `json:"correlationId"`
	Capability     string `json:"capability"`
	Error          string `json:"error,omitempty"`
	CureSuggestion string `json:"cureSuggestion,omitempty"`
	CureAction     string `json:"cureAction,omitempty"`
	Trend          string `json:"trend,omitempty"`
}

type healthSample struct {
	At        time.Time `json:"at"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latencyMs"`
}

type dashboardResponse struct {
	System          string         `json:"system"`
	GeneratedAt     string         `json:"generatedAt"`
	Protocol        string         `json:"protocol"`
	ContractVersion string         `json:"contractVersion"`
	OverallStatus   string         `json:"overallStatus"`
	Checks          []probeResult  `json:"checks"`
	History         map[string]any `json:"history"`
	Summary         map[string]any `json:"summary"`
}

func New() *Dashboard {
	d := &Dashboard{
		Client:  &http.Client{Timeout: 5 * time.Second},
		Peers:   loadPeers(),
		history: make(map[string][]healthSample),
	}
	d.loadHistory()
	return d
}

func Handler() http.Handler { return New() }

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET required"})
		return
	}
	if err := authorizeDashboard(r); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	checks := make([]probeResult, 0, len(d.Peers)*2)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for nucleus, baseURL := range d.Peers {
		for _, capability := range []string{"mesh.ping", "mesh.health"} {
			wg.Add(1)
			go func(nucleus, baseURL, capability string) {
				defer wg.Done()
				result := d.probe(r.Context(), nucleus, baseURL, capability)
				d.recordHistory(result)
				result.Trend = d.trend(nucleus)
				mu.Lock()
				checks = append(checks, result)
				mu.Unlock()
			}(nucleus, baseURL, capability)
		}
	}
	wg.Wait()

	healthy, configured, notConfigured := 0, 0, 0
	for _, check := range checks {
		switch check.Status {
		case "not-configured":
			notConfigured++
		default:
			configured++
		}
		if check.Status == "healthy" {
			healthy++
		}
	}

	status := "degraded"
	switch {
	case len(checks) == 0 || notConfigured == len(checks):
		status = "not-configured"
	case healthy == len(checks):
		status = "online"
	}

	writeJSON(w, http.StatusOK, dashboardResponse{
		System:          "SOUL",
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Protocol:        protocolVersion,
		ContractVersion: contractVersion,
		OverallStatus:   status,
		Checks:          checks,
		History:         d.historySummary(),
		Summary: map[string]any{
			"healthy":          healthy,
			"checks":           len(checks),
			"configuredChecks": configured,
			"notConfigured":    notConfigured,
			"nuclei":           len(d.Peers),
			"historySamples":   historyLimit,
			"source":           "N07",
			"transport":        "HTTP/Soul Mesh",
		},
	})
}

func (d *Dashboard) probe(ctx context.Context, nucleus, baseURL, capability string) probeResult {
	correlationID := fmt.Sprintf("n07-dashboard-%s-%d", nucleus, time.Now().UnixNano())
	result := probeResult{Nucleus: nucleus, Capability: capability, CorrelationID: correlationID}

	if strings.TrimSpace(baseURL) == "" {
		result.Status = "not-configured"
		result.CureSuggestion = fmt.Sprintf("Configure SOUL_MESH_PEERS environment variable with URLs for N01-N06 (ex: {\"N01\":\"http://...\", ...}); currently %s has no endpoint.", nucleus)
		result.CureAction = "configure_peers"
		return result
	}

	message := map[string]any{
		"protocol":        protocolVersion,
		"contractVersion": contractVersion,
		"id":              correlationID,
		"correlationId":   correlationID,
		"source":          "N07",
		"target":          nucleus,
		"kind":            "request",
		"capability":      capability,
		"payload":         map[string]any{"probe": "dashboard"},
		"timestamp":       time.Now().UnixMilli(),
		"meta":            map[string]any{"runtime": "Orquestrador-", "transport": "HTTP", "encoding": "json", "version": contractVersion, "traceId": correlationID},
	}

	data, err := json.Marshal(message)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		result.CureSuggestion = "Falha ao serializar a sonda. Corrija o envelope antes de repetir o health check."
		result.CureAction = "repair_protocol_encoder"
		return result
	}

	endpoint := strings.TrimRight(baseURL, "/") + env("SOUL_MESH_ENDPOINT", "/api/soul-mesh")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		result.CureSuggestion = "Endpoint inválido. Corrija a URL do núcleo no inventário SOUL_MESH_<N>_URL."
		result.CureAction = "configure_endpoint"
		return result
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-soul-correlation-id", correlationID)
	if token := strings.TrimSpace(os.Getenv("SOUL_MESH_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	started := time.Now()
	resp, err := d.Client.Do(req)
	result.LatencyMs = time.Since(started).Milliseconds()

	if err != nil {
		result.Status = "unreachable"
		result.Error = err.Error()
		result.CureSuggestion = fmt.Sprintf("%s está offline ou inacessível. Verifique o serviço, a URL, DNS/ingress e faça redeploy/restart controlado.", nucleus)
		result.CureAction = "redeploy_or_restart"
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		result.Status = "invalid-response"
		result.HTTP = resp.StatusCode
		result.Error = "response is not valid JSON"
		result.CureSuggestion = "O endpoint respondeu sem envelope JSON válido. Verifique a rota Soul Mesh e o adaptador do runtime."
		result.CureAction = "repair_protocol_adapter"
		return result
	}

	returnedCorrelation, _ := reply["correlationId"].(string)
	returnedSource, _ := reply["source"].(string)
	returnedTarget, _ := reply["target"].(string)
	valid := reply["protocol"] == protocolVersion && reply["contractVersion"] == contractVersion && returnedCorrelation == correlationID && returnedSource == nucleus && returnedTarget == "N07"
	result.HTTP = resp.StatusCode

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		result.Status = "unauthorized"
		result.CureSuggestion = "Falha de autenticação. Verifique se SOUL_MESH_TOKEN está presente e sincronizado no núcleo e no N07; não substitua a autenticação por acesso anônimo."
		result.CureAction = "resync_secret"
	case resp.StatusCode == http.StatusNotFound:
		result.Status = "route-not-found"
		result.CureSuggestion = "Rota Soul Mesh não encontrada. Verifique SOUL_MESH_ENDPOINT e o adaptador HTTP do núcleo; preserve o runtime existente."
		result.CureAction = "repair_route_adapter"
	case resp.StatusCode == http.StatusConflict:
		result.Status = "replay-or-conflict"
		result.CureSuggestion = "Conflito/replay detectado. Gere um correlationId e messageId novos e valide janela de nonce/replay."
		result.CureAction = "refresh_message_identity"
	case resp.StatusCode >= 500:
		result.Status = "remote-error"
		result.CureSuggestion = "Erro 5xx remoto. Verifique logs/telemetria do núcleo antes de repetir; use circuit breaker para repetição controlada."
		result.CureAction = "inspect_peer_and_redeploy"
	case resp.StatusCode == http.StatusOK && valid:
		result.Status = "healthy"
	case resp.StatusCode >= 400:
		result.Status = "contract-error"
		result.CureSuggestion = "Rejeição de contrato. Verifique protocol=soul-mesh/1, contractVersion=1.1.0, source/target e correlationId sem forçar migração para 1.2.0."
		result.CureAction = "repair_protocol_adapter"
	default:
		result.Status = "invalid-response"
		result.CureSuggestion = "Resposta HTTP fora do contrato esperado. Compare envelope, status e correlationId com a especificação canônica."
		result.CureAction = "repair_protocol_adapter"
	}

	return result
}

func (d *Dashboard) recordHistory(result probeResult) {
	d.historyMu.Lock()
	defer d.historyMu.Unlock()

	samples := d.history[result.Nucleus]
	samples = append(samples, healthSample{At: time.Now().UTC(), Status: result.Status, LatencyMs: result.LatencyMs})
	if len(samples) > historyLimit {
		samples = samples[len(samples)-historyLimit:]
	}
	d.history[result.Nucleus] = samples
	d.persistHistoryLocked()
}

func (d *Dashboard) trend(nucleus string) string {
	d.historyMu.Lock()
	defer d.historyMu.Unlock()

	samples := d.history[nucleus]
	active := make([]healthSample, 0, len(samples))
	for _, sample := range samples {
		if sample.Status != "not-configured" && sample.LatencyMs > 0 {
			active = append(active, sample)
		}
	}
	if len(active) < 2 {
		return "insufficient-active-history"
	}

	first := active[0]
	last := active[len(active)-1]
	change := ((float64(last.LatencyMs) - float64(first.LatencyMs)) / float64(first.LatencyMs)) * 100
	elapsed := last.At.Sub(first.At).Hours()
	if elapsed > 0 && change >= 5 {
		return fmt.Sprintf("latency-increasing %.1f%% over %.1fh", change, elapsed)
	}
	if change <= -5 {
		return fmt.Sprintf("latency-improving %.1f%% over %.1fh", -change, elapsed)
	}
	if last.Status != "healthy" {
		return "degraded"
	}
	return fmt.Sprintf("stable %.1f%% latency change", change)
}

func (d *Dashboard) historySummary() map[string]any {
	d.historyMu.Lock()
	defer d.historyMu.Unlock()

	out := make(map[string]any, len(d.history))
	for nucleus, samples := range d.history {
		copySamples := append([]healthSample(nil), samples...)
		out[nucleus] = map[string]any{"samples": copySamples, "count": len(copySamples)}
	}
	return out
}

func (d *Dashboard) loadHistory() {
	path := historyPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var history map[string][]healthSample
	if err := json.Unmarshal(data, &history); err != nil {
		return
	}
	for nucleus, samples := range history {
		if len(samples) > historyLimit {
			samples = samples[len(samples)-historyLimit:]
		}
		d.history[nucleus] = samples
	}
}

func (d *Dashboard) persistHistoryLocked() {
	path := historyPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	data, err := json.MarshalIndent(d.history, "", "  ")
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

func historyPath() string {
	return env("SOUL_HEALTH_HISTORY_FILE", ".soul/health-history.json")
}

func authorizeDashboard(r *http.Request) error {
	expected := strings.TrimSpace(os.Getenv("N07_APP_TOKEN"))
	if expected == "" {
		return fmt.Errorf("N07_APP_TOKEN is not configured")
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authorization) < 7 || !strings.EqualFold(authorization[:7], "bearer ") {
		return fmt.Errorf("Bearer authentication required")
	}
	provided := strings.TrimSpace(authorization[7:])
	if provided == "" || provided != expected {
		return fmt.Errorf("invalid application token")
	}
	return nil
}

func loadPeers() map[string]string {
	peers := map[string]string{
		"N01": "",
		"N02": "",
		"N03": "",
		"N04": "",
		"N05": "",
		"N06": "",
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(env("SOUL_MESH_PEERS", "{}")), &raw); err != nil {
		return peers
	}
	for nucleus, value := range raw {
		switch item := value.(type) {
		case string:
			if strings.TrimSpace(item) != "" {
				peers[nucleus] = strings.TrimRight(strings.TrimSpace(item), "/")
			}
		case map[string]any:
			if url, ok := item["url"].(string); ok && strings.TrimSpace(url) != "" {
				peers[nucleus] = strings.TrimRight(strings.TrimSpace(url), "/")
			}
		}
	}
	return peers
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
