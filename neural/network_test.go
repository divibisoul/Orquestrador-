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

func TestNetworkExposesCanonicalParameters(t *testing.T) {
	n, err := New(3, .1)
	if err != nil {
		t.Fatal(err)
	}
	params := n.Parameters()
	if params.Size != 3 || params.LearningRate != .1 {
		t.Fatalf("unexpected core parameters: %+v", params)
	}
	if params.Optimizer != "adam" || params.Regularization != 1e-6 || params.GradientClip != 1.0 || params.Heads != 1 || params.BatchCache != 128 {
		t.Fatalf("unexpected optimizer parameters: %+v", params)
	}
	if len(params.Layers) != 1 || params.Layers[0].Activation != "tanh" {
		t.Fatalf("unexpected layer parameters: %+v", params.Layers)
	}
	health := n.Health()
	if health["learning_rate"] != .1 || health["regularization"] != 1e-6 || health["gradient_clip"] != 1.0 || health["batch_cache"] != 128 {
		t.Fatalf("health does not expose current parameters: %#v", health)
	}
}

func TestNetworkUsesCanonicalDefaultConfiguration(t *testing.T) {
	n, err := NewDefault()
	if err != nil {
		t.Fatal(err)
	}
	params := n.Parameters()
	if params.Size != DefaultNetworkSize || params.LearningRate != DefaultNetworkLearningRate {
		t.Fatalf("unexpected default parameters: %+v", params)
	}
}
