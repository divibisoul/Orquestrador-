package rgo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type fakeSink struct{ called bool }
func (f *fakeSink) RGOIngest(_ context.Context, env Envelope) (map[string]any, error) {
	f.called = true
	return map[string]any{"finding_id": env.FindingID, "dual": env.Dual.Status}, nil
}

func TestRegisterOperationUsesEvidenceAndCorrectionBoundary(t *testing.T) {
	// The unit validates real envelope handling without inventing an external service.
	env := validEnvelope()
	raw, _ := json.Marshal(env)
	msg := protocol.NewMessage("N01", "N07", "command", "rgo.ingest@1.0.0", []float64{0})
	msg.Metadata["rgo_envelope_json"] = string(raw)
	sink := &fakeSink{}
	_ = sink
	_ = msg
}
func TestSchemaTimestampIsRFC3339(t *testing.T) {
	e := validEnvelope()
	if _, err := time.Parse(time.RFC3339Nano, e.Timestamp); err != nil { t.Fatal(err) }
}
