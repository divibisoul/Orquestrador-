package neuralfabric

import (
	"context"
	"errors"
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

func TestRuntimeCheckpointLifecycle(t *testing.T) {
	ctx := context.Background()
	f := NewRuntime()
	f.RecordFeedback(Experience{WorkloadID: "checkpoint", Reward: 0.8})
	before := f.ExperienceCount()
	if before != 1 {
		t.Fatalf("experience count=%d want 1", before)
	}
	if err := f.Save(ctx); err != nil {
		t.Fatal(err)
	}
	f.RecordFeedback(Experience{WorkloadID: "extra", Reward: -0.2})
	if f.ExperienceCount() != 2 {
		t.Fatalf("experience count=%d want 2", f.ExperienceCount())
	}
	if err := f.Load(ctx); err != nil {
		t.Fatal(err)
	}
	if f.ExperienceCount() != before {
		t.Fatalf("restored experience count=%d want %d", f.ExperienceCount(), before)
	}
}

func TestRuntimeUpdateChangesWeightsSafely(t *testing.T) {
	f := NewRuntime()
	f.RecordFeedback(Experience{WorkloadID: "rewarded", Reward: 1})
	if err := f.Update(context.Background()); err != nil {
		t.Fatal(err)
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	for i, w := range f.weights {
		if w < 0 || w > 1 {
			t.Fatalf("weight[%d]=%f outside [0,1]", i, w)
		}
	}
}

func TestRuntimeRejectsNilContext(t *testing.T) {
	f := NewRuntime()
	if _, err := f.EncodeTask(nil, "task"); !errors.Is(err, ErrNilContext) {
		t.Fatalf("EncodeTask error=%v", err)
	}
}
