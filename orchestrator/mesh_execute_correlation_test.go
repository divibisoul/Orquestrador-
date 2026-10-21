package orchestrator

import (
  "context"
  "testing"

  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/protocol"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestExecuteWithCorrelationPreservesCallerCorrelation(t *testing.T) {
  n, err := neural.New(8, 0.05)
  if err != nil { t.Fatal(err) }
  c, err := prefrontal.New(0.1, 32)
  if err != nil { t.Fatal(err) }
  g := supergpu.New(nil)
  e, err := New(n, c, g)
  if err != nil { t.Fatal(err) }

  const op = "test.correlation@1.0.0"
  if err := e.Register(op, func(_ context.Context, m protocol.Message) (protocol.Result, error) {
    return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "test", Target: m.Source, Status: "ok"}, nil
  }); err != nil { t.Fatal(err) }

  result, err := e.ExecuteWithCorrelation(context.Background(), "corr-n07-execute-001", "SARA", op, nil, nil)
  if err != nil { t.Fatal(err) }
  if result.CorrelationID != "corr-n07-execute-001" {
    t.Fatalf("correlationId = %q, want %q", result.CorrelationID, "corr-n07-execute-001")
  }
}
