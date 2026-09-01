package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type FederatedTask struct {
	ID         string
	Capability string
	Payload    map[string]any
	Required   bool
}

type FederatedTaskResult struct {
	ID         string
	Capability string
	Status     string
	Source     string
	DurationMs int64
	Payload    map[string]any
	Error      string
}

// ExecuteParallel fans out independent tasks to the federation and aggregates
// their results without hiding partial failures. Required failures fail the
// aggregate; optional failures remain visible in the returned result set.
func (f *Federation) ExecuteParallel(ctx context.Context, traceID string, tasks []FederatedTask) ([]FederatedTaskResult, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if len(tasks) == 0 {
		return nil, errors.New("tasks are required")
	}
	results := make([]FederatedTaskResult, len(tasks))
	var wg sync.WaitGroup
	for i, task := range tasks {
		i, task := i, task
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			result := FederatedTaskResult{ID: task.ID, Capability: normalizeCapability(task.Capability), Status: "error"}
			if result.ID == "" {
				result.ID = fmt.Sprintf("task-%d", i)
			}
			childTrace := traceID
			if childTrace == "" {
				childTrace = result.ID
			} else {
				childTrace = childTrace + "/" + result.ID
			}
			payload, err := f.Delegate(ctx, childTrace, task.Capability, task.Payload)
			result.DurationMs = time.Since(start).Milliseconds()
			if err != nil {
				result.Error = err.Error()
				results[i] = result
				return
			}
			result.Status = "ok"
			result.Source = responseSource(payload)
			result.Payload = responsePayload(payload)
			results[i] = result
		}()
	}
	wg.Wait()
	for _, result := range results {
		for _, task := range tasks {
			id := task.ID
			if id == "" {
				continue
			}
			if id == result.ID && result.Status != "ok" && task.Required {
				return results, fmt.Errorf("required federated task %s failed: %s", result.ID, result.Error)
			}
		}
	}
	return results, nil
}
