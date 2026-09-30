package mesh

import (
	"encoding/json"
	"testing"
)

func TestStructuredCooperativeMeshPayloadMapping(t *testing.T) {
	t.Run("handshake", func(t *testing.T) {
		metadata := map[string]string{}
		err := copyStructuredCapabilityMetadata(metadata, "cooperation.handshake", map[string]any{
			"target":             "N02",
			"required_capability": "gemini.text.generate",
		})
		if err != nil {
			t.Fatal(err)
		}
		if metadata["target"] != "N02" || metadata["capability"] != "gemini.text.generate" {
			t.Fatalf("unexpected handshake metadata: %#v", metadata)
		}
	})

	t.Run("exchange", func(t *testing.T) {
		metadata := map[string]string{}
		payload := map[string]any{
			"target":     "N05",
			"capability": "inference.analyze",
			"payload":    map[string]any{"prompt": "cooperative continuity"},
		}
		if err := copyStructuredCapabilityMetadata(metadata, "cooperation.exchange", payload); err != nil {
			t.Fatal(err)
		}
		var nested map[string]any
		if err := json.Unmarshal([]byte(metadata["payload"]), &nested); err != nil {
			t.Fatal(err)
		}
		if metadata["target"] != "N05" || metadata["capability"] != "inference.analyze" || nested["prompt"] != "cooperative continuity" {
			t.Fatalf("unexpected exchange metadata: %#v nested=%#v", metadata, nested)
		}
	})

	t.Run("blueprint", func(t *testing.T) {
		metadata := map[string]string{}
		if err := copyStructuredCapabilityMetadata(metadata, "blueprint.resolve", map[string]any{
			"query": "audio multimodal perception",
			"limit": float64(3),
		}); err != nil {
			t.Fatal(err)
		}
		if metadata["query"] != "audio multimodal perception" || metadata["limit"] != "3" {
			t.Fatalf("unexpected blueprint metadata: %#v", metadata)
		}
	})

	t.Run("synergy", func(t *testing.T) {
		metadata := map[string]string{}
		payload := map[string]any{
			"synergy_sequence": map[string]any{"id": "sequence-1"},
			"capabilities": map[string]any{"N02": "gemini.text.generate"},
			"input": "bridge-evidence",
		}
		if err := copyStructuredCapabilityMetadata(metadata, "mesh.synergy.execute", payload); err != nil {
			t.Fatal(err)
		}
		if metadata["synergy_payload_json"] == "" || metadata["synergy_sequence_json"] == "" || metadata["synergy_capabilities_json"] == "" {
			t.Fatalf("structured synergy payload was not preserved: %#v", metadata)
		}
	})
}

func TestStructuredMeshCapabilityIsNarrowlyScoped(t *testing.T) {
	if !structuredMeshCapability("cooperation.handshake") ||
		!structuredMeshCapability("blueprint.resolve") ||
		!structuredMeshCapability("mesh.synergy.execute") {
		t.Fatal("expected registered structured capability families")
	}
	if structuredMeshCapability("neural.forward") || structuredMeshCapability("memory.search") {
		t.Fatal("numeric/memory routes must retain their existing payload contracts")
	}
}
