package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

func TestCallBestDynamicHonorsOSSAffinityAtRuntime(t *testing.T) {
	secret := "test-secret-oss-affinity-2026"
	var n02Tasks atomic.Int32
	var n03Tasks atomic.Int32

	newPeer := func(nucleus string, taskCount *atomic.Int32) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var request canonicalWireEnvelope
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("%s decode request: %v", nucleus, err)
			}

			responsePayload := map[string]any{}
			if request.Capability == "mesh.discovery" || request.Capability == "mesh.describe" {
				responsePayload["executableCapabilities"] = []any{"mesh.discovery", "mesh.describe", "multimodal_cortex"}
			} else if request.Capability == "multimodal_cortex" {
				taskCount.Add(1)
				responsePayload["handledBy"] = nucleus
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
				t.Fatalf("%s sign response: %v", nucleus, err)
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
	}

	n02 := newPeer(protocol.N02, &n02Tasks)
	defer n02.Close()
	n03 := newPeer(protocol.N03, &n03Tasks)
	defer n03.Close()

	client := &PeerClient{
		peers: map[string]PeerInfo{
			protocol.N02: {Nucleus: protocol.N02, URL: n02.URL, Circuit: CircuitClosed},
			protocol.N03: {Nucleus: protocol.N03, URL: n03.URL, Circuit: CircuitClosed},
		},
		client:            &http.Client{Timeout: 5 * time.Second},
		secret:            secret,
		maxRetry:          1,
		cooldown:          time.Second,
		discoveryCache:    make(map[string]discoveryCacheEntry),
		discoveryCacheTTL: time.Minute,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, owner, err := client.CallBestDynamic(ctx, "multimodal_cortex", map[string]any{"text": "probe"}, "oss-affinity-e2e")
	if err != nil {
		t.Fatalf("CallBestDynamic: %v", err)
	}
	if owner != protocol.N03 {
		t.Fatalf("affinity route selected %s, expected %s", owner, protocol.N03)
	}
	if n03Tasks.Load() != 1 {
		t.Fatalf("N03 task count = %d, expected 1", n03Tasks.Load())
	}
	if n02Tasks.Load() != 0 {
		t.Fatalf("N02 task count = %d, expected 0 because N03 is the first affinity target", n02Tasks.Load())
	}
}
