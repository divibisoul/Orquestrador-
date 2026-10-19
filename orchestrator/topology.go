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
	activeAI := []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06}
	activeAIDirectedRoutes := len(activeAI) * (len(activeAI) - 1)
	totalLogicalDirectedRoutes := len(nuclei) * (len(nuclei) - 1)
	structuralN07AdditionalRoutes := totalLogicalDirectedRoutes - activeAIDirectedRoutes
	ops := []string{"mesh.ping", "mesh.health", "mesh.discovery", "mesh.capabilities", "mesh.capability.resolve@1.0.0", "mesh.delegate", "mesh.fusion.describe", "mesh.fusion.execute", "mesh.supergpu.describe", "mesh.supergpu.execute", "mesh.supergpu.parallel", "supergpu.federated.execute", "prefrontal.admission", "prefrontal.orbital.evaluate@1.0.0", "transcendental.estimate@1.0.0", "neural.forward", "neural.learn", "jev.systemone@1.0.0", "cooperation.handshake@1.0.0", "cooperation.exchange@1.0.0", "cooperation.health@1.0.0", "cognitive.planning.sources@1.0.0", "superpowers.agent.describe@1.0.0", "superpowers.agent.route@1.0.0", "soul.integration.engineers.describe@1.0.0", "soul.integration.engineer.prepare@1.0.0", "soul.capability.upgrade.describe@1.0.0", "soul.capability.upgrade.resolve@1.0.0", "soul.capability.augment.describe@1.0.0", "soul.capability.augment.resolve@1.0.0", "grf.describe@2.0.0", "grf.participant.describe@2.0.0", "grce.describe@2.0.0", "grce.bindings.describe@2.0.0", "grce.cycle.execute@2.0.0", "grf.participant.ingest@2.0.0", "grce.sara.authoritative.execute@2.0.0"}
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
		"nucleus":        protocol.N07,
		"nuclei":         nuclei,
		"adjacency":      [][]string{{protocol.N01, protocol.N02}, {protocol.N02, protocol.N03}, {protocol.N03, protocol.N04}, {protocol.N04, protocol.N05}, {protocol.N05, protocol.N06}, {protocol.N06, protocol.N07}},
		"fusion_policy":  "adjacent-only-dynamic",
		"in_peer_count":  6,
		"out_peer_count": 6,
		"channels":       channels,
		// "directional" is retained as the legacy count of adjacent-chain
		// channel entries. It is not the count of all logical Mesh routes.
		"directional":    len(channels),
		"adjacent_directional_channels": len(channels),
		"channel_metrics": map[string]any{
			"active_ai_nuclei": len(activeAI),
			"active_ai_logical_directed_routes": activeAIDirectedRoutes,
			"active_ai_endpoint_surfaces": activeAIDirectedRoutes * 2,
			"structural_n07_additional_logical_directed_routes": structuralN07AdditionalRoutes,
			"structural_n07_additional_endpoint_surfaces": structuralN07AdditionalRoutes * 2,
			"total_identity_nuclei": len(nuclei),
			"total_logical_directed_routes": totalLogicalDirectedRoutes,
			"total_endpoint_surfaces": totalLogicalDirectedRoutes * 2,
			"adjacent_fusion_edges": len(nuclei) - 1,
			"adjacent_fusion_directional_channels": len(channels),
			"topology_state": "LOGICAL_UNVERIFIED",
		},
		"transports":     transports,
		"mesh":           "canonical-soul-mesh",
		"execution":      "orbital-resource-simulation→prefrontal-admission→discovery→routing→delegation→decision(Jev)→hardware-lease/fusion→execution→response→correlation",
		"cooperation": map[string]any{
			"transport":            "canonical-soul-mesh",
			"negotiation":          "discovery→handshake→exchange",
			"correlation_required": true,
			"outcome_observation":  "learning-machine",
			"second_bus":           false,
		},
		"decision_layer": map[string]any{"provider": "TypeSafe/Jev", "operation": "jev.systemone@1.0.0", "transport": "server-side HTTP", "enabled_when": "JEV_API_KEY configured"},
	}
}
