package neural

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

// RemotePeer is the transport-neutral boundary used to expose the N07 neural
// runtime to every nucleus without copying the neural implementation.
type RemotePeer interface {
	Invoke(ctx context.Context, nucleus, operation string, payload map[string]any, correlation string) (map[string]any, error)
}

type NeuralTask struct {
	ID string `json:"id"`
	Operation string `json:"operation"`
	Payload []float64 `json:"payload,omitempty"`
	Priority int `json:"priority"`
	Deadline time.Time `json:"deadline,omitempty"`
	Source string `json:"source"`
	CorrelationID string `json:"correlationId"`
}

type FederatedResult struct {
	Nucleus string `json:"nucleus"`
	Result map[string]any `json:"result,omitempty"`
	Error string `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

type Fabric struct {
	mu sync.RWMutex
	peer RemotePeer
	members map[string]struct{}
}

func NewFabric(peer RemotePeer) (*Fabric, error) {
	if peer == nil { return nil, errors.New("neural federation peer is required") }
	return &Fabric{peer: peer, members: map[string]struct{}{"N01":{},"N02":{},"N03":{},"N04":{},"N05":{},"N06":{}}}, nil
}

func (f *Fabric) Members() []string {
	f.mu.RLock(); defer f.mu.RUnlock()
	out := make([]string, 0, len(f.members))
	for m := range f.members { out = append(out, m) }
	sort.Strings(out)
	return out
}

func (f *Fabric) Assign(task NeuralTask, nucleus string) (FederatedResult, error) {
	if task.ID == "" || task.Operation == "" || task.CorrelationID == "" { return FederatedResult{}, errors.New("task id, operation and correlation are required") }
	f.mu.RLock(); _, ok := f.members[nucleus]; peer := f.peer; f.mu.RUnlock()
	if !ok { return FederatedResult{}, errors.New("unknown neural nucleus: " + nucleus) }
	started := time.Now()
	ctx := context.Background()
	var cancel context.CancelFunc
	if task.Deadline.IsZero() { ctx, cancel = context.WithTimeout(ctx, 15*time.Second) } else { ctx, cancel = context.WithDeadline(ctx, task.Deadline) }
	defer cancel()
	payload := map[string]any{"taskId":task.ID,"operation":task.Operation,"payload":task.Payload,"priority":task.Priority,"source":task.Source}
	result, err := peer.Invoke(ctx, nucleus, task.Operation, payload, task.CorrelationID)
	fr := FederatedResult{Nucleus:nucleus, Result:result, Duration:time.Since(started)}
	if err != nil { fr.Error = err.Error(); return fr, err }
	return fr, nil
}

func (f *Fabric) Broadcast(task NeuralTask) []FederatedResult {
	members := f.Members(); results := make([]FederatedResult, len(members)); var wg sync.WaitGroup; wg.Add(len(members))
	for i, nucleus := range members { go func(i int, nucleus string) { defer wg.Done(); results[i], _ = f.Assign(task, nucleus) }(i, nucleus) }
	wg.Wait()
	return results
}

func (f *Fabric) Parallel(tasks []NeuralTask) []FederatedResult {
	results := make([]FederatedResult, len(tasks)); var wg sync.WaitGroup; wg.Add(len(tasks))
	for i, task := range tasks { go func(i int, task NeuralTask) { defer wg.Done(); target := task.Source; if target == "" { members := f.Members(); if len(members) == 0 { results[i] = FederatedResult{Error:"no neural members available"}; return }; target = members[i%len(members)] }; results[i], _ = f.Assign(task, target) }(i, task) }
	wg.Wait()
	return results
}
