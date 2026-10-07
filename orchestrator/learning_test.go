package orchestrator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/divibisoul/Orquestrador-/learning"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestLearningReceiverAndNeuralParameters(t *testing.T) {
	n, err := neural.New(8, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	m, err := learning.New(n, c, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterLearningOperations(e, m, n); err != nil {
		t.Fatal(err)
	}

	feedback := protocol.NewMessage("N02", "N07", "command", "learning.feedback@1.0.0", []float64{0.8, 0.9})
	feedback.TraceID = "trace-learning-test"
	feedback.CorrelationID = "corr-learning-test"
	feedback.Metadata = map[string]string{
		"learning_target":     "N07",
		"learning_capability": "neural.forward",
		"learning_outcome":    "success",
		"learning_provenance": "mesh-observed",
	}
	result, err := e.Submit(context.Background(), feedback)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("unexpected learning result: %#v", result)
	}
	if m.Weight("N02", "N07", "neural.forward") <= 0.5 {
		t.Fatal("learning route weight did not change")
	}
	observations := c.LearningObservations(0)
	if len(observations) != 1 || observations[0].Capability != "neural.forward" || observations[0].Outcome != "success" {
		t.Fatalf("prefrontal learning observation was not recorded: %#v", observations)
	}

	parameters := protocol.NewMessage("N02", "N07", "command", "neural.parameters@1.0.0", nil)
	parameters.TraceID = "trace-parameters-test"
	parameters.CorrelationID = "corr-parameters-test"
	pResult, err := e.Submit(context.Background(), parameters)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(pResult.Metadata["parameters"]), &decoded); err != nil {
		t.Fatal(err)
	}
	if int(decoded["size"].(float64)) != 8 {
		t.Fatalf("unexpected neural size: %#v", decoded["size"])
	}
}
