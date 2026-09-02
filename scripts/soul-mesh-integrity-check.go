package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type result struct {
	Peer          string `json:"peer"`
	Capability    string `json:"capability"`
	Status        string `json:"status"`
	HTTP          int    `json:"http,omitempty"`
	LatencyMs     int64  `json:"latencyMs,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
	Error         string `json:"error,omitempty"`
}

func main() {
	peers := []string{"N01", "N02", "N03", "N04", "N05", "N06"}
	var configured map[string]any
	if err := json.Unmarshal([]byte(env("SOUL_MESH_PEERS", "{}")), &configured); err != nil {
		panic("SOUL_MESH_PEERS_INVALID_JSON")
	}
	endpoint := env("SOUL_MESH_ENDPOINT", "/api/soul-mesh")
	timeout := 5 * time.Second
	results := make([]result, 0, len(peers)*3)
	client := &http.Client{Timeout: timeout}
	for _, peer := range peers {
		for _, capability := range []string{"mesh.ping", "mesh.health", "mesh.describe"} {
			base := peerURL(configured, peer)
			if base == "" {
				results = append(results, result{Peer: peer, Capability: capability, Status: "not-configured"})
				continue
			}
			correlationID := fmt.Sprintf("n07-check-%d", time.Now().UnixNano())
			message := map[string]any{"protocol": "soul-mesh/1", "contractVersion": "1.1.0", "id": correlationID, "correlationId": correlationID, "source": "N07", "target": peer, "kind": "request", "capability": capability, "payload": map[string]any{"probe": "integrity"}, "timestamp": time.Now().UnixMilli(), "meta": map[string]any{"runtime": "Orquestrador-", "transport": "HTTP", "encoding": "json", "version": "1.1.0", "traceId": correlationID}}
			data, _ := json.Marshal(message)
			started := time.Now()
			req, _ := http.NewRequest(http.MethodPost, strings.TrimRight(base, "/")+endpoint, bytes.NewReader(data))
			req.Header.Set("content-type", "application/json")
			req.Header.Set("x-soul-correlation-id", correlationID)
			resp, err := client.Do(req)
			if err != nil {
				results = append(results, result{Peer: peer, Capability: capability, Status: "unreachable", LatencyMs: time.Since(started).Milliseconds(), CorrelationID: correlationID, Error: err.Error()})
				continue
			}
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			resp.Body.Close()
			var reply map[string]any
			_ = json.Unmarshal(body, &reply)
			valid := reply["protocol"] == "soul-mesh/1" && reply["contractVersion"] == "1.1.0" && reply["correlationId"] == correlationID && reply["source"] == peer && reply["target"] == "N07"
			status := "invalid-response"
			if resp.StatusCode >= 200 && resp.StatusCode < 300 && valid {
				status = "healthy"
			}
			results = append(results, result{Peer: peer, Capability: capability, Status: status, HTTP: resp.StatusCode, LatencyMs: time.Since(started).Milliseconds(), CorrelationID: correlationID})
		}
	}
	healthy := 0
	checked := 0
	for _, r := range results {
		if r.Status != "not-configured" {
			checked++
		}
		if r.Status == "healthy" {
			healthy++
		}
	}
	report := map[string]any{"system": "SOUL", "nucleus": "N07", "protocol": "soul-mesh/1", "contractVersion": "1.1.0", "generatedAt": time.Now().UTC().Format(time.RFC3339), "checks": results, "summary": map[string]any{"healthy": healthy, "checked": checked, "totalPeers": len(peers)}}
	out, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(out))
	if strings.EqualFold(env("SOUL_MESH_REQUIRE_ALL", "false"), "true") && (healthy != checked || checked != len(peers)*3) {
		os.Exit(2)
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
func peerURL(configured map[string]any, id string) string {
	v, ok := configured[id]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimRight(s, "/")
	}
	if m, ok := v.(map[string]any); ok {
		if s, ok := m["url"].(string); ok {
			return strings.TrimRight(s, "/")
		}
	}
	return ""
}
