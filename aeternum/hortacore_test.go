package aeternum

import (
	"context"
	"testing"

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
