package neural

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

type recordingPeer struct {
	mu       sync.Mutex
	calls    []string
	payloads []map[string]any
}

func (p *recordingPeer) Invoke(ctx context.Context, nucleus, operation string, payload map[string]any, correlation string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.calls = append(p.calls, nucleus+":"+operation+":"+correlation)
	p.payloads = append(p.payloads, payload)
	p.mu.Unlock()
	return map[string]any{"nucleus": nucleus, "operation": operation}, nil
}

func TestNewFabricRequiresPeer(t *testing.T) {
	if _, err := NewFabric(nil); err == nil {
		t.Fatal("expected nil peer to be rejected")
	}
}

func TestAssignRoutesAndPropagatesCorrelation(t *testing.T) {
	peer := &recordingPeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}

	result, err := fabric.Assign(NeuralTask{
		ID:            "task-1",
		Operation:     "neural.forward@1.0.0",
		Payload:       []float64{1, 2, 3},
		Source:        "N07",
		CorrelationID: "corr-1",
	}, "N03")
	if err != nil {
		t.Fatal(err)
	}
	if result.Nucleus != "N03" || result.Error != "" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(peer.calls) != 1 || peer.calls[0] != "N03:neural.forward@1.0.0:corr-1" {
		t.Fatalf("unexpected peer calls: %#v", peer.calls)
	}
	if got, ok := peer.payloads[0]["payload.values"].([]float64); !ok || len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("canonical payload.values missing or incorrect: %#v", peer.payloads[0])
	}
}

func TestAssignRejectsNonFinitePayload(t *testing.T) {
	peer := &recordingPeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []float64{math.NaN(), math.Inf(1)} {
		_, err := fabric.Assign(NeuralTask{ID: "t", Operation: "neural.forward@1.0.0", Payload: []float64{value}, CorrelationID: "c"}, "N01")
		if err == nil {
			t.Fatal("expected non-finite payload to be rejected")
		}
	}
}

func TestParallelHonorsExplicitTargetAndLegacySourceHint(t *testing.T) {
	peer := &recordingPeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	results := fabric.Parallel([]NeuralTask{
		{ID: "a", Operation: "neural.forward@1.0.0", CorrelationID: "ca", Target: "N06"},
		{ID: "b", Operation: "neural.forward@1.0.0", CorrelationID: "cb", Source: "N05"},
	})
	if len(results) != 2 {
		t.Fatalf("expected two results, got %d", len(results))
	}
	if results[0].Nucleus != "N06" || results[1].Nucleus != "N05" {
		t.Fatalf("unexpected routing: %+v", results)
	}
}

func TestParallelEmptyTasksIsDeterministic(t *testing.T) {
	peer := &recordingPeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	if got := fabric.Parallel(nil); got == nil || len(got) != 0 {
		t.Fatalf("expected empty result, got %#v", got)
	}
}

func TestAssignDeadlineIsEnforced(t *testing.T) {
	peer := &deadlinePeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	result, err := fabric.Assign(NeuralTask{
		ID: "deadline", Operation: "neural.forward@1.0.0", CorrelationID: "deadline-c", Deadline: time.Now().Add(5 * time.Millisecond),
	}, "N01")
	if err == nil || result.Error == "" {
		t.Fatalf("expected deadline failure, result=%+v err=%v", result, err)
	}
}

func TestParallelContextPropagatesCancellation(t *testing.T) {
	peer := &deadlinePeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := fabric.ParallelContext(ctx, []NeuralTask{
		{ID: "a", Operation: "neural.forward@1.0.0", CorrelationID: "ca", Target: "N01"},
		{ID: "b", Operation: "neural.forward@1.0.0", CorrelationID: "cb", Target: "N02"},
	})
	for i, result := range results {
		if result.Error != context.Canceled.Error() {
			t.Fatalf("result %d was not cancelled: %+v", i, result)
		}
	}
}

func TestParallelContextPreservesInputOrderUnderBoundedAdmission(t *testing.T) {
	peer := &blockingPeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	const taskCount = maxFederatedParallelism + 8
	tasks := make([]NeuralTask, taskCount)
	for i := range tasks {
		tasks[i] = NeuralTask{ID: string(rune('a' + i%26)), Operation: "neural.forward@1.0.0", CorrelationID: "corr", Target: "N01"}
	}
	results := fabric.ParallelContext(context.Background(), tasks)
	if len(results) != taskCount {
		t.Fatalf("expected %d results, got %d", taskCount, len(results))
	}
	for i, result := range results {
		if result.Nucleus != "N01" || result.Error != "" {
			t.Fatalf("result %d was not successful/in order: %+v", i, result)
		}
	}
}

func TestParallelContextCapsActivePeerCalls(t *testing.T) {
	peer := &concurrencyPeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	const taskCount = maxFederatedParallelism*2 + 7
	tasks := make([]NeuralTask, taskCount)
	for i := range tasks {
		tasks[i] = NeuralTask{ID: "task", Operation: "neural.forward@1.0.0", CorrelationID: "corr", Target: "N01"}
	}
	results := fabric.ParallelContext(context.Background(), tasks)
	if len(results) != taskCount {
		t.Fatalf("expected %d results, got %d", taskCount, len(results))
	}
	if peer.maxActive > maxFederatedParallelism {
		t.Fatalf("peer concurrency exceeded limit: got %d want <= %d", peer.maxActive, maxFederatedParallelism)
	}
}

func TestBroadcastContextCancellationPreservesMembers(t *testing.T) {
	peer := &deadlinePeer{}
	fabric, err := NewFabric(peer)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := fabric.BroadcastContext(ctx, NeuralTask{ID: "b", Operation: "neural.forward@1.0.0", CorrelationID: "bc"})
	if len(results) != len(fabric.Members()) {
		t.Fatalf("expected one result per member, got %d", len(results))
	}
	for _, result := range results {
		if result.Nucleus == "" || result.Error != context.Canceled.Error() {
			t.Fatalf("unexpected cancellation result: %+v", result)
		}
	}
}

type deadlinePeer struct{}

func (*deadlinePeer) Invoke(ctx context.Context, nucleus, operation string, payload map[string]any, correlation string) (map[string]any, error) {
	<-ctx.Done()
	return nil, errors.New(ctx.Err().Error())
}

type blockingPeer struct{}

func (*blockingPeer) Invoke(ctx context.Context, nucleus, operation string, payload map[string]any, correlation string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]any{"nucleus": nucleus, "operation": operation}, nil
}

type concurrencyPeer struct {
	mu        sync.Mutex
	active    int
	maxActive int
}

func (p *concurrencyPeer) Invoke(ctx context.Context, nucleus, operation string, payload map[string]any, correlation string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.active++
	if p.active > p.maxActive {
		p.maxActive = p.active
	}
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
	}()
	select {
	case <-time.After(2 * time.Millisecond):
		return map[string]any{"nucleus": nucleus, "operation": operation}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
