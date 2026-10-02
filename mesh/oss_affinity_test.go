package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestLoadOSSAffinityFromBytes(t *testing.T) {
	data := []byte(`{
		"ossAffinityRouting": {
			"schemaVersion": "1.0.0",
			"capabilities": {
				"multimodal_cortex": ["N03", "N02", "N06"]
			}
		}
	}`)
	got, err := loadOSSAffinityFromBytes(data)
	if err != nil {
		t.Fatalf("load affinity: %v", err)
	}
	if got["multimodal_cortex"][0] != "N03" {
		t.Fatalf("unexpected first target: %#v", got["multimodal_cortex"])
	}
}

func TestCallBestDynamicRoutesAllOSSCapabilitiesByAffinity(t *testing.T) {
	affinity := loadOSSAffinityFile()
	if len(affinity) == 0 {
		t.Fatal("canonical OSS affinity manifest was not loaded")
	}

	secret := "test-secret-oss-affinity-2026"
	var mu sync.Mutex
	taskCounts := map[string]map[string]int{}
	servers := make(map[string]*httptest.Server, 6)

	for _, nucleus := range []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06} {
		taskCounts[nucleus] = map[string]int{}
		nucleus := nucleus
		servers[nucleus] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var request canonicalWireEnvelope
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("%s decode request: %v", nucleus, err)
				return
			}
			responsePayload := map[string]any{}
			if request.Capability == "mesh.discovery" || request.Capability == "mesh.describe" {
				capabilities := []any{"mesh.discovery", "mesh.describe"}
				for capability := range affinity {
					capabilities = append(capabilities, capability)
				}
				responsePayload["executableCapabilities"] = capabilities
			} else {
				mu.Lock()
				taskCounts[nucleus][request.Capability]++
				mu.Unlock()
				responsePayload["handledBy"] = nucleus
				responsePayload["capability"] = request.Capability
			}
			envelope := protocol.MeshEnvelope{
				Version:         protocol.SoulMeshVersion,
				ContractVersion: protocol.SoulMeshContractVersion,
				MessageID:       protocol.NewTraceID(),
				Source:          nucleus,
				Target:          protocol.N07,
				Timestamp:       time.Now().UnixMilli(),
				Nonce:           protocol.NewTraceID(),
				CorrelationID:   request.CorrelationID,
				Type:            "TASK_RESULT",
				Payload: map[string]any{
					"capability": request.Capability,
					"payload":    responsePayload,
				},
			}
			if err := protocol.SignHMAC(&envelope, secret); err != nil {
				t.Errorf("%s sign response: %v", nucleus, err)
				return
			}
			body := map[string]any{
				"protocol":        "soul-mesh/1",
				"contractVersion": protocol.SoulMeshContractVersion,
				"id":              envelope.MessageID,
				"source":          envelope.Source,
				"target":          envelope.Target,
				"timestamp":       envelope.Timestamp,
				"nonce":           envelope.Nonce,
				"correlationId":   envelope.CorrelationID,
				"kind":            "response",
				"capability":      request.Capability,
				"hmac":            envelope.HMAC,
				"payload":         responsePayload,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
		}))
		defer servers[nucleus].Close()
	}

	client := &PeerClient{
		peers:             map[string]PeerInfo{},
		client:            &http.Client{Timeout: 5 * time.Second},
		secret:            secret,
		maxRetry:          1,
		cooldown:          time.Second,
		discoveryCache:    make(map[string]discoveryCacheEntry),
		discoveryCacheTTL: time.Minute,
	}
	for nucleus, server := range servers {
		client.peers[nucleus] = PeerInfo{Nucleus: nucleus, URL: server.URL, Circuit: CircuitClosed}
	}

	for capability := range affinity {
		targets := affinity[capability]
		expected := ""
		for _, target := range targets {
			if _, ok := client.peers[target]; ok {
				expected = target
				break
			}
		}
		if expected == "" {
			t.Fatalf("%s has no routable N01..N06 affinity target: %#v", capability, targets)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, owner, err := client.CallBestDynamic(ctx, capability, map[string]any{"probe": capability}, "oss-affinity-"+capability)
		cancel()
		if err != nil {
			t.Fatalf("%s CallBestDynamic: %v", capability, err)
		}
		if owner != expected {
			t.Fatalf("%s routed to %s, expected affinity target %s from %#v", capability, owner, expected, targets)
		}
		mu.Lock()
		got := taskCounts[owner][capability]
		mu.Unlock()
		if got != 1 {
			t.Fatalf("%s task count on %s = %d, expected 1", capability, owner, got)
		}
	}

	var totalTasks int
	mu.Lock()
	for _, counts := range taskCounts {
		for _, count := range counts {
			totalTasks += count
		}
	}
	mu.Unlock()
	if totalTasks != len(affinity) {
		t.Fatalf("executed OSS capabilities = %d, manifest entries = %d", totalTasks, len(affinity))
	}
}
