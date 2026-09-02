package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	protocolVersion = "soul-mesh/1"
	contractVersion = "1.1.0"
	maxBodyBytes    = 1 << 20
)

type Dashboard struct {
	Client *http.Client
	Peers  map[string]string
}

type probeResult struct {
	Nucleus       string `json:"nucleus"`
	Status        string `json:"status"`
	HTTP          int    `json:"http,omitempty"`
	LatencyMs     int64  `json:"latencyMs"`
	CorrelationID string `json:"correlationId"`
	Capability    string `json:"capability"`
	Error         string `json:"error,omitempty"`
}

type dashboardResponse struct {
	System          string         `json:"system"`
	GeneratedAt     string         `json:"generatedAt"`
	Protocol        string         `json:"protocol"`
	ContractVersion string         `json:"contractVersion"`
	OverallStatus   string         `json:"overallStatus"`
	Checks          []probeResult  `json:"checks"`
	Summary         map[string]any `json:"summary"`
}

func New() *Dashboard {
	return &Dashboard{Client: &http.Client{Timeout: 5 * time.Second}, Peers: loadPeers()}
}

func Handler() http.Handler { return New() }

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET required"})
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
				mu.Lock()
				checks = append(checks, result)
				mu.Unlock()
			}(nucleus, baseURL, capability)
		}
	}
	wg.Wait()

	healthy := 0
	configured := 0
	for _, check := range checks {
		if check.Status != "not-configured" {
			configured++
		}
		if check.Status == "healthy" {
			healthy++
		}
	}
	status := "degraded"
	if len(checks) > 0 && healthy == len(checks) {
		status = "online"
	}
	if len(checks) == 0 {
		status = "not-configured"
	}

	writeJSON(w, http.StatusOK, dashboardResponse{
		System:          "SOUL",
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Protocol:        protocolVersion,
		ContractVersion: contractVersion,
		OverallStatus:   status,
		Checks:          checks,
		Summary: map[string]any{
			"healthy":          healthy,
			"checks":           len(checks),
			"configuredChecks": configured,
			"nuclei":           len(d.Peers),
			"source":           "N07",
			"transport":        "HTTP/Soul Mesh",
		},
	})
}

func (d *Dashboard) probe(ctx context.Context, nucleus, baseURL, capability string) probeResult {
	correlationID := fmt.Sprintf("n07-dashboard-%d", time.Now().UnixNano())
	if strings.TrimSpace(baseURL) == "" {
		return probeResult{Nucleus: nucleus, Capability: capability, Status: "not-configured", CorrelationID: correlationID}
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
		return probeResult{Nucleus: nucleus, Capability: capability, Status: "error", CorrelationID: correlationID, Error: err.Error()}
	}
	endpoint := strings.TrimRight(baseURL, "/") + env("SOUL_MESH_ENDPOINT", "/api/soul-mesh")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return probeResult{Nucleus: nucleus, Capability: capability, Status: "error", CorrelationID: correlationID, Error: err.Error()}
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-soul-correlation-id", correlationID)
	if token := strings.TrimSpace(os.Getenv("SOUL_MESH_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	started := time.Now()
	resp, err := d.Client.Do(req)
	latency := time.Since(started).Milliseconds()
	if err != nil {
		return probeResult{Nucleus: nucleus, Capability: capability, Status: "unreachable", LatencyMs: latency, CorrelationID: correlationID, Error: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		return probeResult{Nucleus: nucleus, Capability: capability, Status: "invalid-response", HTTP: resp.StatusCode, LatencyMs: latency, CorrelationID: correlationID, Error: "response is not valid JSON"}
	}
	returnedCorrelation, _ := reply["correlationId"].(string)
	returnedSource, _ := reply["source"].(string)
	returnedTarget, _ := reply["target"].(string)
	valid := reply["protocol"] == protocolVersion && reply["contractVersion"] == contractVersion && returnedCorrelation == correlationID && returnedSource == nucleus && returnedTarget == "N07"
	status := "invalid-response"
	if resp.StatusCode == http.StatusOK && valid {
		status = "healthy"
	}
	if resp.StatusCode == http.StatusUnauthorized {
		status = "unauthorized"
	}
	return probeResult{Nucleus: nucleus, Capability: capability, Status: status, HTTP: resp.StatusCode, LatencyMs: latency, CorrelationID: correlationID}
}

func loadPeers() map[string]string {
	peers := make(map[string]string)
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
