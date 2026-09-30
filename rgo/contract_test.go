package rgo

import (
	"strings"
	"testing"
	"time"
)

func validEnvelope() Envelope {
	var e Envelope
	e.SchemaVersion = SchemaVersion
	e.FindingID = "finding-1"
	e.ObjectID = "object-1"
	e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	e.CorrelationID = "corr-1"
	e.TraceID = "trace-1"
	e.Source.System = "BugShield"
	e.Source.Module = "Scanner"
	e.Source.Version = "1.2.0"
	e.Epistemic.Mode = Inspection
	e.Epistemic.Verification = Verified
	e.Actionability.Status = Actionable
	e.Failure.Type = "BUG"
	e.Failure.Description = "known failing behavior"
	e.Failure.Nature = "execution"
	e.CorrectionBoundary.ProblemToResolve = "prevent failing behavior"
	e.CorrectionBoundary.RequiredProperty = "deterministic validation before execution"
	e.Evidence = []EvidenceRef{{ID: "ev-1", Kind: "test", Ref: "test://case-1"}}
	e.Provenance.Origin = "test"
	e.Provenance.InputHash = "sha256:test"
	return DeriveDual(e)
}

func TestEnvelopeValidationAndDual(t *testing.T) {
	e := validEnvelope()
	if err := e.Validate(); err != nil { t.Fatal(err) }
	if e.Dual.Status != DualDerived || e.Dual.Property == "" { t.Fatalf("dual not derived: %#v", e.Dual) }
	if len(e.Dual.EvidenceRefs) != 1 || e.Dual.EvidenceRefs[0] != "ev-1" { t.Fatalf("dual evidence not bound: %#v", e.Dual.EvidenceRefs) }
	h, err := CanonicalHash(e)
	if err != nil || !strings.HasPrefix(h, "sha256:") { t.Fatalf("hash failure: %v %q", err, h) }
}

func TestMissingRequiredPropertyDoesNotInventDual(t *testing.T) {
	e := validEnvelope()
	e.CorrectionBoundary.RequiredProperty = ""
	e = DeriveDual(e)
	if e.Dual.Status != DualUnresolved { t.Fatalf("expected unresolved dual, got %s", e.Dual.Status) }
}
