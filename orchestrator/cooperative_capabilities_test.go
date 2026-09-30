package orchestrator

import (
  "context"
  "testing"

  "github.com/divibisoul/Orquestrador-/learning"
  "github.com/divibisoul/Orquestrador-/memory"
  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/protocol"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

type integrationMemory struct { records int; searches int }

func (m *integrationMemory) Record(_ context.Context, _ memory.Record) error { m.records++; return nil }
func (m *integrationMemory) Search(_ context.Context, _ string, _ []float64, _ float64, _ int) ([]memory.Match, error) { m.searches++; return nil, nil }

func newCooperativeEngine(t *testing.T) *Engine {
  t.Helper()
  n, err := neural.New(2, .05); if err != nil { t.Fatal(err) }
  c, err := prefrontal.New(.10, 32); if err != nil { t.Fatal(err) }
  e, err := New(n, c, supergpu.New(nil)); if err != nil { t.Fatal(err) }
  return e
}

func TestNeuralParametersIsCanonicalBuiltIn(t *testing.T) {
  e := newCooperativeEngine(t)
  m := protocol.NewMessage("N03", "N07", "command", "neural.parameters@1.0.0", nil)
  r, err := e.Submit(context.Background(), m)
  if err != nil { t.Fatal(err) }
  if r.Status != "ok" || r.Metadata["parameters"] == "" { t.Fatalf("unexpected result: %+v", r) }
}

func TestLearningBecomesCanonicalNeuralLearnExecutionPath(t *testing.T) {
  e := newCooperativeEngine(t)
  machine, err := learning.New(e.neural, e.cortex, nil); if err != nil { t.Fatal(err) }
  if err := e.SetLearningMachine(machine); err != nil { t.Fatal(err) }
  m := protocol.NewMessage("N06", "N07", "command", "neural.learn@1.0.0", []float64{.1,.2,.2,.3})
  r, err := e.Submit(context.Background(), m)
  if err != nil { t.Fatal(err) }
  if r.Status != "ok" { t.Fatalf("unexpected result: %+v", r) }
  if machine.Snapshot().Supervised != 1 { t.Fatalf("learning machine did not observe supervised execution: %+v", machine.Snapshot()) }
}

func TestMemoryRouteIsRegisteredAfterStoreAttachment(t *testing.T) {
  e := newCooperativeEngine(t)
  store := &integrationMemory{}
  if err := e.SetMemoryStore(store); err != nil { t.Fatal(err) }
  payload := make([]float64, memory.EmbeddingDimensions)
  m := protocol.NewMessage("N06", "N07", "command", "memory.record@1.0.0", payload)
  m.Metadata = map[string]string{
    "memory_user_id":"00000000-0000-0000-0000-000000000001",
    "memory_session_id":"00000000-0000-0000-0000-000000000002",
    "memory_summary":"test",
  }
  r, err := e.Submit(context.Background(), m)
  if err != nil { t.Fatal(err) }
  if r.Status != "ok" || store.records != 1 { t.Fatalf("memory route failed: %+v records=%d", r, store.records) }
}
