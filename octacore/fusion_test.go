package octacore

import (
  "context"
  "encoding/json"
  "testing"

  "github.com/divibisoul/Orquestrador-/mesh"
  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/orchestrator"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestFusionUsesCanonicalFourStagePipeline(t *testing.T) {
  t.Setenv("N07_TCE_ENABLED", "true")
  n, err := neural.New(4, 0.05); if err != nil { t.Fatal(err) }
  c, err := prefrontal.New(0.01, 8); if err != nil { t.Fatal(err) }
  g := supergpu.New(nil); g.Discover()
  e, err := orchestrator.New(n, c, g); if err != nil { t.Fatal(err) }
  if err := orchestrator.RegisterAdvancedOperations(e); err != nil { t.Fatal(err) }
  if err := orchestrator.RegisterSuperGPUOperations(e); err != nil { t.Fatal(err) }
  if err := orchestrator.RegisterOrbitalReasoningOperations(e); err != nil { t.Fatal(err) }
  peers, err := mesh.NewPeerClient(nil); if err != nil { t.Fatal(err) }
  f, err := NewFusion(e, peers, g); if err != nil { t.Fatal(err) }; if err := f.Register(); err != nil { t.Fatal(err) }
  workloads := []map[string]any{{"ID":"octa-workload","Operation":"matrix","Precision":"fp16","MatrixSize":32,"BatchSize":1,"DataBytes":1024,"MemoryNeeded":1024,"Priority":50}}
  candidate := map[string]any{"ID":"octa-candidate","Cost":0.01,"Risk":0.01,"Utility":0.8,"Uncertainty":0.01,"Urgency":0.2,"Impact":0.2}
  wj, _ := json.Marshal(workloads); cj, _ := json.Marshal(candidate)
  result, err := e.Execute(context.Background(), OpFusionExecute, []float64{1,2,3,4}, map[string]string{"workloads_json":string(wj),"candidate_json":string(cj),"operation":"identity","correlation_id":"octa-fusion-correlation"})
  if err != nil { t.Fatal(err) }; if result.Status != "ok" { t.Fatalf("unexpected status: %#v", result) }
  for _, key := range []string{"orchestrator_stage","orbital_stage","prefrontal_stage","supergpu_stage"} { if result.Metadata[key] != "completed" { t.Fatalf("missing stage %s: %#v", key, result.Metadata) } }
  if result.Metadata["hardware_claim"] != "none" { t.Fatalf("hardware claim must remain non-fabricated: %#v", result.Metadata) }
}

func TestFusionFailsClosedWithoutTCE(t *testing.T) {
  t.Setenv("N07_TCE_ENABLED", "false")
  n, _ := neural.New(4, 0.05); c, _ := prefrontal.New(0.01, 8); g := supergpu.New(nil); peers, _ := mesh.NewPeerClient(nil)
  e, _ := orchestrator.New(n, c, g); _ = orchestrator.RegisterAdvancedOperations(e); _ = orchestrator.RegisterSuperGPUOperations(e); _ = orchestrator.RegisterOrbitalReasoningOperations(e)
  f, _ := NewFusion(e, peers, g); _ = f.Register()
  wj := `[{"ID":"w","Operation":"matrix","Precision":"fp16","MatrixSize":8,"BatchSize":1,"DataBytes":128,"MemoryNeeded":128}]`
  candidate := map[string]any{"ID":"c","Cost":0.01,"Risk":0.01,"Utility":0.5,"Uncertainty":0.01,"Urgency":0.1,"Impact":0.1}; cj, _ := json.Marshal(candidate)
  _, err := e.Execute(context.Background(), OpFusionExecute, []float64{1,2}, map[string]string{"workloads_json":wj,"candidate_json":string(cj),"correlation_id":"octa-blocked"})
  if err == nil { t.Fatal("fusion must fail closed when TCE is disabled") }
}