package mesh

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestRegistrationRegistryPreservesPeerIdentity(t *testing.T) {
	r := NewRegistrationRegistry()

	message := protocol.NewMessage("N01", "N07", "command", "mesh.register@1.0.0", nil)
	message.CorrelationID = "registration-test"
	message.Metadata = map[string]string{
		"endpoint":     "http://127.0.0.1:18081",
		"capabilities": "mesh.ping mesh.discovery grce.cycle.execute@1.0.0",
	}

	registration, err := r.Register(message)
	if err != nil {
		t.Fatal(err)
	}
	if registration.Nucleus != "N01" || registration.Endpoint != "http://127.0.0.1:18081" {
		t.Fatalf("unexpected registration: %+v", registration)
	}

	entries := r.Snapshot()
	if len(entries) != 1 || entries[0].Nucleus != "N01" {
		t.Fatalf("unexpected registration snapshot: %+v", entries)
	}
}
