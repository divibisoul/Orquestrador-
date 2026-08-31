package neuralfabric

import (
	"context"
	"sync"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

func TestFabricRoute(t *testing.T) {
	f := NewFabric(trinity.FabricConfig{})
	r, err := f.Route(context.Background(), trinity.Strategy{Precision: "fp32"}, trinity.Workload{ID: "x", Kind: "matrix", MatrixSize: 5000})
	if err != nil || r.Model == "" {
		t.Fatalf("route=%+v err=%v", r, err)
	}
	if r.Model != "vera_rubin" {
		t.Fatalf("expected matrix heuristic vera_rubin, got %q", r.Model)
	}
}

func TestFabricHeuristics(t *testing.T) {
	f := NewFabric(trinity.FabricConfig{})
	cases := []struct {
		name string
		w    trinity.Workload
		want string
	}{
		{"fp64", trinity.Workload{ID: "1", Kind: "matrix", Precision: "fp64"}, "blackwell"},
		{"memory", trinity.Workload{ID: "2", Kind: "matrix", MemoryNeeded: 401}, "mi400"},
		{"large", trinity.Workload{ID: "3", Kind: "matrix", MatrixSize: 4097}, "vera_rubin"},
		{"small", trinity.Workload{ID: "4", Kind: "matrix", MatrixSize: 512}, "trillium"},
		{"batch", trinity.Workload{ID: "5", Kind: "matrix", BatchSize: 65}, "blackwell"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := f.Route(context.Background(), trinity.Strategy{}, tc.w)
			if err != nil {
				t.Fatal(err)
			}
			if r.Model != tc.want {
				t.Fatalf("got %q want %q", r.Model, tc.want)
			}
		})
	}
}

func TestFabricFeedback(t *testing.T) {
	f := NewFabric(trinity.FabricConfig{})
	r, err := f.Route(context.Background(), trinity.Strategy{}, trinity.Workload{ID: "x", Kind: "text"})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Feedback(context.Background(), trinity.Feedback{WorkloadID: "x", Route: r, Success: true, Quality: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrency(t *testing.T) {
	f := NewFabric(trinity.FabricConfig{})
	w := trinity.Workload{ID: "x", Kind: "text"}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = f.Route(context.Background(), trinity.Strategy{}, w)
			_ = f.Feedback(context.Background(), trinity.Feedback{WorkloadID: "x", Route: trinity.Route{Model: "blackwell"}, Success: true})
		}()
	}
	wg.Wait()
}
