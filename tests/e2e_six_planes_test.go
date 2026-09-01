package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/neuralfabric"
	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/core/prefrontal"
	"github.com/divibisoul/Orquestrador-/core/soul"
	"github.com/divibisoul/Orquestrador-/core/superagi"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/mesh/transport"
	"github.com/divibisoul/Orquestrador-/state"
)

func TestSixPlanesE2E(t *testing.T) {
	ctx := context.Background()
	cortex := prefrontal.New()
	plan := cortex.GeneratePlan("recover orchestrator", []string{"discover", "route", "execute"})
	decision := cortex.Decide([]prefrontal.Plan{plan}, 10)
	if decision.ID != plan.ID {
		t.Fatalf("cortex selected %q, want %q", decision.ID, plan.ID)
	}

	engine := orchestrator.NewEngine(2)
	steps := []orchestrator.Step{{ID: "discover", Run: func(context.Context) error { return nil }}, {ID: "route", Run: func(context.Context) error { return nil }}, {ID: "execute", Run: func(context.Context) error { return nil }}}
	if err := engine.CreateWorkflow("recovery-e2e", steps); err != nil {
		t.Fatal(err)
	}
	for range steps {
		if err := engine.ExecuteStep(ctx, "recovery-e2e"); err != nil {
			t.Fatal(err)
		}
	}
	status, err := engine.GetWorkflowStatus("recovery-e2e")
	if err != nil || status != orchestrator.Completed {
		t.Fatalf("workflow status=%q err=%v", status, err)
	}

	registry := mesh.NewRegistry()
	if err := registry.Announce(mesh.Node{ID: "local-node", Status: "ready", Capabilities: []string{"orchestration", "inference"}, CPU: true}); err != nil {
		t.Fatal(err)
	}
	if nodes := registry.Discover(ctx, "inference"); len(nodes) != 1 {
		t.Fatalf("mesh discovered %d inference nodes, want 1", len(nodes))
	}

	nf := neuralfabric.NewRuntime()
	route, err := nf.Route(ctx, neuralfabric.State{Goal: "recover orchestrator"}, []neuralfabric.Route{{NodeID: "local-node", DeviceID: "local-node", Precision: "int8", BatchSize: 1}})
	if err != nil || route.NodeID != "local-node" {
		t.Fatalf("neural fabric route=%+v err=%v", route, err)
	}

	agi := superagi.NewWithProvider(testProvider{})
	embedding := agi.GenerateEmbedding("recovery-e2e")
	if len(embedding) != 4 {
		t.Fatalf("embedding length=%d", len(embedding))
	}

	store := state.NewStore()
	version := store.Put("workflow", []byte("recovery-e2e"))
	value, gotVersion, ok := store.Get("workflow")
	if !ok || gotVersion != version || string(value) != "recovery-e2e" {
		t.Fatalf("state read value=%q version=%d ok=%v", value, gotVersion, ok)
	}
	if !store.CompareAndSwap("workflow", version, []byte("completed")) {
		t.Fatal("state CAS failed")
	}

	var mu sync.Mutex
	received := make(map[string]int)
	servers := make([]*httptest.Server, 0, 6)
	endpoints := make(map[soul.NucleusID]string, 6)
	for _, id := range []soul.NucleusID{soul.N01, soul.N02, soul.N03, soul.N04, soul.N05, soul.N06} {
		nucleusID := id
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			var in transport.Envelope
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if in.Target != string(nucleusID) || in.EventType != "soul.task.request" || in.TraceID == "" {
				http.Error(w, "invalid envelope", http.StatusBadRequest)
				return
			}
			mu.Lock()
			received[string(nucleusID)]++
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(transport.Envelope{EventID: in.EventID + "-response", EventType: "soul.task.response", TraceID: in.TraceID, Source: string(nucleusID), Target: "orquestrador", Timestamp: in.Timestamp, CompatibleSystems: []string{"SOUL", soul.MeshProtocol}, Payload: map[string]interface{}{"accepted": true}})
		}))
		servers = append(servers, srv)
		endpoints[nucleusID] = srv.URL
	}
	defer func() {
		for _, srv := range servers {
			srv.Close()
		}
	}()

	soulFabric := soul.NewFabric(engine, registry, transport.Client{})
	if err := soulFabric.RegisterDefaults(endpoints); err != nil {
		t.Fatal(err)
	}
	for _, id := range []soul.NucleusID{soul.N01, soul.N02, soul.N03, soul.N04, soul.N05, soul.N06} {
		traceID := "trace-" + string(id)
		response, err := soulFabric.Dispatch(ctx, id, traceID, "recover orchestrator", map[string]interface{}{"plane": "e2e"})
		if err != nil {
			t.Fatalf("dispatch %s: %v", id, err)
		}
		if response.TraceID != traceID || response.Source != string(id) {
			t.Fatalf("dispatch %s response=%+v", id, response)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	for _, id := range []soul.NucleusID{soul.N01, soul.N02, soul.N03, soul.N04, soul.N05, soul.N06} {
		if received[string(id)] != 1 {
			t.Fatalf("nucleus %s received %d requests, want 1", id, received[string(id)])
		}
	}

	t.Log("six-plane E2E: cortex -> orchestrator -> mesh -> neural fabric -> super AGI -> state -> SOUL N01-N06 transport")
}
