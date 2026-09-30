package neural

import "testing"

func TestParametersSnapshotReturnsActualConfiguration(t *testing.T) {
	n, err := New(8, 0.05)
	if err != nil { t.Fatal(err) }
	p := n.Parameters()
	if p["size"] != 8 || p["learning_rate"] != 0.05 || p["optimizer"] != "adam" || p["heads"] != 1 || p["batch_cache"] != 128 {
		t.Fatalf("unexpected parameter snapshot: %#v", p)
	}
	layers, ok := p["layers"].([]map[string]any)
	if !ok || len(layers) != 1 || layers[0]["activation"] != "tanh" {
		t.Fatalf("unexpected layers: %#v", p["layers"])
	}
}
