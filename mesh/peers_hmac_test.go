package mesh

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestPeerResponseHMACRoundTrip(t *testing.T) {
	secret := "n07-e2e-secret-0123456789abcdef"
	env := protocol.MeshEnvelope{
		Version: protocol.SoulMeshVersion,
		ContractVersion: protocol.SoulMeshContractVersion,
		MessageID: "n04-response-hmac-test",
		Source: "N04",
		Target: protocol.N07,
		Timestamp: time.Now().UnixMilli(),
		Nonce: "n04-response-nonce-test",
		CorrelationID: "peer-response-hmac-test",
		Type: "TASK_RESULT",
		Payload: map[string]any{
			"capability": "mesh.discovery",
			"payload": map[string]any{
				"executableCapabilities": []string{"e2e.n04"},
			},
		},
	}
	if err := protocol.SignHMAC(&env, secret); err != nil {
		t.Fatal(err)
	}
	wire := map[string]any{
		"protocol": "soul-mesh/1",
		"contractVersion": protocol.SoulMeshContractVersion,
		"id": env.MessageID,
		"correlationId": env.CorrelationID,
		"source": env.Source,
		"target": env.Target,
		"kind": "response",
		"capability": "mesh.discovery",
		"payload": map[string]any{"executableCapabilities": []string{"e2e.n04"}},
		"timestamp": env.Timestamp,
		"nonce": env.Nonce,
		"hmac": env.HMAC,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip map[string]any
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if err := verifyResponseHMAC(roundTrip, secret); err != nil {
		t.Fatalf("signed peer response did not survive JSON round-trip: %v; hmac=%v; len=%d", err, roundTrip["hmac"], len(roundTrip["hmac"].(string)))
	}
}
