package supergpu

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRuntime(t *testing.T) {
	r := New(nil)
	d := r.Discover()
	if len(d) == 0 {
		t.Fatal("no compute device")
	}
	if err := r.Reserve(d[0].ID, "test"); err != nil {
		t.Fatal(err)
	}
	v, err := r.Execute(context.Background(), d[0], "square", []float64{2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if v[0] != 4 || v[1] != 9 {
		t.Fatal("compute result incorrect")
	}
	if _, err = r.Batch(context.Background(), d[0], "identity", [][]float64{{1}, {2}}); err != nil {
		t.Fatal(err)
	}
	if err = r.Release(d[0].ID, "test"); err != nil {
		t.Fatal(err)
	}
	if r.Health()["status"] != "ready" {
		t.Fatal("runtime unhealthy")
	}
	if err = r.Shutdown(); err != nil {
		t.Fatal(err)
	}
}

type parallelProbeBackend struct {
	active int32
	peak   int32
}

func (b *parallelProbeBackend) Supports(Device) bool         { return true }
func (b *parallelProbeBackend) Capabilities(Device) []string { return []string{"probe"} }
func (b *parallelProbeBackend) ConcurrentSafe() bool         { return true }
func (b *parallelProbeBackend) Execute(ctx context.Context, _ Device, _ string, in []float64) ([]float64, error) {
	now := atomic.AddInt32(&b.active, 1)
	for {
		peak := atomic.LoadInt32(&b.peak)
		if now <= peak || atomic.CompareAndSwapInt32(&b.peak, peak, now) {
			break
		}
	}
	defer atomic.AddInt32(&b.active, -1)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(20 * time.Millisecond):
	}
	return append([]float64(nil), in...), nil
}
func TestBatchParallelPreservesOrderAndUsesWorkers(t *testing.T) {
	backend := &parallelProbeBackend{}
	r := New(backend)
	device := Device{ID: "probe-0", Vendor: "test", Name: "parallel-probe", Available: true, Backend: "probe"}
	inputs := make([][]float64, 8)
	for i := range inputs {
		inputs[i] = []float64{float64(i)}
	}
	results, err := r.BatchParallel(context.Background(), device, "probe", inputs, 4)
	if err != nil {
		t.Fatal(err)
	}
	if peak := atomic.LoadInt32(&backend.peak); peak < 2 {
		t.Fatalf("expected concurrent workers, peak=%d", peak)
	}
	if len(results) != len(inputs) {
		t.Fatalf("result length mismatch: got %d want %d", len(results), len(inputs))
	}
	for i, result := range results {
		if len(result) != 1 || result[0] != float64(i) {
			t.Fatalf("result order mismatch at %d: %#v", i, result)
		}
	}
}
func TestBatchParallelHonorsCancellation(t *testing.T) {
	backend := &parallelProbeBackend{}
	r := New(backend)
	device := Device{ID: "probe-0", Vendor: "test", Name: "parallel-probe", Available: true, Backend: "probe"}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	inputs := make([][]float64, 20)
	for i := range inputs {
		inputs[i] = []float64{float64(i)}
	}
	if _, err := r.BatchParallel(ctx, device, "probe", inputs, 4); err == nil {
		t.Fatal("expected cancellation error")
	}
}


func TestRuntimeExecutionReporterReceivesCorrelationAndLifecycle(t *testing.T) {
	r := New(nil)
	r.Discover()

	var events []ExecutionEvent
	var mu sync.Mutex
	r.SetExecutionReporter(ReporterFunc(func(_ context.Context, event ExecutionEvent) error {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
		return nil
	}))

	ctx := WithCorrelationID(context.Background(), "corr-clareira-1")
	device, err := r.Select("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Execute(ctx, device, "square", []float64{2, 3}); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 2 {
		t.Fatalf("expected started+completed reports, got %d", len(events))
	}
	if events[0].Phase != "started" || events[1].Phase != "completed" {
		t.Fatalf("unexpected lifecycle: %#v", events)
	}
	for _, event := range events {
		if event.CorrelationID != "corr-clareira-1" {
			t.Fatalf("correlation not preserved: %#v", event)
		}
	}
	if events[1].OutputSize != 2 {
		t.Fatalf("expected output size 2, got %d", events[1].OutputSize)
	}
}
