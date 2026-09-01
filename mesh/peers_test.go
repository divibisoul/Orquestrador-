package mesh

import (
	"context"
	"net/http"
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
