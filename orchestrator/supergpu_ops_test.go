package orchestrator

import (
	"context"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestSuperGPUExecuteRequiresExplicitOperation(t *testing.T) {
	n, err := neural.New(2, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterSuperGPUOperations(e); err != nil {
		t.Fatal(err)
	}

	message := protocol.NewMessage("N07", "N07", "command", "supergpu.execute@1.0.0", []float64{1, 2})
	message.CorrelationID = "corr-supergpu-missing-op"
	_, err = e.Submit(context.Background(), message)
	if err == nil || !strings.Contains(err.Error(), "metadata.operation is required") {
		t.Fatalf("expected explicit operation error, got %v", err)
	}
}

func TestCognitivePipelineRequiresExplicitOperation(t *testing.T) {
	n, err := neural.New(2, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}

	message := protocol.NewMessage("N07", "N07", "command", "cognitive.execute@1.0.0", []float64{1, 2})
	message.CorrelationID = "corr-cognitive-missing-op"
	_, err = e.Submit(context.Background(), message)
	if err == nil || !strings.Contains(err.Error(), "metadata.operation is required") {
		t.Fatalf("expected explicit operation error, got %v", err)
	}
}
