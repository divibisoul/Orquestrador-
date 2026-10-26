package main

import (
	"context"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestN01PrefrontalBridgeRegistersAndFailsClosedWhenPeerIsUnconfigured(t *testing.T) {
	t.Setenv("SOUL_MESH_N01_URL", "")
	t.Setenv("SOUL_MESH_HMAC_SECRET", "")
	n, err := neural.New(8, .05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(.10, 8)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	engine, err := orchestrator.New(n, c, g)
	if err != nil { t.Fatal(err) }
	peers, err := mesh.NewPeerClient(nil)
	if err != nil { t.Fatal(err) }
	if err := registerN01PrefrontalBridge(engine, peers); err != nil { t.Fatal(err) }
	found := false
	for _, operation := range engine.Operations() {
		if operation == n01PrefrontalOptimizeCapability { found = true; break }
	}
	if !found { t.Fatal("N01 prefrontal bridge operation not registered") }

	_, err = engine.Execute(context.Background(), n01PrefrontalOptimizeCapability, nil, map[string]string{
		"correlation_id": "corr-bridge-unconfigured",
		"trace_id": "trace-bridge-unconfigured",
		"goal_description": "Validate the N01 prefrontal optimizer boundary",
		"weak_point": "Mesh correlation",
	})
	if err == nil || !strings.Contains(err.Error(), "peer not configured") {
		t.Fatalf("expected fail-closed missing-peer error; got %v", err)
	}
}
