package neural

import "testing"

func TestParametersSnapshotReturnsActualConfiguration(t *testing.T) {
	n, err := New(8, 0.05)
	if err != nil { t.Fatal(err) }
	p := n.Parameters()
	if p.Size != 8 || p.LearningRate != 0.05 || p.Optimizer != "adam" || p.Heads != 1 || p.BatchCache != 128 {
		t.Fatalf("unexpected parameter snapshot: %+v", p)
	}
	if len(p.Layers) != 1 || p.Layers[0].Activation != "tanh" { t.Fatalf("unexpected layers: %+v", p.Layers) }
}
