package protocol

import (
	"testing"
	"time"
)

func TestValidateMeshWireResponseRejectsMissingRequiredField(t *testing.T) {
	response := map[string]any{
		"protocol":        "soul-mesh/1",
		"contractVersion": SoulMeshContractVersion,
		"id":              "response-1",
		"correlationId":   "corr-1",
		"source":          N04,
		"target":          N07,
		"kind":            "response",
		"capability":      "mesh.discovery",
		"payload":         map[string]any{"ok": true},
		"timestamp":       float64(time.Now().UnixMilli()),
		"nonce":           "nonce-1",
		"hmac":            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	delete(response, "id")
	if err := ValidateMeshWireResponse(response); err != nil {
		if contractErr, ok := err.(*MeshContractError); !ok || contractErr.Field != "id" {
			t.Fatalf("expected structured id error, got %T: %v", err, err)
		}
		return
	}
	t.Fatal("expected missing id to be rejected")
}

func TestValidateMeshWireResponseAcceptsCanonicalEnvelope(t *testing.T) {
	response := map[string]any{
		"protocol":        "soul-mesh/1",
		"contractVersion": SoulMeshContractVersion,
		"id":              "response-1",
		"correlationId":   "corr-1",
		"source":          N04,
		"target":          N07,
		"kind":            "response",
		"capability":      "mesh.discovery",
		"payload":         map[string]any{"ok": true},
		"timestamp":       float64(time.Now().UnixMilli()),
		"nonce":           "nonce-1",
		"hmac":            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	if err := ValidateMeshWireResponse(response); err != nil {
		t.Fatalf("canonical response rejected: %v", err)
	}
}
