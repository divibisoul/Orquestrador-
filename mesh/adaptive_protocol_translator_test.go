package mesh

import "testing"

func TestAdaptiveProtocolTranslatorNormalizesLegacyContract(t *testing.T) {
	translator := AdaptiveProtocolTranslator{}
	got, err := translator.Translate(map[string]any{
		"protocol":        "soul-mesh/1",
		"contractVersion": "1.2.0",
		"id":              "legacy-id",
		"correlationId":   "corr-123",
		"source":          "N02",
		"target":          "N07",
		"payload":         map[string]any{"value": "x", "_upgraded": true, "ttl": float64(300)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ContractVersion != canonicalMeshContract || got.CorrelationID != "corr-123" || !got.NeedsResign {
		t.Fatalf("unexpected translation: %#v", got)
	}
	if _, ok := got.Payload["_upgraded"]; ok {
		t.Fatal("legacy upgrade marker leaked into canonical payload")
	}
	if _, ok := got.Payload["ttl"]; ok {
		t.Fatal("legacy ttl leaked into canonical payload")
	}
}

func TestAdaptiveProtocolTranslatorKeepsCanonicalContract(t *testing.T) {
	got, err := (AdaptiveProtocolTranslator{}).Translate(map[string]any{
		"protocol":        "soul-mesh/1",
		"contractVersion": "1.1.0",
		"correlationId":   "corr-456",
		"source":          "N01",
		"target":          "N07",
		"payload":         map[string]any{"state": "ready"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ContractVersion != canonicalMeshContract || got.CorrelationID != "corr-456" || got.NeedsResign {
		t.Fatalf("unexpected canonical translation: %#v", got)
	}
}
