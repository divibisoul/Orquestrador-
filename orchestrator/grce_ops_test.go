package orchestrator

import (
	"context"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/grf"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newGRFTestEngine(t *testing.T) *Engine {
	t.Helper()
	n, err := neural.New(4, .05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(.1, 8)
	if err != nil { t.Fatal(err) }
	e, err := New(n,c,supergpu.New(nil))
	if err != nil { t.Fatal(err) }
	if err := RegisterGRFGRCEOperations(e); err != nil { t.Fatal(err) }
	return e
}

func TestGRFAndGRCEOperationsRegistered(t *testing.T) {
	e := newGRFTestEngine(t)
	for _, op := range []string{
		GRFDescribeOperation, GRCEDescribeOperation, GRCEBindingsOperation,
		GRCECycleOperation, GRFParticipantDescribeOperation,
	} {
		if !containsOperation(e.Operations(), op) {
			t.Fatalf("missing operation %s", op)
		}
	}
}

func TestGRCECycleFailsClosedUntilRealHooksAreBound(t *testing.T) {
	e := newGRFTestEngine(t)
	m := map[string]string{"correlation_id":"grf-test"}
	_, err := e.Execute(context.Background(), GRCECycleOperation, nil, m)
	if err == nil || !strings.Contains(err.Error(), "GRF_TRACE_ID_REQUIRED") {
		t.Fatalf("expected context gate, got %v", err)
	}
	msg, err := e.Execute(context.Background(), GRCECycleOperation, nil, map[string]string{"trace_id":"t","correlation_id":"c"})
	if err == nil || !strings.Contains(err.Error(), "GRCE_RUNTIME_HOOKS_NOT_BOUND") {
		t.Fatalf("expected explicit blocked state, got %v", err)
	}
}

func TestGRFInvariantSetContainsI1ThroughI18(t *testing.T) {
	set := grf.CanonicalInvariantSet()
	if err := grf.ValidateInvariantSet(set); err != nil { t.Fatal(err) }
	if len(set.Invariants) != 18 { t.Fatalf("invariants=%d",len(set.Invariants)) }
}
