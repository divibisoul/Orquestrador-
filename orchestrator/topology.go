package orchestrator

import "github.com/divibisoul/Orquestrador-/protocol"

type PeerChannel struct {
	Peer       string   `json:"peer"`
	Direction  string   `json:"direction"`
	Operations []string `json:"operations"`
	Transports []string `json:"transports"`
}

// SOULTopology models the canonical seven-nucleus chain. Fusion is adjacent-only;
// non-adjacent work is delegated through normal Mesh routing.
func SOULTopology() map[string]any {
	nuclei := []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06, protocol.N07}
	ops := []string{"mesh.ping", "mesh.health", "mesh.discovery", "mesh.capabilities", "mesh.capability.resolve", "mesh.delegate", "mesh.fusion.describe", "mesh.fusion.execute", "mesh.supergpu.describe", "mesh.supergpu.execute", "mesh.supergpu.parallel", "supergpu.federated.execute", "prefrontal.admission", "neural.forward", "neural.learn"}
	transports := []string{"IN_PROCESS", "LOOPBACK_HTTP", "HTTP", "REALTIME", "EVENT"}
	channels := make([]PeerChannel, 0, 12)
	for i := 0; i < len(nuclei)-1; i++ {
		left, right := nuclei[i], nuclei[i+1]
		channels = append(channels,
			PeerChannel{Peer: right, Direction: "OUT", Operations: append([]string(nil), ops...), Transports: append([]string(nil), transports...)},
			PeerChannel{Peer: left, Direction: "IN", Operations: append([]string(nil), ops...), Transports: append([]string(nil), transports...)},
		)
	}
	return map[string]any{
		"nucleus": protocol.N07,
		"nuclei": nuclei,
		"adjacency": [][]string{{protocol.N01, protocol.N02}, {protocol.N02, protocol.N03}, {protocol.N03, protocol.N04}, {protocol.N04, protocol.N05}, {protocol.N05, protocol.N06}, {protocol.N06, protocol.N07}},
		"fusion_policy": "adjacent-only-dynamic",
		"in_peer_count": 6,
		"out_peer_count": 6,
		"channels": channels,
		"directional": len(channels),
		"transports": transports,
		"mesh": "canonical-soul-mesh",
		"execution": "prefrontal-admission→discovery→routing→delegation→hardware-lease/fusion→execution→response→correlation",
	}
}
