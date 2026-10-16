package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestPeerClientLoadsAllSixMeshPeers(t *testing.T) {
	for _, n := range []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06} {
		t.Setenv("SOUL_MESH_"+n+"_URL", "http://"+n)
	}
	t.Setenv("SOUL_MESH_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	p, err := NewPeerClient(&http.Client{Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	peers := p.ConfiguredPeers()
	if len(peers) != 6 {
		t.Fatalf("expected six configured peers, got %d", len(peers))
	}
}

func TestPeerClientCircuitOpensAfterThreeFailures(t *testing.T) {
	for _, n := range []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06} {
		t.Setenv("SOUL_MESH_"+n+"_URL", "http://127.0.0.1:1")
	}
	t.Setenv("SOUL_MESH_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	p, err := NewPeerClient(&http.Client{Timeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	p.maxRetry = 1
	p.cooldown = time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for i := 0; i < 3; i++ {
		_, _ = p.Call(ctx, protocol.N01, "mesh.ping", map[string]any{})
	}
	var state PeerInfo
	for _, peer := range p.ConfiguredPeers() {
		if peer.Nucleus == protocol.N01 {
			state = peer
			break
		}
	}
	if state.Circuit != CircuitOpen || state.Failures < 3 {
		t.Fatalf("expected open circuit after three failures, got %+v", state)
	}
	_, err = p.Call(ctx, protocol.N01, "mesh.ping", map[string]any{})
	if err == nil {
		t.Fatal("expected open circuit to reject request")
	}
}

func TestPeerClientDoesNotUseLegacySelfRoute(t *testing.T) {
	_ = os.Setenv("SOUL_MESH_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	p, err := NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Call(context.Background(), protocol.N07, "mesh.ping", nil); err == nil {
		t.Fatal("expected N07 self-call to be rejected")
	}
}

func TestDiscoveryCacheReturnsCopyAndInvalidates(t *testing.T) {
	p, err := NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	p.discoveryCacheTTL = time.Minute
	original := map[string]any{"executableCapabilities": []any{"neural.forward@1.0.0"}}
	p.storeDiscovery(protocol.N01, original)
	original["mutated"] = true

	cached, ok := p.discoveryFromCache(protocol.N01)
	if !ok {
		t.Fatal("expected fresh discovery cache entry")
	}
	if _, exists := cached["mutated"]; exists {
		t.Fatal("cache returned aliased map data")
	}

	p.invalidateDiscovery(protocol.N01)
	if _, ok := p.discoveryFromCache(protocol.N01); ok {
		t.Fatal("expected discovery cache entry to be invalidated")
	}
}

func TestDiscoveryCacheExpires(t *testing.T) {
	p, err := NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	p.discoveryCacheTTL = time.Nanosecond
	p.storeDiscovery(protocol.N01, map[string]any{"status": "healthy"})
	time.Sleep(2 * time.Millisecond)
	if _, ok := p.discoveryFromCache(protocol.N01); ok {
		t.Fatal("expected expired discovery cache entry")
	}
}

type fixedRouteScorer map[string]float64

func (s fixedRouteScorer) Weight(_, target, _ string) float64 { return s[target] }

type recordingRouteObserver struct {
	outcomes []string
}

func (o *recordingRouteObserver) ObserveRoute(_ context.Context, _, target, capability, correlation string, success bool) error {
	o.outcomes = append(o.outcomes, target+":"+capability+":"+correlation+":"+fmt.Sprint(success))
	return nil
}

func TestPeerClientOrdersPeersByLearnedRouteWeight(t *testing.T) {
	p := &PeerClient{
		peers: map[string]PeerInfo{
			protocol.N01: {Nucleus: protocol.N01, URL: "http://n01", Healthy: true, Latency: 50 * time.Millisecond},
			protocol.N02: {Nucleus: protocol.N02, URL: "http://n02", Healthy: false, Latency: 5 * time.Millisecond},
			protocol.N03: {Nucleus: protocol.N03, URL: "http://n03", Healthy: true, Latency: 10 * time.Millisecond},
		},
	}
	p.SetRouteScorer(fixedRouteScorer{protocol.N01: 0.2, protocol.N02: 0.9, protocol.N03: 0.9})
	ordered := p.orderedPeers("ai.generate")
	if ordered[0].Nucleus != protocol.N03 || ordered[1].Nucleus != protocol.N02 || ordered[2].Nucleus != protocol.N01 {
		t.Fatalf("unexpected learned route order: %+v", ordered)
	}
}

