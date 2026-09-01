package orchestrator

import "testing"

func TestSOULTopologyHasSixPeersAndTwelveDirectionalChannels(t *testing.T) {
	topology := SOULTopology()
	if topology["nucleus"] != "N07" {
		t.Fatalf("unexpected nucleus: %#v", topology["nucleus"])
	}
	if topology["in_peer_count"] != 6 || topology["out_peer_count"] != 6 {
		t.Fatalf("expected 6 IN and 6 OUT peers: %#v", topology)
	}
	if topology["directional"] != 12 {
		t.Fatalf("expected 12 directional channels: %#v", topology["directional"])
	}
}
