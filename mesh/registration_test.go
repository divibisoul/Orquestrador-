package mesh

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestRegistrationRegistryPreservesPeerIdentity(t *testing.T) {
	r := NewRegistrationRegistry()
	e := &orchestrator.Engine{}
	if err := RegisterRegistrationOperation(e, r); err != nil {
		t.Fatal(err)
	}

	message := protocol.NewMessage("N01", "N07", "command", "mesh.register@1.0.0", nil)
	message.CorrelationID = "registration-test"
	message.Metadata = map[string]string{
		"endpoint":    "http://127.0.0.1:18081",
		"capabilities": "mesh.ping mesh.discovery grce.cycle.execute@1.0.0",
	}

	result, err := e.Submit(context.Background(), message)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("unexpected registration status: %s", result.Status)
	}

	entries := r.Snapshot()
	if len(entries) != 1 || entries[0].Nucleus != "N01" {
		t.Fatalf("unexpected registration snapshot: %+v", entries)
	}
}
