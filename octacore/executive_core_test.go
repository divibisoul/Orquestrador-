package octacore

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/executor"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestUnifiedExecutiveCoreMeshE2E(t *testing.T) {
	t.Setenv("N07_TCE_ENABLED", "true")
	t.Setenv("N07_MESH_HMAC_SECRET", "")
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "true")

	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	cortex, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	gpu := supergpu.New(nil)
	gpu.Discover()
	engine, err := orchestrator.New(n, cortex, gpu)
	if err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterSuperGPUOperations(engine); err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterAdvancedOperations(engine); err != nil {
		t.Fatal(err)
	}
	tce, err := orchestrator.NewTranscendentalSimulatorFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if tce == nil {
		t.Fatal("expected enabled TCE simulator")
	}
	if err := orchestrator.RegisterOrbitalReasoningOperationsWithSimulator(engine, tce); err != nil {
		t.Fatal(err)
	}
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	neocortex, err := prefrontal.NewNeocortex(cortex, n)
	if err != nil {
		t.Fatal(err)
	}
	processor, err := NewProcessor(DefaultConfig(), gpu, peers, nil)
	if err != nil {
		t.Fatal(err)
	}
	unifiedCore, err := NewExecutiveCore(engine, neocortex, tce, gpu, peers)
	if err != nil {
		t.Fatal(err)
	}
	if err := processor.SetExecutiveCore(unifiedCore); err != nil {
		t.Fatal(err)
	}
	if err := RegisterOperations(engine, processor); err != nil {
		t.Fatal(err)
	}

	workload := core.Workload{
		ID: "mesh-octa-workload",
		Operation: "identity",
		Precision: core.FP32,
		MatrixSize: 256,
		BatchSize: 1,
		DataBytes: 4096,
		MemoryNeeded: 1,
	}
	request := ExecutiveCoreRequest{
		JobID: "mesh-octa-job",
		CorrelationID: "mesh-octa-correlation",
		TaskID: "mesh-octa-task",
		Operation: "identity",
		Input: []float64{1, 2, 3, 4},
		Workloads: []core.Workload{workload},
		Risk: 0,
		Cost: 0,
		Urgency: 0.1,
		Impact: 0.1,
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	wire := map[string]any{
		"protocol": "soul-mesh/1",
		"contractVersion": protocol.SoulMeshContractVersion,
		"id": "mesh-octa-message",
		"correlationId": "mesh-octa-correlation",
		"source": "N01",
		"target": "N07",
		"kind": "request",
		"capability": OpCoreExecute,
		"payload": map[string]any{"values": request.Input},
		"metadata": map[string]string{"octacore_core_request_json": string(encoded)},
		"timestamp": time.Now().UnixMilli(),
		"nonce": "mesh-octa-nonce",
	}
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}

	gateway := mesh.NewHTTPGateway(engine)
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Mesh execution failed: code=%d body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		CorrelationID string `json:"correlationId"`
		Payload struct {
			Status   string            `json:"status"`
			Metadata map[string]string `json:"metadata"`
		} `json:"payload"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.CorrelationID != request.CorrelationID {
		t.Fatalf("correlation mismatch: %s", response.CorrelationID)
	}
	if response.Payload.Status != "ok" {
		t.Fatalf("unexpected core status: %#v", response.Payload)
	}

	description := unifiedCore.Describe()
	if description["type"] != "unified_logical_processing_core" {
		t.Fatalf("unexpected unified core type: %#v", description["type"])
	}
	if description["communication"] != "SOUL Mesh" {
		t.Fatalf("unexpected communication contract: %#v", description["communication"])
	}
}

func TestUnifiedExecutiveCoreUsesExplicitEightLogicalLanes(t *testing.T) {
	n, _ := neural.New(8, .05)
	cortex, _ := prefrontal.New(.1, 8)
	gpu := supergpu.New(nil)
	gpu.Discover()
	engine, _ := orchestrator.New(n, cortex, gpu)
	peers, _ := mesh.NewPeerClient(nil)
	neocortex, _ := prefrontal.NewNeocortex(cortex, n)
	cfg := core.DefaultConfig()
	cfg.Enabled = true
	cfg.Simulation.EfficiencyFactor = .7
	tceEngine, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	tce, err := executor.New(tceEngine, models.DefaultCatalog(cfg), cfg.Mode)
	if err != nil {
		t.Fatal(err)
	}
	unifiedCore, err := NewExecutiveCore(engine, neocortex, tce, gpu, peers)
	if err != nil {
		t.Fatal(err)
	}
	description := unifiedCore.Describe()
	if description["topology"] != "4-components-x-2-lanes=8-logical-lanes" {
		t.Fatalf("unexpected topology: %#v", description["topology"])
	}
	components, ok := description["components"].([]string)
	if !ok || len(components) != 4 {
		t.Fatalf("unexpected component list: %#v", description["components"])
	}
	lanes, ok := description["lanes"].([]ExecutiveLane)
	if !ok || len(lanes) != 8 {
		t.Fatalf("expected eight executable logical lanes, got %#v", description["lanes"])
	}
	for _, lane := range lanes {
		if lane.ID == "" || lane.Component == "" || lane.Role == "" || lane.Transport == "" || !lane.Required {
			t.Fatalf("invalid executive lane: %#v", lane)
		}
	}
}
