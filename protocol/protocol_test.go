package protocol

import (
	"testing"
	"time"
)

func TestProtocol(t *testing.T) {
	m := NewMessage("N01", "N07", "command", "compute.execute", []float64{1, 2})
	b, err := Encode(m)
	if err != nil {
		t.Fatal(err)
	}
	d, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if d.TraceID != m.TraceID || d.CorrelationID != m.CorrelationID || d.Version != Version || d.ContractVersion != SoulMeshContractVersion {
		t.Fatal("protocol roundtrip failed")
	}
	p := Propagate(m, "N07", "N02", "reply", []float64{3})
	if p.TraceID != m.TraceID || p.CorrelationID != m.CorrelationID || p.Sequence != m.Sequence+1 {
		t.Fatal("trace/correlation propagation failed")
	}
	m.Deadline = time.Now().Add(-time.Second)
	if err := m.Validate(); err == nil {
		t.Fatal("expired deadline accepted")
	}
}
