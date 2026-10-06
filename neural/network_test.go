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


func TestBackpropMatchesFiniteDifference(t *testing.T) {
	n, err := New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	if err := n.AddEdge(0, 1, .2); err != nil {
		t.Fatal(err)
	}
	inputs := []float64{0.3, -0.2}
	target := []float64{0.1, 0.2}
	gradient, err := n.Backprop(inputs, target)
	if err != nil {
		t.Fatal(err)
	}

	lossAt := func(x []float64) float64 {
		pred, err := n.Forward(context.Background(), x)
		if err != nil {
			t.Fatal(err)
		}
		loss, err := mseLoss(pred, target)
		if err != nil {
			t.Fatal(err)
		}
		return loss
	}
	const epsilon = 1e-6
	for i := range inputs {
		plus := append([]float64(nil), inputs...)
		minus := append([]float64(nil), inputs...)
		plus[i] += epsilon
		minus[i] -= epsilon
		numerical := (lossAt(plus) - lossAt(minus)) / (2 * epsilon)
		if math.Abs(numerical-gradient[i]) > 1e-4 {
			t.Fatalf("gradient[%d] mismatch: analytical=%g numerical=%g", i, gradient[i], numerical)
		}
	}
}

func TestLearnUsesConfiguredGraphAndReducesLoss(t *testing.T) {
	n, err := New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	if err := n.Configure(Config{
		Layers:         []Layer{{Activation: "linear"}},
		Optimizer:      "sgd",
		Regularization: 0,
		GradientClip:   0,
		Heads:          1,
		BatchCache:     8,
	}); err != nil {
		t.Fatal(err)
	}
	inputs := []float64{0, 0}
	target := []float64{1, -1}
	lossAt := func() float64 {
		pred, err := n.Forward(context.Background(), inputs)
		if err != nil {
			t.Fatal(err)
		}
		loss, err := mseLoss(pred, target)
		if err != nil {
			t.Fatal(err)
		}
		return loss
	}
	before := lossAt()
	for i := 0; i < 20; i++ {
		if err := n.Learn(inputs, target); err != nil {
			t.Fatal(err)
		}
	}
	after := lossAt()
	if !(after < before) {
		t.Fatalf("training did not reduce loss: before=%g after=%g", before, after)
	}
	if got := n.Health()["learning_steps"].(uint64); got != 20 {
		t.Fatalf("expected 20 learning steps, got %d", got)
	}
}

func TestNetworkConfigureValidatesExecutableConfig(t *testing.T) {
	n, err := New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	invalid := []Config{
		{Layers: nil, Optimizer: "adam", Heads: 1, BatchCache: 1},
		{Layers: []Layer{{Activation: "unknown"}}, Optimizer: "adam", Heads: 1, BatchCache: 1},
		{Layers: []Layer{{Activation: "tanh", DropoutRate: 1}}, Optimizer: "adam", Heads: 1, BatchCache: 1},
		{Layers: []Layer{{Activation: "tanh"}}, Optimizer: "unknown", Heads: 1, BatchCache: 1},
	}
	for i, config := range invalid {
		if err := n.Configure(config); err == nil {
			t.Fatalf("invalid config %d was accepted", i)
		}
	}
}

func TestLearnExecutesDropoutAndRemainsFinite(t *testing.T) {
	n, err := New(4, .03)
	if err != nil {
		t.Fatal(err)
	}
	if err := n.Configure(Config{
		Layers: []Layer{
			{Activation: "tanh", DropoutRate: .25},
			{Activation: "linear", DropoutRate: .10},
		},
		Optimizer:      "adam",
		Regularization: 1e-6,
		GradientClip:   1.0,
		Heads:          1,
		BatchCache:     16,
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if err := n.Learn(
			[]float64{1, 0, -1, .5},
			[]float64{.2, -.1, .3, .4},
		); err != nil {
			t.Fatal(err)
		}
	}
	health := n.Health()
	if health["dropout_enabled"] != true {
		t.Fatalf("dropout was configured but not reported enabled: %#v", health)
	}
	if loss, ok := health["last_loss"].(float64); !ok || math.IsNaN(loss) || math.IsInf(loss, 0) {
		t.Fatalf("dropout training produced invalid loss: %#v", health["last_loss"])
	}
}

func TestNormalizeRemainsFiniteAtFloatExtremes(t *testing.T) {
	n, err := New(3, .05)
	if err != nil {
		t.Fatal(err)
	}
	values := []float64{math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64 / 2}
	out, err := n.Normalize(values)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range out {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("normalization overflowed at index %d: %v", i, value)
		}
	}
	zeros, err := n.Normalize([]float64{0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range zeros {
		if value != 0 {
			t.Fatalf("zero normalization changed at index %d: %v", i, value)
		}
	}
}
