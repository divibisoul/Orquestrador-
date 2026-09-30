package orchestrator

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newCognitiveRecoveryEngine(t *testing.T) (*Engine, *mesh.PeerClient) {
	t.Helper()
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	return e, peers
}

func TestRecoveredCognitiveOperationsAreRegisteredWithoutRemoteExecution(t *testing.T) {
	e, peers := newCognitiveRecoveryEngine(t)
	if err := RegisterCognitiveOperations(e, peers, backend.NewSARAProxy(backend.Config{}), backend.NewSupabaseStore(backend.Config{})); err != nil {
		t.Fatal(err)
	}

	health, err := e.Execute(context.Background(), "cognitive.health@1.0.0", nil, map[string]string{
		"correlation_id": "cognitive-health-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if health.Status != "ok" {
		t.Fatalf("unexpected health status: %#v", health)
	}
	if !strings.Contains(health.Metadata["cognitive_json"], "READY") {
		t.Fatalf("unexpected cognitive health payload: %#v", health.Metadata)
	}
}

func TestRecoveredCognitivePlanFailsClosedWhenCapabilityIsNotDiscoverable(t *testing.T) {
	e, peers := newCognitiveRecoveryEngine(t)
	if err := RegisterCognitiveOperations(e, peers, backend.NewSARAProxy(backend.Config{}), backend.NewSupabaseStore(backend.Config{})); err != nil {
		t.Fatal(err)
	}

	goal, _ := json.Marshal(map[string]any{
		"goal_id": "plan-test",
		"objective": "discover an unconfigured capability",
		"capabilities": []string{"capability.not.configured"},
		"correlation_id": "corr-plan-test",
	})
	_, err := e.Execute(context.Background(), "cognitive.goal.plan@1.0.0", nil, map[string]string{
		"cognitive_goal_json": string(goal),
	})
	if err == nil {
		t.Fatal("planner must not fabricate an unconfigured capability")
	}
	if !strings.Contains(err.Error(), "TOOL_NOT_DISCOVERED") {
		t.Fatalf("unexpected planner error: %v", err)
	}
}
