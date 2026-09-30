package orchestrator

import "testing"

func TestTopologyPreservesCanonicalAndAddsFavoriteSynergyGraph(t *testing.T) {
	topology := SOULTopology()
	if topology["fusion_policy"] != "adjacent-only-dynamic" {
		t.Fatalf("canonical fusion policy changed: %#v", topology["fusion_policy"])
	}
	graph, ok := topology["composition_graph"].(map[string]any)
	if !ok {
		t.Fatalf("composition graph missing: %#v", topology["composition_graph"])
	}
	if graph["favorite_sequence_id"] != FavoriteSynergySequenceID {
		t.Fatalf("favorite synergy sequence not exposed: %#v", graph["favorite_sequence_id"])
	}
	sequence, ok := graph["favorite_sequence"].(SynergySequence)
	if !ok {
		t.Fatalf("favorite sequence type mismatch: %#v", graph["favorite_sequence"])
	}
	if err := ValidateSynergySequence(sequence); err != nil {
		t.Fatal(err)
	}
}
