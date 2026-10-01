package cooperation

import (
	"context"
	"errors"
	"testing"
)

type fakePeers struct {
	discovery map[string]map[string]any
	calls     int
}

func (f *fakePeers) Discover(_ context.Context, target string) (map[string]any, error) {
	value, ok := f.discovery[target]
	if !ok {
		return nil, errors.New("peer not found")
	}
	return value, nil
}

func (f *fakePeers) CallWithCorrelation(_ context.Context, _, _ string, _ map[string]any, _ string) (map[string]any, error) {
	f.calls++
	return map[string]any{"ok": true}, nil
}

type fakeObserver struct {
	successes int
	failures  int
}

func (f *fakeObserver) ObserveRoute(_ context.Context, _, _, _, _ string, success bool) error {
	if success {
		f.successes++
	} else {
		f.failures++
	}
	return nil
}

func TestHandshakeRequiresAdvertisedCapability(t *testing.T) {
	peers := &fakePeers{discovery: map[string]map[string]any{
		"N02": {"payload": map[string]any{"capabilities": []any{"gemini.text.generate", "neural.bnc_v2"}}},
	}}
	c, err := New(peers)
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Handshake(context.Background(), "N02", "gemini.text.generate", "corr-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "ready" || !containsCapability(got.Capabilities, "gemini.text.generate") {
		t.Fatalf("unexpected handshake: %#v", got)
	}
	if _, err := c.Handshake(context.Background(), "N02", "missing.capability", "corr-2"); err == nil {
		t.Fatal("missing advertised capability must fail closed")
	}
}

func TestExchangeObservesRealOutcome(t *testing.T) {
	peers := &fakePeers{discovery: map[string]map[string]any{
		"N01": {"capabilities": []any{"cooperation.exchange@1.0.0"}},
	}}
	obs := &fakeObserver{}
	c, err := New(peers, obs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Exchange(context.Background(), "N01", "cooperation.exchange@1.0.0", map[string]any{"x": 1}, "corr-ok"); err != nil {
		t.Fatal(err)
	}
	if peers.calls != 1 || obs.successes != 1 || obs.failures != 0 {
		t.Fatalf("unexpected outcome: calls=%d success=%d failure=%d", peers.calls, obs.successes, obs.failures)
	}
}
