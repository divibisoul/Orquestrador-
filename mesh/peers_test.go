package mesh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestPeerClientLoadsAllSixMeshPeers(t *testing.T) {
	for _, n := range []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06} {
		t.Setenv("SOUL_MESH_"+n+"_URL", "http://"+n)
	}
	t.Setenv("SOUL_MESH_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	p, err := NewPeerClient(nil)
	if err != nil { t.Fatal(err) }
	if got := len(p.ConfiguredPeers()); got != 6 { t.Fatalf("expected six configured peers, got %d", got) }
}

func TestPeerClientCircuitOpensAfterThreeFailures(t *testing.T) {
	for _, n := range []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06} {
		t.Setenv("SOUL_MESH_"+n+"_URL", "http://127.0.0.1:1")
	}
	t.Setenv("SOUL_MESH_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	p, err := NewPeerClient(nil)
	if err != nil { t.Fatal(err) }
	p.maxRetry = 1; p.cooldown = time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), time.Second); defer cancel()
	for i := 0; i < 3; i++ { _, _ = p.Call(ctx, protocol.N01, "mesh.ping", map[string]any{}) }
	var state PeerInfo
	for _, peer := range p.ConfiguredPeers() { if peer.Nucleus == protocol.N01 { state = peer; break } }
	if state.Circuit != CircuitOpen || state.Failures < 3 { t.Fatalf("expected open circuit after three failures, got %+v", state) }
	if _, err = p.Call(ctx, protocol.N01, "mesh.ping", map[string]any{}); err == nil { t.Fatal("expected open circuit to reject request") }
}

func TestPeerClientDoesNotUseLegacySelfRoute(t *testing.T) {
	_ = os.Setenv("SOUL_MESH_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	p, err := NewPeerClient(nil)
	if err != nil { t.Fatal(err) }
	if _, err := p.Call(context.Background(), protocol.N07, "mesh.ping", nil); err == nil { t.Fatal("expected N07 self-call to be rejected") }
}

func TestPeerClientDiscoveryCacheAvoidsDuplicateCalls(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	t.Setenv("SOUL_MESH_HMAC_SECRET", secret)
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request protocol.MeshEnvelope
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil { t.Fatal(err) }
		atomic.AddInt32(&calls, 1)
		response := protocol.MeshEnvelope{Version: protocol.SoulMeshVersion, ContractVersion: protocol.SoulMeshContractVersion, MessageID: protocol.NewTraceID(), Source: protocol.N01, Target: protocol.N07, Timestamp: time.Now().UnixMilli(), Nonce: protocol.NewTraceID(), CorrelationID: request.CorrelationID, Type: "TASK_RESULT", Payload: map[string]any{"operations": []string{"document.validate"}}}
		if err := protocol.SignHMAC(&response, secret); err != nil { t.Fatal(err) }
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	t.Setenv("SOUL_MESH_N01_URL", server.URL)
	p, err := NewPeerClient(nil)
	if err != nil { t.Fatal(err) }
	p.discoveryTTL = time.Second
	ctx := context.Background()
	if _, err := p.Discover(ctx, protocol.N01); err != nil { t.Fatal(err) }
	if _, err := p.Discover(ctx, protocol.N01); err != nil { t.Fatal(err) }
	if got := atomic.LoadInt32(&calls); got != 1 { t.Fatalf("expected cached discovery, got %d network calls", got) }
	p.InvalidateDiscovery(protocol.N01)
	if _, err := p.Discover(ctx, protocol.N01); err != nil { t.Fatal(err) }
	if got := atomic.LoadInt32(&calls); got != 2 { t.Fatalf("expected invalidation to force discovery, got %d calls", got) }
}
