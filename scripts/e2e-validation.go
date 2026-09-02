package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	timeout := 5 * time.Second
	client := &http.Client{Timeout: timeout}
	ctx := context.Background()
	results := make([]e2eResult, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			first, second := peers[i], peers[j]
			for _, pair := range [2]peerConfig{first, second} {
				var source, target peerConfig
				if pair.Nucleus == first.Nucleus {
					source, target = first, second
				} else {
					source, target = second, first
				}
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
	report := map[string]any{
		"system":          "SOUL",
		"runner":          "N07",
		"protocol":        protocolVersion,
		"contractVersion": contractVersion,
		"generatedAt":     time.Now().UTC().Format(time.RFC3339),
		"overallStatus":   map[bool]string{true: "PASS", false: "FAIL"}[len(results) == pass && len(results) > 0],
		"summary": map[string]any{
			"totalDirectedCalls": len(results),
			"passed":              pass,
			"failed":              len(results) - pass,
			"pairCount":           len(peers) * (len(peers) - 1) / 2,
			"note":                "Runner executes directed HTTP probes for every unordered pair using the canonical Mesh envelope. This proves endpoint, auth, HTTP 200 and correlation invariants; it does not prove that packets originated inside each peer runtime unless each peer independently runs the validator.",
		},
		"results": results,
	}
	out, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(out))
	if pass != len(results) || len(results) == 0 {
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
		"payload":         map[string]any{"probe": "e2e", "sourceRuntime": source.Nucleus},
		"timestamp":       time.Now().UnixMilli(),
		"meta": map[string]any{
			"runtime":  "Orquestrador-",
			"transport": "HTTP",
			"encoding": "json",
			"version":  contractVersion,
			"traceId":  correlationID,
		},
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

var _ = hex.EncodeToString
var _ = hmac.New
var _ = sha256.New