func TestPeerClientNeutralizesInvalidLearnedWeight(t *testing.T) {
	p := &PeerClient{peers: map[string]PeerInfo{
		protocol.N01: {Nucleus: protocol.N01, URL: "http://n01"},
		protocol.N02: {Nucleus: protocol.N02, URL: "http://n02"},
	}}
	p.SetRouteScorer(fixedRouteScorer{protocol.N01: 2, protocol.N02: 0.9})
	ordered := p.orderedPeers("neural.forward")
	if ordered[0].Nucleus != protocol.N02 {
		t.Fatalf("invalid learned weight not neutralized: %+v", ordered)
	}
}

func TestPeerClientObserveRouteDoesNotRewriteOutcome(t *testing.T) {
	p := &PeerClient{peers: map[string]PeerInfo{protocol.N01: {Nucleus: protocol.N01, URL: "http://n01"}}}
	o := &recordingRouteObserver{}
	p.SetRouteOutcomeObserver(o)
	p.observeRoute(protocol.N01, "core.health", "corr-test", true)
	if len(o.outcomes) != 1 || o.outcomes[0] != "N01:core.health:corr-test:true" {
		t.Fatalf("unexpected observed outcome: %#v", o.outcomes)
	}
}


func TestPeerClientRetriesReuseLogicalMessageID(t *testing.T) {
	secret := "n07-e2e-secret-0123456789abcdef"
	var requestIDs []string
	attempt := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var wire map[string]any
		if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
			t.Errorf("decode retry request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, _ := wire["id"].(string)
		requestIDs = append(requestIDs, id)
		attempt++

		if attempt == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "transient"})
			return
		}

		correlation, _ := wire["correlationId"].(string)
		timestamp := time.Now().UnixMilli()
		env := protocol.MeshEnvelope{
			Version:         protocol.SoulMeshVersion,
			ContractVersion: protocol.SoulMeshContractVersion,
			MessageID:       "retry-response",
			Source:          protocol.N01,
			Target:          protocol.N07,
			Timestamp:       timestamp,
			Nonce:           "retry-response-nonce-unique",
			CorrelationID:   correlation,
			Type:            "TASK_RESULT",
			Payload: map[string]any{
				"capability": "mesh.ping",
				"payload":    map[string]any{"ok": true},
			},
		}
		if err := protocol.SignHMAC(&env, secret); err != nil {
			t.Fatalf("sign response: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"protocol":        "soul-mesh/1",
			"contractVersion": env.ContractVersion,
			"id":              env.MessageID,
			"correlationId":   env.CorrelationID,
			"source":          env.Source,
			"target":          env.Target,
			"kind":            "response",
			"capability":      "mesh.ping",
			"payload":         map[string]any{"ok": true},
			"timestamp":       env.Timestamp,
			"nonce":           env.Nonce,
			"hmac":            env.HMAC,
		})
	}))
	defer server.Close()

	t.Setenv("SOUL_MESH_N01_URL", server.URL)
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	p, err := NewPeerClient(server.Client())
	if err != nil {
		t.Fatal(err)
	}
	p.maxRetry = 2

	correlation := "mesh-retry-id-stability"
	if _, err := p.CallWithCorrelation(context.Background(), protocol.N01, "mesh.ping", map[string]any{"probe": true}, correlation); err != nil {
		t.Fatalf("expected second retry to succeed: %v", err)
	}
	if len(requestIDs) != 2 {
		t.Fatalf("expected exactly two wire attempts, got %d", len(requestIDs))
	}
	if requestIDs[0] == "" || requestIDs[1] == "" {
		t.Fatalf("expected message IDs on both attempts: %#v", requestIDs)
	}
	if requestIDs[0] != requestIDs[1] {
		t.Fatalf("logical message ID changed across retry: %#v", requestIDs)
	}
}


func TestPeerClientRegisterPeerUpdatesCanonicalTable(t *testing.T) {
	p, err := NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.RegisterPeer(protocol.N01, "http://127.0.0.1:18081", []string{"mesh.health"}); err != nil {
		t.Fatalf("register peer: %v", err)
	}
	peers := p.ConfiguredPeers()
	var got PeerInfo
	for _, peer := range peers {
		if peer.Nucleus == protocol.N01 {
			got = peer
			break
		}
	}
	if got.URL != "http://127.0.0.1:18081" || !got.Healthy || got.Circuit != CircuitClosed {
		t.Fatalf("unexpected registered peer: %+v", got)
	}
}
