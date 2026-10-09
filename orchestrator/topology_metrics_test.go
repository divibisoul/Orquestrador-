package orchestrator

import "testing"

func TestSOULTopologyChannelMetricsSeparateLogicalRoutesFromChainSurfaces(t *testing.T) {
	topology := SOULTopology()
	metrics, ok := topology["channel_metrics"].(map[string]any)
	if !ok {
		t.Fatal("channel_metrics missing")
	}
	want := map[string]int{
		"active_ai_nuclei": 6,
		"active_ai_logical_directed_routes": 30,
		"active_ai_endpoint_surfaces": 60,
		"structural_n07_additional_logical_directed_routes": 12,
		"structural_n07_additional_endpoint_surfaces": 24,
		"total_identity_nuclei": 7,
		"total_logical_directed_routes": 42,
		"total_endpoint_surfaces": 84,
		"adjacent_fusion_edges": 6,
		"adjacent_fusion_directional_channels": 12,
	}
	for key, expected := range want {
		got, ok := metrics[key].(int)
		if !ok || got != expected {
			t.Errorf("%s=%v, want %d", key, metrics[key], expected)
		}
	}
	if metrics["topology_state"] != "LOGICAL_UNVERIFIED" {
		t.Errorf("topology_state=%v, want LOGICAL_UNVERIFIED", metrics["topology_state"])
	}
	if got, ok := topology["directional"].(int); !ok || got != 12 {
		t.Errorf("legacy directional=%v, want adjacent-chain count 12", topology["directional"])
	}
}
