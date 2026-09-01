package mesh

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
			response := map[string]any{"protocol": "soul-mesh/1", "contractVersion": protocol.SoulMeshContractVersion, "id": responseID, "correlationId": in.CorrelationID, "source": nucleus, "target": protocol.N07, "kind": "response", "capability": capability, "payload": payload, "timestamp": time.Now().UnixMilli(), "nonce": responseNonce}
			if capability == "mesh.discovery" || capability == "mesh.describe" {
				response["executableCapabilities"] = payload["executableCapabilities"]
			}
			env := protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: protocol.SoulMeshContractVersion, MessageID: responseID, Source: nucleus, Target: protocol.N07, Timestamp: response["timestamp"].(int64), Nonce: responseNonce, CorrelationID: in.CorrelationID, Type: "TASK_RESULT", Payload: map[string]any{"capability": capability, "payload": payload}}
			if _, ok := payload["error"]; ok {
				env.Type = "ERROR"
			}
			if err := protocol.SignHMAC(&env, federationE2ESecret); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response["hmac"] = env.HMAC
			w.Header().Set("content-type", "application/json")
			w.Header().Set("x-soul-mesh-nonce", env.Nonce)
			w.Header().Set("x-soul-mesh-hmac", env.HMAC)
			_ = json.NewEncoder(w).Encode(response)
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
	wire := canonicalRequest("request", "supergpu.parallel", correlation, nil)
	wire["payload"] = payload
	raw := wire["payload"].(map[string]any)
	raw["capability"] = "supergpu.parallel"
	canonical, err := canonicalN01Bytes(canonicalWireEnvelope{Protocol: "soul-mesh/1", ContractVersion: protocol.SoulMeshContractVersion, ID: wire["id"].(string), CorrelationID: correlation, Source: "N01", Target: "N07", Kind: "request", Capability: "supergpu.parallel", Payload: raw, Timestamp: wire["timestamp"].(int64), Nonce: wire["nonce"].(string)}, wire["nonce"].(string))
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte(federationE2ESecret))
	_, _ = mac.Write(canonical)
	hmacValue := hex.EncodeToString(mac.Sum(nil))
	wire["hmac"] = hmacValue
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	req.Header.Set("x-soul-mesh-nonce", wire["nonce"].(string))
	req.Header.Set("x-soul-mesh-hmac", hmacValue)
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
