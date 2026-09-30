package orchestrator

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
	"github.com/divibisoul/Orquestrador-/octacore"
)

func TestBlueprintOperationsExposeAdditiveAffinityLayer(t *testing.T) {
	n, err := neural.New(2, .05)
	if err != nil { t.Fatal(err) }
	c, err := prefrontal.New(.1, 4)
	if err != nil { t.Fatal(err) }
	g := supergpu.New(nil)
	e, err := New(n, c, g)
	if err != nil { t.Fatal(err) }
	processor, err := octacore.NewProcessor(octacore.DefaultConfig(), g, nil, nil)
	if err != nil { t.Fatal(err) }
	if err := RegisterBlueprintOperations(e, processor); err != nil { t.Fatal(err) }
	result, err := e.Execute(context.Background(), "blueprint.resolve@1.0.0", nil, map[string]string{"query":"audio multimodal perception"})
	if err != nil { t.Fatal(err) }
	if result.Status != "ok" || result.Metadata["count"] == "0" {
		t.Fatalf("blueprint resolve failed: %#v", result)
	}
	_ = e.Execute(context.Background(), "blueprint.compose@1.0.0", nil, map[string]string{"query":"ethics governance memory"})
}
func TestBlueprintOperationMissingQueryFailsClosed(t *testing.T) {
	n, _ := neural.New(2, .05)
	c, _ := prefrontal.New(.1, 4)
	g := supergpu.New(nil)
	e, _ := New(n, c, g)
	processor, err := octacore.NewProcessor(octacore.DefaultConfig(), g, nil, nil)
	if err != nil { t.Fatal(err) }
	if err := RegisterBlueprintOperations(e, processor); err != nil { t.Fatal(err) }
	msg := protocol.NewMessage("N01", "N07", "command", "blueprint.resolve@1.0.0", nil)
	if _, err := e.Submit(context.Background(), msg); err == nil {
		t.Fatal("missing blueprint query must fail closed")
	}
}
