package neural

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

const (
	defaultNeuralTaskTimeout = 15 * time.Second
	maxFederatedParallelism  = 32
)

// RemotePeer is the transport-neutral boundary used to expose the N07 neural
// runtime to every nucleus without copying the neural implementation.
type RemotePeer interface {
	Invoke(ctx context.Context, nucleus, operation string, payload map[string]any, correlation string) (map[string]any, error)
}

type NeuralTask struct {
	ID            string    `json:"id"`
	Operation     string    `json:"operation"`
	Payload       []float64 `json:"payload,omitempty"`
	Priority      int       `json:"priority"`
	Deadline      time.Time `json:"deadline,omitempty"`
	Source        string    `json:"source"`
	Target        string    `json:"target,omitempty"`
	CorrelationID string    `json:"correlationId"`
}

type FederatedResult struct {
	Nucleus  string         `json:"nucleus"`
	Result   map[string]any `json:"result,omitempty"`
	Error    string         `json:"error,omitempty"`
	Duration time.Duration  `json:"duration"`
}

type Fabric struct {
	mu      sync.RWMutex
	peer    RemotePeer
	members map[string]struct{}
}

func NewFabric(peer RemotePeer) (*Fabric, error) {
	if peer == nil {
		return nil, errors.New("neural federation peer is required")
	}
	return &Fabric{peer: peer, members: map[string]struct{}{
		"N01": {}, "N02": {}, "N03": {}, "N04": {}, "N05": {}, "N06": {},
	}}, nil
}

func (f *Fabric) Members() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]string, 0, len(f.members))
	for m := range f.members {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

func (f *Fabric) Assign(task NeuralTask, nucleus string) (FederatedResult, error) {
	return f.assignContext(context.Background(), task, nucleus)
}

func (f *Fabric) assignContext(parent context.Context, task NeuralTask, nucleus string) (FederatedResult, error) {
	if parent == nil {
		return FederatedResult{}, errors.New("neural parent context is required")
	}
	if task.ID == "" || task.Operation == "" || task.CorrelationID == "" {
		return FederatedResult{}, errors.New("task id, operation and correlation are required")
	}
	if nucleus == "" {
		return FederatedResult{}, errors.New("neural nucleus is required")
	}
	for _, value := range task.Payload {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return FederatedResult{}, errors.New("neural payload must contain only finite numbers")
		}
	}

	f.mu.RLock()
	_, ok := f.members[nucleus]
	peer := f.peer
	f.mu.RUnlock()
	if !ok {
		return FederatedResult{}, errors.New("unknown neural nucleus: " + nucleus)
	}

	started := time.Now()
	ctx := parent
	var cancel context.CancelFunc
	if task.Deadline.IsZero() {
		ctx, cancel = context.WithTimeout(ctx, defaultNeuralTaskTimeout)
	} else {
		ctx, cancel = context.WithDeadline(ctx, task.Deadline)
	}
	defer cancel()

	payload := map[string]any{
		"taskId":         task.ID,
		"operation":      task.Operation,
		"payload":        task.Payload,  // legacy compatibility
		"payload.values": task.Payload,  // canonical neural Mesh vector field
		"values":         task.Payload,  // compact adapter compatibility
		"priority":       task.Priority,
		"source":         task.Source,
		"target":         nucleus,
		"correlationId":  task.CorrelationID,
	}
	result, err := peer.Invoke(ctx, nucleus, task.Operation, payload, task.CorrelationID)
	fr := FederatedResult{Nucleus: nucleus, Result: result, Duration: time.Since(started)}
	if err != nil {
		fr.Error = err.Error()
		return fr, err
	}
	return fr, nil
}

func (f *Fabric) Broadcast(task NeuralTask) []FederatedResult {
	return f.BroadcastContext(context.Background(), task)
}

// BroadcastContext fans a task out to all registered nuclei while respecting
// caller cancellation and the global federation concurrency bound.
func (f *Fabric) BroadcastContext(ctx context.Context, task NeuralTask) []FederatedResult {
	members := f.Members()
	if len(members) == 0 {
		return []FederatedResult{}
	}
	return f.parallelMembers(ctx, task, members)
}

func (f *Fabric) Parallel(tasks []NeuralTask) []FederatedResult {
	return f.ParallelContext(context.Background(), tasks)
}

// ParallelContext executes independent neural tasks with a fixed worker pool.
// Admission is therefore bounded by maxFederatedParallelism rather than merely
// limiting active peer calls while creating one goroutine per submitted task.
// Result ordering remains identical to input ordering.
func (f *Fabric) ParallelContext(ctx context.Context, tasks []NeuralTask) []FederatedResult {
	results := make([]FederatedResult, len(tasks))
	if len(tasks) == 0 {
		return results
	}
	if ctx == nil {
		for i := range results {
			results[i].Error = "neural parent context is required"
		}
		return results
	}

	members := f.Members()
	if len(members) == 0 {
		for i := range results {
			results[i].Error = "no neural members available"
		}
		return results
	}

	workerCount := len(tasks)
	if workerCount > maxFederatedParallelism {
		workerCount = maxFederatedParallelism
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case index, ok := <-jobs:
					if !ok {
						return
					}
					task := tasks[index]
					target := task.Target
					// Source was historically used as the routing hint. Preserve
					// that compatibility while Target remains canonical.
					if target == "" {
						target = task.Source
					}
					if target == "" {
						target = members[index%len(members)]
					}
					results[index], _ = f.assignContext(ctx, task, target)
				}
			}
		}()
	}

	dispatch:
	for index := range tasks {
		select {
		case jobs <- index:
		case <-ctx.Done():
			break dispatch
		}
	}
	close(jobs)
	wg.Wait()

	if err := ctx.Err(); err != nil {
		for i := range results {
			if results[i].Nucleus == "" && results[i].Error == "" {
				results[i].Error = err.Error()
			}
		}
	}
	return results
}

func (f *Fabric) parallelMembers(ctx context.Context, task NeuralTask, members []string) []FederatedResult {
	results := make([]FederatedResult, len(members))
	if ctx == nil {
		for i := range results {
			results[i].Nucleus = members[i]
			results[i].Error = "neural parent context is required"
		}
		return results
	}

	var wg sync.WaitGroup
	workerCount := len(members)
	if workerCount > maxFederatedParallelism {
		workerCount = maxFederatedParallelism
	}
	jobs := make(chan int)
	wg.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case index, ok := <-jobs:
					if !ok {
						return
					}
					results[index], _ = f.assignContext(ctx, task, members[index])
				}
			}
		}()
	}

	dispatch:
	for index := range members {
		select {
		case jobs <- index:
		case <-ctx.Done():
			break dispatch
		}
	}
	close(jobs)
	wg.Wait()

	if err := ctx.Err(); err != nil {
		for i := range results {
			if results[i].Nucleus == "" {
				results[i].Nucleus = members[i]
			}
			if results[i].Error == "" && results[i].Result == nil {
				results[i].Error = err.Error()
			}
		}
	}
	return results
}
