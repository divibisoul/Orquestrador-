package mesh

import "testing"

func TestNervoVagoCapabilityUsesCanonicalMeshBoundary(t *testing.T) {
  capability := "nervo.vago.publish@1.0.0"
  if !structuredMeshCapability(capability) {
    t.Fatalf("%s is not recognized as a structured canonical Mesh capability", capability)
  }

  metadata := map[string]string{}
  payload := map[string]any{
    "vagus_version": "1.0",
    "correlation_id": "nervo-vago-contract-test",
    "source": "N02",
    "target": "NervoVago",
    "type": "test.signal",
    "payload": map[string]any{"value": "preserve"},
  }
  if err := copyStructuredCapabilityMetadata(metadata, capability, payload); err != nil {
    t.Fatal(err)
  }
  if metadata["sara_vagus_json"] == "" {
    t.Fatal("canonical NervoVago payload was not preserved for SARA")
  }
}
