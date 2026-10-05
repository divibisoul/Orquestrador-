package n07_test

import (
  "context"
  "testing"
  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestRecoveredNativePublicAPIs(t *testing.T) {
  n, err := neural.New(4, 0.05); if err != nil { t.Fatal(err) }
  if err := n.Configure(neural.Config{Layers: []neural.Layer{{Activation:"tanh"}}, Optimizer:"adam", Regularization:1e-6, GradientClip:1, Heads:1, BatchCache:8}); err != nil { t.Fatal(err) }
  c, err := prefrontal.New(0.1, 4); if err != nil { t.Fatal(err) }
  _ = c.HistorySize()
  ctx := supergpu.WithCorrelationID(context.Background(), "corr-n07")
  if supergpu.CorrelationIDFromContext(ctx) != "corr-n07" { t.Fatal("correlation context not preserved") }
}
