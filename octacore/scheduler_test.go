package octacore

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type slowBackend struct{}

func (slowBackend) ConcurrentSafe() bool { return true }

func (slowBackend) Execute(ctx context.Context, _ supergpu.Device, _ string, input []float64) ([]float64, error) {
	timer := time.NewTimer(75 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
	}
	out := append([]float64(nil), input...)
	return out, nil
}

func newTestScheduler(t *testing.T) *Scheduler {
	t.Helper()
	runtime := supergpu.New(slowBackend{})
	runtime.Discover()
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	return newScheduler(Config{
		MaxInflight:          8,
		TokenCapacity:        8,
		TokenRefillPerSecond: 1000,
		FailureThreshold:     2,
		CircuitCooldown:      100 * time.Millisecond,
	}, runtime, peers, nil)
}

func job(id string, source, target SlotID, group, barrier *string) Job {
	return Job{
		JobID:         id,
		CorrelationID: "corr-" + id,
		Kind:          KindCustom,
		Source:        source,
		Target:        string(target),
		BackendPrefs:  []Backend{BackendInProcess},
		ParallelGroup: group,
		Barrier:       barrier,
		Payload:       map[string]any{"operation": "identity", "values": []float64{1, 2, 3}},
		Priority:      50,
		TTLMS:         5000,
	}
}

func TestOctacoreInventoryHasEightSlots(t *testing.T) {
	s := newTestScheduler(t)
	slots := s.inventory()
	if len(slots) != 8 {
		t.Fatalf("expected 8 slots, got %d", len(slots))
	}
	for i, slot := range slots {
		expected := SlotID("G" + string(rune('0'+i)))
		if slot.Slot != expected {
			t.Fatalf("slot %d expected %s, got %s", i, expected, slot.Slot)
		}
	}
}

func TestOctacoreParallelGroupRunsConcurrently(t *testing.T) {
	s := newTestScheduler(t)
	group := "pre"
	start := time.Now()
	results := s.executePlan(context.Background(), []Job{
		job("a", G7, G7, &group, nil),
		job("b", G7, G7, &group, nil),
	})
	elapsed := time.Since(start)
	for _, result := range results {
		if !result.OK {
			t.Fatalf("parallel job failed: %#v", result.Error)
		}
	}
	if elapsed >= 140*time.Millisecond {
		t.Fatalf("parallel_group was not concurrent: elapsed=%s", elapsed)
	}
}

func TestOctacoreBarrierWaitsForIndependentGroup(t *testing.T) {
	s := newTestScheduler(t)
	producerGroup := "pre"
	barrier := "pre"
	consumerGroup := "post"

	start := time.Now()
	results := s.executePlan(context.Background(), []Job{
		job("p1", G7, G7, &producerGroup, nil),
		job("p2", G7, G7, &producerGroup, nil),
		{
			JobID:         "consumer",
			CorrelationID: "corr-consumer",
			Kind:          KindCustom,
			Source:        G7,
			Target:        "G7",
			BackendPrefs:  []Backend{BackendInProcess},
			ParallelGroup: &consumerGroup,
			Barrier:       &barrier,
			Payload:       map[string]any{"operation": "identity", "values": []float64{1}},
			Priority:      40,
			TTLMS:         5000,
		},
	})
	elapsed := time.Since(start)
	for _, result := range results {
		if !result.OK {
			t.Fatalf("barrier plan failed: %#v", result.Error)
		}
	}
	if elapsed < 70*time.Millisecond {
		t.Fatalf("barrier consumer did not wait for producers: elapsed=%s", elapsed)
	}
}

func TestOctacoreThrottleReducesInflight(t *testing.T) {
	s := newTestScheduler(t)
	if err := s.setThrottle(3); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	group := "throttle"
	results := s.executePlan(context.Background(), []Job{
		job("t1", G7, G7, &group, nil),
		job("t2", G7, G7, &group, nil),
	})
	elapsed := time.Since(start)
	for _, result := range results {
		if !result.OK {
			t.Fatalf("throttle job failed: %#v", result.Error)
		}
	}
	if elapsed < 140*time.Millisecond {
		t.Fatalf("throttle did not serialize execution: elapsed=%s", elapsed)
	}
}

