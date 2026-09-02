package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type peerConfig struct {
	Nucleus string
	URL     string
}

type e2eResult struct {
	Source        string `json:"source"`
	Target        string `json:"target"`
	Status        string `json:"status"`
	HTTP          int    `json:"http,omitempty"`
	LatencyMs     int64  `json:"latencyMs"`
	CorrelationID string `json:"correlationId"`
	Error         string `json:"error,omitempty"`
}

const (
	protocolVersion = "soul-mesh/1"
	contractVersion = "1.1.0"
)

func main() {
	peers := loadPeers()
	if len(peers) < 2 {
		fmt.Println(`{"system":"SOUL","runner":"N07","overallStatus":"FAIL","error":"at least two configured Mesh peers are required in SOUL_MESH_PEERS"}`)
		os.Exit(2)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	ctx := context.Background()
	results := make([]e2eResult, 0, len(peers)*(len(peers)-1))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			first, second := peers[i], peers[j]
			for _, direction := range [2][2]peerConfig{{first, second}, {second, first}} {
				source, target := direction[0], direction[1]
				wg.Add(1)
				go func(source, target peerConfig) {
					defer wg.Done()
					result := executeProbe(ctx, client, source, target)
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
				}(source, target)
			}
		}
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Source != results[j].Source {
			return results[i].Source < results[j].Source
		}
		return results[i].Target < results[j].Target
	})
	pass := 0
	for _, result := range results {
		if result.Status == "pass" {
			pass++
		}
	}
	status := "FAIL"
	if pass == len(results) {
		status = "PASS"
	}
	report := map[string]any{
		"system":          "SOUL",
		"runner":          "N07",
		"protocol":        protocolVersion,
		"contractVersion": contractVersion,
		"generatedAt":     time.Now().UTC().Format(time.RFC3339),
		"overallStatus":   status,
		"summary": map[string]any{
			"totalDirectedCalls": len(results),
			"passed":             pass,
			"failed":             len(results) - pass,
			"pairCount":          len(peers) * (len(peers) - 1) / 2,
			"note":               "N07 executes the direct HTTP probe for every unordered pair in both directions. This validates the canonical endpoint, HTTP 200, response identity and correlation invariants. It is not proof of physical packet origin from the declared source runtime; source-origin proof requires each runtime to execute its own validator.",
		},
		"results": results,
	}
	out, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(out))
	if status != "PASS" {
		os.Exit(2)
	}
}

func executeProbe(ctx context.Context, client *http.Client, source, target peerConfig) e2eResult {
	correlationID := fmt.Sprintf("n07-e2e-%s-%s-%d", source.Nucleus, target.Nucleus, time.Now().UnixNano())
	message := map[string]any{
		"protocol":        protocolVersion,
		"contractVersion": contractVersion,
		"id":              correlationID,
		"correlationId":   correlationID,
		"source":          source.Nucleus,
		"target":          target.Nucleus,
		"kind":            "request",
		"capability":      "mesh.ping",
		"payload":         map[string]any{"probe": "e2e"},
		"timestamp":       time.Now().UnixMilli(),
		"meta":            map[string]any{"runtime": "Orquestrador-", "transport": "HTTP", "encoding": "json", "version": contractVersion, "traceId": correlationID},
	}
	data, err := json.Marshal(message)
	if err != nil {
		return e2eResult{Source: source.Nucleus, Target: target.Nucleus, Status: "fail", CorrelationID: correlationID, Error: err.Error()}
	}
	endpoint := strings.TrimRight(target.URL, "/") + env("SOUL_MESH_ENDPOINT", "/api/soul-mesh")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return e2eResult{Source: source.Nucleus, Target: target.Nucleus, Status: "fail", CorrelationID: correlationID, Error: err.Error()}
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-soul-correlation-id", correlationID)
	if token := strings.TrimSpace(os.Getenv("SOUL_MESH_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	started := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(started).Milliseconds()
	if err != nil {
		return e2eResult{Source: source.Nucleus, Target: target.Nucleus, Status: "fail", LatencyMs: latency, CorrelationID: correlationID, Error: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var reply map[string]any
	if err := json.Unmarshal(body, &reply); err != nil {
		return e2eResult{Source: source.Nucleus, Target: target.Nucleus, Status: "fail", HTTP: resp.StatusCode, LatencyMs: latency, CorrelationID: correlationID, Error: "invalid JSON response"}
	}
	returnedCorrelation, _ := reply["correlationId"].(string)
	returnedSource, _ := reply["source"].(string)
	returnedTarget, _ := reply["target"].(string)
	valid := resp.StatusCode == http.StatusOK && reply["protocol"] == protocolVersion && reply["contractVersion"] == contractVersion && returnedCorrelation == correlationID && returnedSource == target.Nucleus && returnedTarget == source.Nucleus
	status := "fail"
	if valid {
		status = "pass"
	}
	return e2eResult{Source: source.Nucleus, Target: target.Nucleus, Status: status, HTTP: resp.StatusCode, LatencyMs: latency, CorrelationID: correlationID}
}

func loadPeers() []peerConfig {
	var raw map[string]any
	if err := json.Unmarshal([]byte(env("SOUL_MESH_PEERS", "{}")), &raw); err != nil {
		return nil
	}
	peers := make([]peerConfig, 0, len(raw))
	for nucleus, value := range raw {
		url := ""
		switch item := value.(type) {
		case string:
			url = item
		case map[string]any:
			url, _ = item["url"].(string)
		}
		if strings.TrimSpace(url) != "" {
			peers = append(peers, peerConfig{Nucleus: nucleus, URL: strings.TrimRight(strings.TrimSpace(url), "/")})
		}
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].Nucleus < peers[j].Nucleus })
	return peers
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
