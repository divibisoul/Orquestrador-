package mesh

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestCircuitEligibleFailureRejectsAuthenticationAndContractErrors(t *testing.T) {
	for _, message := range []string{
		"invalid Mesh HMAC",
		"invalid HMAC-SHA256 value",
		"peer response HMAC credentials are missing",
		"replay detected",
		protocol.InvalidMeshContract + ": field id is required",
		"peer correlation mismatch",
		"mesh response protocol mismatch",
	} {
		if circuitEligibleFailure(message) {
			t.Fatalf("authentication/contract failure must not open circuit: %q", message)
		}
	}
}

func TestCircuitEligibleFailureAcceptsConnectivityFailures(t *testing.T) {
	for _, message := range []string{
		"dial tcp: connection refused",
		"context deadline exceeded",
		"unexpected EOF",
		"peer request failed: 503 Service Unavailable",
	} {
		if !circuitEligibleFailure(message) {
			t.Fatalf("connectivity failure should remain circuit eligible: %q", message)
		}
	}
}
