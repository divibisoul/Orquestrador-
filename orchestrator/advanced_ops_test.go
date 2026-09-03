package orchestrator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestSOULTopologyIsSevenNucleusAdjacentChain(t *testing.T) {
	top := SOULTopology()
	nuclei, ok := top["nuclei"].([]string); if !ok || len(nuclei)!=7 { t.Fatalf("invalid nuclei topology: %#v", top["nuclei"]) }
	if top["directional"] != 12 { t.Fatalf("expected 12 directional channels, got %v", top["directional"]) }
	if top["fusion_policy"] != "adjacent-only-dynamic" { t.Fatalf("unexpected fusion policy: %v", top["fusion_policy"]) }
}

func TestAdvancedOperationsExecuteWithRealServices(t *testing.T) {
	n, err := neural.New(4, .05); if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(.01, 8); if err != nil { t.Fatal(err) }
	g := supergpu.New(nil); g.Discover()
	e, err := New(n,c,g); if err != nil { t.Fatal(err) }
	if err := RegisterSuperGPUOperations(e); err != nil { t.Fatal(err) }
	if err := RegisterAdvancedOperations(e); err != nil { t.Fatal(err) }

	candidate := map[string]any{"ID":"safe-action","Cost":0.01,"Risk":0.01,"Utility":0.5,"Uncertainty":0.01,"Urgency":0.1,"Impact":0.1}
	candidateJSON, _ := json.Marshal(candidate)
	result, err := e.Execute(context.Background(), "prefrontal.admission@1.0.0", nil, map[string]string{"candidate_json":string(candidateJSON)})
	if err != nil { t.Fatal(err) }
	if result.Status != "ok" { t.Fatalf("unexpected admission status: %#v", result) }
	if len(c.Recall(1)) != 1 { t.Fatal("expected committed prefrontal decision") }

	gpuResult, err := e.Execute(context.Background(), "supergpu.federated.execute@1.0.0", []float64{2,3}, map[string]string{"nucleus":"N01","operation":"square"})
	if err != nil { t.Fatal(err) }
	if len(gpuResult.Payload)!=2 || gpuResult.Payload[0]!=4 || gpuResult.Payload[1]!=9 { t.Fatalf("unexpected federated compute: %#v", gpuResult.Payload) }
}
