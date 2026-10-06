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
