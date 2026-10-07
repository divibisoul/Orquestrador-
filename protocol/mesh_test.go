package protocol

import (
	"sync"
	"testing"
	"time"
)

func validEnvelope() MeshEnvelope {
	return MeshEnvelope{Version: SoulMeshVersion, ContractVersion: SoulMeshContractVersion, MessageID: NewTraceID(), Operation: "execute", Payload: map[string]any{"x": 1}, CorrelationID: "corr-1", Source: N07, Target: N01, Timestamp: time.Now().UnixMilli(), Nonce: NewTraceID(), Type: "CAPABILITY_REQUEST", Metadata: map[string]string{"trace": "corr-1"}}
}

func TestCanonicalMeshContractVersion(t *testing.T) {
	if SoulMeshContractVersion != "1.1.0" {
		t.Fatalf("canonical Mesh contract drifted: got %s", SoulMeshContractVersion)
	}
	m := validEnvelope()
	m.ContractVersion = "1.2"
	if err := m.Validate(); err == nil {
		t.Fatal("legacy/incompatible contract version was accepted")
	}
}
func TestMeshRejectsUnknownNucleus(t *testing.T) {
	m := validEnvelope()
	m.Source = "N99"
	if err := m.Validate(); err == nil {
		t.Fatal("unknown source nucleus was accepted")
	}
	m = validEnvelope()
	m.Target = "N99"
	if err := m.Validate(); err == nil {
		t.Fatal("unknown target nucleus was accepted")
	}
}
func TestMeshRoundTrip(t *testing.T) {
	m := validEnvelope()
	b, err := EncodeMesh(m)
	if err != nil {
		t.Fatal(err)
	}
	d, err := DecodeMesh(b)
	if err != nil {
		t.Fatal(err)
	}
	if d.CorrelationID != m.CorrelationID || d.Operation != m.Operation || d.ContractVersion != SoulMeshContractVersion || d.Metadata["trace"] != "corr-1" {
		t.Fatal("mesh round-trip failed")
	}
}
func TestHMACAuthenticatesOperationAndMetadata(t *testing.T) {
	m := validEnvelope()
	secret := "0123456789abcdef"
	if err := SignHMAC(&m, secret); err != nil {
		t.Fatal(err)
	}
	if err := VerifyHMAC(m, secret, time.UnixMilli(m.Timestamp)); err != nil {
		t.Fatal(err)
	}
	m.Operation = "tampered"
	if err := VerifyHMAC(m, secret, time.UnixMilli(m.Timestamp)); err == nil {
		t.Fatal("operation tampering was not detected")
	}
	m = validEnvelope()
	if err := SignHMAC(&m, secret); err != nil {
		t.Fatal(err)
	}
	m.Metadata["trace"] = "tampered"
	if err := VerifyHMAC(m, secret, time.UnixMilli(m.Timestamp)); err == nil {
		t.Fatal("metadata tampering was not detected")
	}
}
func TestMeshRejectsInvalidEnvelope(t *testing.T) {
	cases := []MeshEnvelope{{}, {ContractVersion: SoulMeshContractVersion}, {Version: SoulMeshVersion, ContractVersion: SoulMeshContractVersion, Operation: "execute", CorrelationID: "c", Source: N07}, {Version: SoulMeshVersion, ContractVersion: SoulMeshContractVersion, MessageID: "m", CorrelationID: "c", Source: N07, Target: N01, Timestamp: time.Now().UnixMilli(), Nonce: "n"}}
	for i, m := range cases {
		if _, err := EncodeMesh(m); err == nil {
			t.Fatalf("case %d accepted", i)
		}
	}
}
func TestVerifyHMACRejectsReplayConcurrently(t *testing.T) {
	m := validEnvelope()
	secret := "0123456789abcdef"
	if err := SignHMAC(&m, secret); err != nil {
		t.Fatal(err)
	}
	const workers = 8
	var wg sync.WaitGroup
	results := make(chan error, workers)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() { defer wg.Done(); results <- VerifyHMAC(m, secret, time.UnixMilli(m.Timestamp)) }()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("expected exactly one successful verification, got %d", success)
	}
}