func TestOctacoreG0RequiresSARAAndCannotUseInProcessGPU(t *testing.T) {
	s := newTestScheduler(t)
	result := s.execute(context.Background(), Job{
		JobID:         "g0",
		CorrelationID: "corr-g0",
		Kind:          KindSARACycle,
		Source:        G6,
		Target:        "G0",
		BackendPrefs:  []Backend{BackendInProcess},
		Payload:       map[string]any{"input": "test"},
		Priority:      90,
		TTLMS:         1000,
	})
	if result.OK {
		t.Fatal("G0 execution must not use local N07 SuperGPU")
	}
	if result.Error == nil || result.Error.Code != "SARA_UNAVAILABLE" {
		t.Fatalf("expected SARA_UNAVAILABLE, got %#v", result.Error)
	}
}

func TestOctacoreWebGPUIsExplicitlyUnavailable(t *testing.T) {
	s := newTestScheduler(t)
	group := "gpu"
	result := s.execute(context.Background(), Job{
		JobID:         "webgpu",
		CorrelationID: "corr-webgpu",
		Kind:          KindCustom,
		Source:        G7,
		Target:        "G7",
		BackendPrefs:  []Backend{BackendWebGPU},
		ParallelGroup: &group,
		Payload:       map[string]any{"operation": "identity", "values": []float64{1}},
		Priority:      10,
		TTLMS:         1000,
	})
	if result.OK {
		t.Fatal("WebGPU cannot be marked available without a real backend")
	}
	if result.Error == nil || result.Error.Code != "BACKEND_UNAVAILABLE" {
		t.Fatalf("unexpected WebGPU error: %#v", result.Error)
	}
}

type halfOpenBackend struct {
	mu    sync.Mutex
	calls int
	release chan struct{}
}

func (b *halfOpenBackend) ConcurrentSafe() bool { return true }

func (b *halfOpenBackend) Execute(ctx context.Context, _ supergpu.Device, _ string, input []float64) ([]float64, error) {
	b.mu.Lock()
	b.calls++
	call := b.calls
	b.mu.Unlock()
	if call <= 2 {
		return nil, errors.New("synthetic backend failure for circuit test")
	}
	select {
	case <-b.release:
		return append([]float64(nil), input...), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestOctacoreHalfOpenAllowsOnlyOneProbe(t *testing.T) {
	backend := &halfOpenBackend{release: make(chan struct{})}
	runtime := supergpu.New(backend)
	runtime.Discover()
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	s := newScheduler(Config{
		MaxInflight:          2,
		TokenCapacity:        2,
		TokenRefillPerSecond: 1000,
		FailureThreshold:     2,
		CircuitCooldown:      20 * time.Millisecond,
	}, runtime, peers, nil)

	makeJob := func(id string) Job {
		return Job{
			JobID:         id, CorrelationID: "corr-" + id, Kind: KindDispatch,
			Source: G7, Target: "G7", BackendPrefs: []Backend{BackendInProcess},
			Payload: map[string]any{"operation": "identity", "values": []float64{1}},
			Priority: 50, TTLMS: 1000,
		}
	}

	first := s.execute(context.Background(), makeJob("fail-1"))
	second := s.execute(context.Background(), makeJob("fail-2"))
	if first.OK || second.OK {
		t.Fatalf("expected initial failures to open circuit: first=%+v second=%+v", first, second)
	}

	time.Sleep(30 * time.Millisecond)
	resultCh := make(chan Result, 1)
	go func() { resultCh <- s.execute(context.Background(), makeJob("probe")) }()
	time.Sleep(5 * time.Millisecond)
	contender := s.execute(context.Background(), makeJob("contender"))
	if contender.OK || contender.Error == nil || contender.Error.Code != "ADMISSION_THROTTLED" {
		t.Fatalf("second half-open admission was not rejected: %+v", contender)
	}
	close(backend.release)
	probe := <-resultCh
	if !probe.OK {
		t.Fatalf("half-open probe failed: %+v", probe)
	}
}
