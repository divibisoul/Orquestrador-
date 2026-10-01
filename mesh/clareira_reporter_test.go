package mesh

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/supergpu"
)

type clareiraReporterPeerProbe struct {
	nucleus    string
	capability string
	payload    map[string]any
	correlation string
}

func (p *clareiraReporterPeerProbe) CallWithCorrelation(_ context.Context, nucleus, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	p.nucleus, p.capability, p.payload, p.correlation = nucleus, capability, payload, correlation
	return map[string]any{"accepted": true}, nil
}

func TestClareiraReporterBuildsCorrelatedStatePacket(t *testing.T) {
	peer := &clareiraReporterPeerProbe{}
	reporter, err := NewClareiraReporter(peer)
	if err != nil {
		t.Fatal(err)
	}

	err = reporter.Report(context.Background(), supergpu.ExecutionEvent{
		Phase:         "completed",
		Operation:     "square",
		DeviceID:      "cpu-0",
		Backend:       "cpu",
		InputSize:     2,
		OutputSize:    2,
		CorrelationID: "corr-clareira-2",
	})
	if err != nil {
		t.Fatal(err)
	}

	if peer.nucleus != "N01" || peer.capability != ClareiraCapability {
		t.Fatalf("unexpected route: %s/%s", peer.nucleus, peer.capability)
	}
	if peer.correlation != "corr-clareira-2" {
		t.Fatalf("correlation lost: %s", peer.correlation)
	}
	packet, ok := peer.payload["packet"].(map[string]any)
	if !ok {
		t.Fatalf("packet missing: %#v", peer.payload)
	}
	if packet["sourceId"] != "N07.SuperGPU" || packet["packetType"] != "StateReport" {
		t.Fatalf("unexpected packet identity: %#v", packet)
	}
	if packet["correlationId"] != "corr-clareira-2" {
		t.Fatalf("packet correlation lost: %#v", packet)
	}
}
