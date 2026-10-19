package neural

import (
	"context"
	"math"
	"testing"
)

func TestNetworkRuntime(t *testing.T) {
	n, err := New(3, .1)
	if err != nil {
		t.Fatal(err)
	}
	if err = n.AddEdge(0, 1, .5); err != nil {
		t.Fatal(err)
	}
	if _, err = n.Forward(context.Background(), []float64{1, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if _, err = n.Forward(context.Background(), []float64{1, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if n.Health()["cache_hits"].(uint64) == 0 {
		t.Fatal("forward cache not used")
	}
	if err = n.Learn([]float64{1, 0, 0}, []float64{0, 1, 0}); err != nil {
		t.Fatal(err)
	}
	if err = n.AddEdge(1, 2, .2); err != nil {
		t.Fatal(err)
	}
	if _, err = n.Normalize([]float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if _, err = n.Attention([]float64{1}, []float64{1, 2}, []float64{3, 4}); err != nil {
		t.Fatal(err)
	}
	if _, err = n.Backprop([]float64{1, 2, 3}, []float64{0, 1, 0}); err != nil {
		t.Fatal(err)
	}
	if err = n.RemoveEdge(0, 1); err != nil {
		t.Fatal(err)
	}
	if n.Health()["status"] != "ready" {
		t.Fatal("network unhealthy")
	}
}
func TestNetworkRejectsNonFiniteAndMalformedAttention(t *testing.T) {
	n, _ := New(2, .1)
	if _, err := n.Forward(context.Background(), []float64{math.NaN(), 0}); err == nil {
		t.Fatal("NaN accepted")
	}
	if _, err := n.Attention(nil, []float64{1}, []float64{1}); err == nil {
		t.Fatal("empty query accepted")
	}
	if _, err := n.Attention([]float64{1}, []float64{1, 2}, []float64{1}); err == nil {
		t.Fatal("mismatched attention vectors accepted")
	}
}

func TestNetworkConfigurePreservesExistingGraphAndRuntime(t *testing.T) {
	n, err := New(3, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	if err := n.AddEdge(0, 1, 0.5); err != nil {
		t.Fatal(err)
	}
	input := []float64{1, 0, 0}
	if _, err := n.Forward(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if _, err := n.Forward(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	before := n.Health()
	if before["cache_hits"].(uint64) == 0 {
		t.Fatal("expected the preserved forward cache to record a hit")
	}

	config := Config{
		Layers: []Layer{{Activation: "sigmoid"}},
		Optimizer: "sgd",
		Regularization: 1e-6,
		GradientClip: 1,
		Heads: 1,
		BatchCache: 8,
	}
	if err := n.Configure(config); err != nil {
		t.Fatal(err)
	}
	after := n.Health()
	if after["optimizer"] != "sgd" || after["edges"].(int) != 1 {
		t.Fatalf("configuration did not preserve graph or apply optimizer: before=%v after=%v", before, after)
	}
	if _, err := n.Forward(context.Background(), input); err != nil {
		t.Fatalf("forward failed after Configure: %v", err)
	}
	if n.Health()["cache_hits"].(uint64) != before["cache_hits"].(uint64) {
		t.Fatal("Configure did not invalidate the forward cache")
	}
}

func TestNetworkConfigureRejectsUnsupportedConfigurationWithoutMutation(t *testing.T) {
	n, err := New(2, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	before := n.Health()
	err = n.Configure(Config{
		Layers: []Layer{{Activation: "made-up-activation"}},
		Optimizer: "adam",
		Regularization: 1e-6,
		GradientClip: 1,
		Heads: 1,
		BatchCache: 8,
	})
	if err == nil {
		t.Fatal("unsupported activation was accepted")
	}
	after := n.Health()
	if after["optimizer"] != before["optimizer"] || after["heads"] != before["heads"] {
		t.Fatalf("invalid configuration mutated live network: before=%v after=%v", before, after)
	}
}
