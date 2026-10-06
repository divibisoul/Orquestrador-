package aeternum

import (
	"context"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newHortaCoreForTest(t *testing.T) *HortaCore {
	t.Helper()
	n, err := neural.New(4, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterSuperGPUOperations(e); err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterAdvancedOperations(e); err != nil {
		t.Fatal(err)
	}
	h, err := NewHortaCore(e, nil)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestCatalogIsCompleteAndNonDuplicated(t *testing.T) {
	if err := ValidateCatalog(); err != nil {
		t.Fatal(err)
	}
	h := newHortaCoreForTest(t)
	if len(h.Capabilities()) != 29 {
		t.Fatalf("expected 29 modules, got %d", len(h.Capabilities()))
	}
}

func TestHortaCoreDoesNotFabricateBlockedModuleExecution(t *testing.T) {
	h := newHortaCoreForTest(t)
	if _, err := h.Execute(context.Background(), "einstein_quantum", []float64{1, 2, 3, 4}, nil); err == nil {
		t.Fatal("blocked quantum capability must never report synthetic success")
	}
}

func TestHortaCoreUsesOnlyRegisteredNativeOperation(t *testing.T) {
	h := newHortaCoreForTest(t)
	result, err := h.Execute(context.Background(), "uci", nil, nil)
	if err != nil {
		t.Fatalf("real Mesh health operation should execute: %v", err)
	}
	if result["status"] != string(StatusNative) {
		t.Fatalf("unexpected module status: %#v", result["status"])
	}
}

func TestN02AdapterMappingsAreExplicitAndFailClosed(t *testing.T) {
	h := newHortaCoreForTest(t)

	cases := map[string]string{
		"bnc_v2": "neural.bnc_v2",
		"csae":   "cognitive.csae",
		"dcrs":   "resource.dcrs",
	}
	for moduleID, want := range cases {
		got, ok := h.AdapterCapability(moduleID)
		if !ok {
			t.Fatalf("%s should expose an explicit adapter capability", moduleID)
		}
		if got != want {
			t.Fatalf("%s maps to %q, want %q", moduleID, got, want)
		}
		_, err := h.Execute(context.Background(), moduleID, []float64{1, 2, 3, 4}, map[string]string{"input": "continuity-check"})
		if err == nil || !strings.HasPrefix(err.Error(), "AETERNUM_PEER_REQUIRED:N02:") {
			t.Fatalf("%s should fail closed without a configured Mesh peer, got %v", moduleID, err)
		}
	}
}


func TestHortaCoreHealthDoesNotReportReadyWhenAdapterMeshIsAbsent(t *testing.T) {
	t.Helper()
	n, err := neural.New(4, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}

	sara := backend.NewSARAProxy(backend.Config{
		SARAServiceURL:   "http://sara.local",
		SARAServiceToken: "configured-for-health-test",
	})
	h, err := NewHortaCore(e, sara)
	if err != nil {
		t.Fatal(err)
	}

	health := h.Health()
	if health["status"] != "DEGRADED" {
		t.Fatalf("HortaCore must not report READY without Mesh when adapter modules exist: %#v", health)
	}
	if health["adapter_modules"] != 23 {
		t.Fatalf("unexpected adapter module count: %#v", health["adapter_modules"])
	}
	if health["peer_client_attached"] != false {
		t.Fatalf("peer attachment state was fabricated: %#v", health["peer_client_attached"])
	}
	if health["executable_adapter_modules"] != 3 {
		t.Fatalf("unexpected executable adapter count: %#v", health["executable_adapter_modules"])
	}
	if health["peer_client_configured"] != false || health["adapter_execution_available"] != false {
		t.Fatalf("adapter execution was falsely reported available: %#v", health)
	}
}
