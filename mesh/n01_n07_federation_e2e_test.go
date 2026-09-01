package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

const federationE2ESecret = "n07-e2e-secret-0123456789abcdef"

func TestN01ToN07FederatesAcrossN04N05N06(t *testing.T) {
	for _, key := range []string{"SOUL_MESH_N04_URL", "SOUL_MESH_N05_URL", "SOUL_MESH_N06_URL", "SOUL_MESH_HMAC_SECRET", "N07_MESH_ALLOW_UNAUTH_LOCAL"} {
		t.Setenv(key, "")
	}
	t.Setenv("SOUL_MESH_HMAC_SECRET", federationE2ESecret)

	servers := make(map[string]*httptest.Server, 3)
	var mu sync.Mutex
	for _, nucleus := range []string{"N04", "N05", "N06"} {
		nucleus := nucleus
		servers[nucleus] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var in protocol.MeshEnvelope
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := protocol.VerifyHMAC(in, federationE2ESecret, time.Now()); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			capability := in.Capability()
			payload := map[string]any{}
			switch capability {
			case "mesh.discovery", "mesh.describe":
				payload = map[string]any{"executableCapabilities": []string{"e2e." + strings.ToLower(nucleus)}}
			case "e2e.n04", "e2e.n05", "e2e.n06":
				payload = map[string]any{"values": []any{float64(len(nucleus)), float64(len(in.Payload))}, "status": "ok"}
			default:
				payload = map[string]any{"error": "unsupported test capability"}
			}
			responseID := protocol.NewTraceID()
			responseNonce := protocol.NewTraceID()
			responseTimestamp := time.Now().UnixMilli()
			envType := "TASK_RESULT"
			if _, ok := payload["error"]; ok {
				envType = "ERROR"
			}
			env := protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: protocol.SoulMeshContractVersion, MessageID: responseID, Source: nucleus, Target: protocol.N07, Timestamp: responseTimestamp, Nonce: responseNonce, CorrelationID: in.CorrelationID, Type: envType, Payload: map[string]any{"capability": capability, "payload": payload}}
			if err := protocol.SignHMAC(&env, federationE2ESecret); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			wirePayload, err := json.Marshal(env)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var roundTrip protocol.MeshEnvelope
			if err := json.Unmarshal(wirePayload, &roundTrip); err != nil {
				http.Error(w, "internal test response JSON round-trip failed", http.StatusInternalServerError)
				return
			}
			if err := protocol.VerifyHMAC(roundTrip, federationE2ESecret, time.Now()); err == nil {
				http.Error(w, "internal test response replay validation unexpectedly succeeded", http.StatusInternalServerError)
				return
			} else if !strings.Contains(err.Error(), "replay detected") {
				http.Error(w, "internal test response HMAC round-trip validation failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			_ = json.NewEncoder(w).Encode(roundTrip)
		}))
	}
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()
	t.Setenv("SOUL_MESH_N04_URL", servers["N04"].URL)
	t.Setenv("SOUL_MESH_N05_URL", servers["N05"].URL)
	t.Setenv("SOUL_MESH_N06_URL", servers["N06"].URL)

	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	gateway := NewEnhancedFederatedHTTPGateway(e)
	for _, nucleus := range []string{"N04", "N05", "N06"} {
		discovery, discoveryErr := gateway.base.peers.Discover(context.Background(), nucleus)
		if discoveryErr != nil {
			t.Fatalf("discovery %s failed: %v", nucleus, discoveryErr)
		}
		if !supportsExecutableCapability(discovery, "e2e."+strings.ToLower(nucleus)) {
			t.Fatalf("discovery %s omitted executable capability: %#v", nucleus, discovery)
		}
	}

	payload := map[string]any{"tasks": []any{
		map[string]any{"id": "n04", "capability": "e2e.n04", "payload": map[string]any{"values": []any{4, 4}}, "required": true},
		map[string]any{"id": "n05", "capability": "e2e.n05", "payload": map[string]any{"values": []any{5, 5}}, "required": true},
		map[string]any{"id": "n06", "capability": "e2e.n06", "payload": map[string]any{"values": []any{6, 6}}, "required": true},
	}}
	correlation := "e2e-n01-to-n07"
	request := protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: protocol.SoulMeshContractVersion, MessageID: protocol.NewTraceID(), Source: "N01", Target: protocol.N07, Timestamp: time.Now().UnixMilli(), Nonce: protocol.NewTraceID(), CorrelationID: correlation, Type: "request", Payload: map[string]any{"capability": "supergpu.parallel", "payload": payload}}
	if err := protocol.SignHMAC(&request, federationE2ESecret); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", strings.NewReader(string(body)))
	req.Header.Set("x-soul-mesh-nonce", request.Nonce)
	req.Header.Set("x-soul-mesh-hmac", request.HMAC)
	rec := httptest.NewRecorder()
	gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("N01->N07 federation failed: code=%d body=%s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["correlationId"] != correlation || out["kind"] != "response" {
		t.Fatalf("unexpected final envelope: %#v", out)
	}
	result, ok := out["payload"].(map[string]any)
	if !ok || result["requiredFailure"] != false {
		t.Fatalf("federation reported required failure: %#v", result)
	}
	tasks, ok := result["tasks"].([]any)
	if !ok || len(tasks) != 3 {
		t.Fatalf("expected three federated tasks: %#v", result["tasks"])
	}
	for _, taskID := range []string{"n04", "n05", "n06"} {
		found := false
		for _, item := range tasks {
			row := item.(map[string]any)
			if row["id"] == taskID && row["status"] == "ok" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("task %s did not complete successfully: %#v", taskID, tasks)
		}
	}
}
