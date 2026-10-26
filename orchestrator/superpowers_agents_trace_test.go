package orchestrator

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestSuperpowersDelegationMetadataPreservesTraceAndCorrelation(t *testing.T) {
	m := protocol.NewMessage("N01", "N07", "command", SuperpowersAgentRouteOperation, nil)
	m.TraceID = "trace-parent-5cc47e"
	m.CorrelationID = "corr-parent-24c901"
	m.Metadata["trace_id"] = "stale-trace"
	m.Metadata["correlation_id"] = "stale-correlation"
	m.Metadata["agent_id"] = "superpowers.cortex-orbital-supergpu"

	got := superpowersDelegationMetadata(m)
	if got["trace_id"] != "" {
		t.Fatalf("nested call must generate its own trace id, got parent trace reused: %q", got["trace_id"])
	}
	if got["parent_trace_id"] != m.TraceID {
		t.Fatalf("parent trace link mismatch: got=%q want=%q", got["parent_trace_id"], m.TraceID)
	}
	if got["correlation_id"] != m.CorrelationID {
		t.Fatalf("delegated correlation id mismatch: got=%q want=%q", got["correlation_id"], m.CorrelationID)
	}
	if got["agent_id"] != m.Metadata["agent_id"] {
		t.Fatalf("agent metadata was not preserved: %#v", got)
	}
	if m.Metadata["trace_id"] != "stale-trace" || m.Metadata["correlation_id"] != "stale-correlation" {
		t.Fatalf("parent metadata was mutated: %#v", m.Metadata)
	}
	t.Logf("DELEGATION_TRACE_PRESERVED trace_id=%s correlation_id=%s", got["trace_id"], got["correlation_id"])
}
