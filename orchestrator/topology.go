package orchestrator

import (
	"sort"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type PeerChannel struct {
	Peer       string   `json:"peer"`
	Direction  string   `json:"direction"`
	Operations []string `json:"operations"`
	Transports []string `json:"transports"`
}

func SOULTopology() map[string]any {
	peers := []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06}
	operations := []string{
		"mesh.discovery",
		"mesh.capability.resolve",
		"mesh.ping",
		"mesh.health",
		"mesh.delegate",
		"mesh.fusion.describe",
		"mesh.fusion.execute",
		"mesh.supergpu.execute",
		"mesh.supergpu.parallel",
	}
	transports := []string{"IN_PROCESS", "LOOPBACK_HTTP", "HTTP", "REALTIME", "EVENT"}
	channels := make([]PeerChannel, 0, len(peers)*2)
	for _, peer := range peers {
		channels = append(channels,
			PeerChannel{Peer: peer, Direction: "IN", Operations: append([]string(nil), operations...), Transports: append([]string(nil), transports...)},
			PeerChannel{Peer: peer, Direction: "OUT", Operations: append([]string(nil), operations...), Transports: append([]string(nil), transports...)},
		)
	}
	sort.Slice(channels, func(i, j int) bool {
		if channels[i].Peer == channels[j].Peer {
			return channels[i].Direction < channels[j].Direction
		}
		return channels[i].Peer < channels[j].Peer
	})
	return map[string]any{
		"nucleus":         protocol.N07,
		"base_nuclei":     peers,
		"in_peer_count":   len(peers),
		"out_peer_count":  len(peers),
		"channels":        channels,
		"bidirectional":   len(peers) * 2,
		"logical_pairs":   len(peers),
		"logical_links":   len(peers),
		"transports":      transports,
		"mesh":             "canonical-soul-mesh",
		"capability_layer": "discovery→routing→delegation→execution→response→correlation→composition",
	}
}
