package octacore

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/supergpu"
)

func testScheduler(t *testing.T) *OctaCoreScheduler {
	t.Helper()
	cfg := DefaultSchedulerConfig()
	cfg.MaxInflight = 8
	cfg.TokenCapacity = 8
	cfg.TokenRefillPerSec = 1000
	s, err := NewScheduler(cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func testJob(id string, priority int, group, barrier *string) OctaCoreJob {
	return OctaCoreJob{
		JobID: id, CorrelationID: "corr-" + id, Kind: KindCustom, Source: G7, Target: "G7",
		BackendPrefs: []Backend{BackendInProcess}, ParallelGroup: group, Barrier: barrier,
		Payload: map[string]any{"payload": id}, Priority: priority, TTLMS: 5000,
	}
}

func TestOctaCoreInventoryHasEightSlots(t *testing.T) {
	s := testScheduler(t)
	slots := s.Inventory()
	if len(slots) != 8 {
		t.Fatalf("expected 8 slots, got %d", len(slots))
	}
	if slots[0].Slot != G0 || slots[7].Slot != G7 {
		t.Fatalf("unexpected slot boundaries: %#v", slots)
	}
}

func TestOctaCoreParallelGroupRunsConcurrently(t *testing.T) {
	s := testScheduler(t)
	var mu sync.Mutex
	starts := make(map[string]time.Time)
	ends := make(map[string]time.Time)
	if err := s.RegisterLocalKernel(G7, func(ctx context.Context, job OctaCoreJob) (map[string]any, error) {
		start := time.Now()
		mu.Lock()
		starts[job.JobID] = start
		mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(80 * time.Millisecond):
		}
		mu.Lock()
		ends[job.JobID] = time.Now()
		mu.Unlock()
		return map[string]any{"job_id": job.JobID}, nil
	}); err != nil {
		t.Fatal(err)
	}
	group := "pre"
	started := time.Now()
	results := s.ExecutePlan(context.Background(), []OctaCoreJob{
		testJob("j1", 10, &group, nil),
		testJob("j2", 9, &group, nil),
	})
	elapsed := time.Since(started)
	for _, result := range results {
		if !result.OK {
			t.Fatalf("job failed: %#v", result.Error)
		}
	}
	if elapsed >= 140*time.Millisecond {
		t.Fatalf("parallel group appears serial: elapsed=%v", elapsed)
	}
	mu.Lock()
	defer mu.Unlock()
	if starts["j1"].IsZero() || starts["j2"].IsZero() {
		t.Fatal("missing start timestamps")
	}
	if starts["j1"].After(ends["j2"]) || starts["j2"].After(ends["j1"]) {
		t.Fatal("jobs did not overlap")
	}
}

func TestOctaCoreBarrierWaitsForPriorGroup(t *testing.T) {
	s := testScheduler(t)
	var mu sync.Mutex
	ends := make(map[string]time.Time)
	var consumerStart time.Time
	if err := s.RegisterLocalKernel(G7, func(ctx context.Context, job OctaCoreJob) (map[string]any, error) {
		if job.JobID == "consumer" {
			mu.Lock()
			consumerStart = time.Now()
			mu.Unlock()
			return map[string]any{"joined": true}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(60 * time.Millisecond):
		}
		mu.Lock()
		ends[job.JobID] = time.Now()
		mu.Unlock()
		return map[string]any{"done": true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	group := "pre"
	barrier := "pre"
	consumer := testJob("consumer", 8, nil, &barrier)
	consumer.Payload["barrier_role"] = "consumer"
	jobs := []OctaCoreJob{
		testJob("p1", 10, &group, &barrier),
		testJob("p2", 9, &group, &barrier),
		consumer,
	}
	results := s.ExecutePlan(context.Background(), jobs)
	for _, result := range results {
		if !result.OK {
			t.Fatalf("job failed: %#v", result.Error)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	latestProducer := ends["p1"]
	if ends["p2"].After(latestProducer) {
		latestProducer = ends["p2"]
	}
	if consumerStart.Before(latestProducer) {
		t.Fatalf("barrier released before producer completion: consumer=%v producer=%v", consumerStart, latestProducer)
	}
}

func TestOctaCoreThrottleReducesInflight(t *testing.T) {
	s := testScheduler(t)
	if err := s.SetThrottle(3); err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterLocalKernel(G7, func(ctx context.Context, job OctaCoreJob) (map[string]any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return map[string]any{"ok": true}, nil
		}
	}); err != nil {
		t.Fatal(err)
	}
	group := "throttle"
	started := time.Now()
	results := s.ExecutePlan(context.Background(), []OctaCoreJob{testJob("t1", 10, &group, nil), testJob("t2", 9, &group, nil)})
	elapsed := time.Since(started)
	for _, result := range results {
		if !result.OK {
			t.Fatalf("job failed: %#v", result.Error)
		}
	}
	if elapsed < 90*time.Millisecond {
		t.Fatalf("throttle did not reduce concurrency: elapsed=%v", elapsed)
	}
}

func TestOctaCoreCircuitBreakerOpensAfterFailures(t *testing.T) {
	cfg := DefaultSchedulerConfig()
	cfg.FailureThreshold = 2
	cfg.TokenRefillPerSec = 1000
	s, err := NewScheduler(cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterLocalKernel(G7, func(context.Context, OctaCoreJob) (map[string]any, error) { return nil, errors.New("kernel_failure") }); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		result := s.Execute(context.Background(), testJob("f"+string(rune('1'+i)), 1, nil, nil))
		if result.OK {
			t.Fatal("expected execution failure")
		}
	}
	third := s.Execute(context.Background(), testJob("f3", 1, nil, nil))
	if third.OK {
		t.Fatal("expected open circuit to reject third execution")
	}
	if third.Error == nil || third.Error.Code != "THROTTLED" {
		t.Fatalf("unexpected third error: %#v", third.Error)
	}
	health := s.Health()
	for _, slot := range health.Slots {
		if slot.Slot == G7 && slot.Circuit != string(CircuitOpen) {
			t.Fatalf("expected G7 open circuit, got %s", slot.Circuit)
		}
	}
}

func TestOctaCoreFinalSuperGPUConnectionIsExplicit(t *testing.T) {
	s := testScheduler(t)
	if s.SuperGPUConnected() {
		t.Fatal("test scheduler must start without a connected SuperGPU runtime")
	}
	runtime := supergpu.New(nil)
	if err := s.AttachSuperGPU(runtime); err != nil {
		t.Fatal(err)
	}
	if !s.SuperGPUConnected() {
		t.Fatal("expected canonical SuperGPU runtime to be connected")
	}
	if !s.Health().SuperGPUConnected {
		t.Fatal("health must expose SuperGPU connectivity")
	}
}
