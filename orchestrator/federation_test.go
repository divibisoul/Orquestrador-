package orchestrator

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"
)

func TestRouteScorePrefersHealthyFastPeer(t *testing.T) {
	fast := routeScore(true, 0.99, 25*time.Millisecond, 1)
	slow := routeScore(true, 0.75, 900*time.Millisecond, 4)
	unhealthy := routeScore(false, 1, time.Millisecond, 0)
	if fast <= slow {
		t.Fatalf("expected fast peer score > slow peer: fast=%v slow=%v", fast, slow)
	}
	if unhealthy != 0 {
		t.Fatalf("expected unhealthy peer score 0, got %v", unhealthy)
	}
}

func TestFederationRegisterAndDiscover(t *testing.T) {
	f := NewFederation()
	if err := f.RegisterPeer("N01", "http://n01.example", []string{"analysis", "compute.execute@1.0.0", "analysis"}); err != nil {
		t.Fatal(err)
	}
	candidates := f.Discover("compute.execute")
	if len(candidates) != 1 || candidates[0].Nucleus != "N01" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
	if got := len(f.Snapshot()[0].Capabilities); got != 2 {
		t.Fatalf("expected unique capabilities, got %d", got)
	}
}

func TestFederationRejectsN07SelfPeer(t *testing.T) {
	f := NewFederation()
	if err := f.RegisterPeer("N07", "http://n07.example", nil); err == nil {
		t.Fatal("expected self-peer rejection")
	}
}

func TestFederationRejectsInvalidNucleus(t *testing.T) {
	f := NewFederation()
	if err := f.RegisterPeer("N08", "http://n08.example", nil); err == nil {
		t.Fatal("expected invalid nucleus rejection")
	}
}

func TestCallWithRetryHonorsContext(t *testing.T) {
	f := NewFederation()
	f.maxRetries = 2
	f.baseBackoff = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	attempts := 0
	start := time.Now()
	_, err := f.callWithRetry(ctx, "N01", func() (map[string]any, error) {
		attempts++
		return nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected retry failure")
	}
	if attempts != 1 {
		t.Fatalf("expected context to prevent second attempt, got %d attempts", attempts)
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("retry loop exceeded reasonable context bound")
	}
}

func TestFederationURLIsNormalizedByAdapter(t *testing.T) {
	f := NewFederation()
	if err := f.RegisterPeer("N02", "http://n02.example///", []string{"search"}); err != nil {
		t.Fatal(err)
	}
	peer := f.Snapshot()[0]
	u, err := url.Parse(peer.URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "" {
		t.Fatalf("expected normalized empty path, got %q", u.Path)
	}
}

func TestFederationDiscoverRequiresPeers(t *testing.T) {
	f := NewFederation()
	if _, err := f.DiscoverLive(context.Background()); err == nil {
		t.Fatal("expected no-peer error")
	}
}
