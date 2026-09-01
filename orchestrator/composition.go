package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type CapabilityStep struct {
	ID           string
	Capability   string
	Payload      map[string]any
	DependsOn    []string
	Parallel     bool
	Required     bool
}

type CapabilityPlan struct {
	ID         string
	Name       string
	Version    string
	Steps      []CapabilityStep
	Components []string
}

type CapabilityResult struct {
	ID         string
	StepID     string
	Capability string
	Source     string
	Status     string
	DurationMs int64
	Payload    map[string]any
	Error      string
}

func NewCapabilityPlan(id, name, version string, steps []CapabilityStep) (CapabilityPlan, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if id == "" || name == "" || version == "" {
		return CapabilityPlan{}, errors.New("composition id, name and version are required")
	}
	if len(steps) == 0 {
		return CapabilityPlan{}, errors.New("composition requires at least one step")
	}
	seen := map[string]struct{}{}
	components := make([]string, 0, len(steps))
	for _, step := range steps {
		if strings.TrimSpace(step.ID) == "" || strings.TrimSpace(step.Capability) == "" {
			return CapabilityPlan{}, errors.New("composition step id and capability are required")
		}
		if _, ok := seen[step.ID]; ok {
			return CapabilityPlan{}, fmt.Errorf("duplicate composition step: %s", step.ID)
		}
		seen[step.ID] = struct{}{}
		components = append(components, normalizeCapability(step.Capability))
	}
	for _, step := range steps {
		for _, dependency := range step.DependsOn {
			if _, ok := seen[dependency]; !ok {
				return CapabilityPlan{}, fmt.Errorf("step %s depends on unknown step %s", step.ID, dependency)
			}
			if dependency == step.ID {
				return CapabilityPlan{}, fmt.Errorf("step %s cannot depend on itself", step.ID)
			}
		}
	}
	return CapabilityPlan{ID: id, Name: name, Version: version, Steps: append([]CapabilityStep(nil), steps...), Components: uniqueStrings(components)}, nil
}

func (p CapabilityPlan) Validate() error {
	_, err := NewCapabilityPlan(p.ID, p.Name, p.Version, p.Steps)
	return err
}

func (p CapabilityPlan) Execute(ctx context.Context, federation *Federation) ([]CapabilityResult, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if federation == nil {
		return nil, errors.New("federation is required")
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}

	completed := map[string]bool{}
	results := make([]CapabilityResult, 0, len(p.Steps))
	remaining := append([]CapabilityStep(nil), p.Steps...)
	for len(remaining) > 0 {
		ready := make([]CapabilityStep, 0, len(remaining))
		blocked := make([]CapabilityStep, 0, len(remaining))
		for _, step := range remaining {
			ok := true
			for _, dependency := range step.DependsOn {
				if !completed[dependency] {
					ok = false
					break
				}
			}
			if ok {
				ready = append(ready, step)
			} else {
				blocked = append(blocked, step)
			}
		}
		if len(ready) == 0 {
			return results, errors.New("composition contains a dependency cycle or unresolved dependency")
		}

		parallel := make([]CapabilityStep, 0, len(ready))
		serial := make([]CapabilityStep, 0, len(ready))
		for _, step := range ready {
			if step.Parallel {
				parallel = append(parallel, step)
			} else {
				serial = append(serial, step)
			}
		}

		serialResults, err := executeSteps(ctx, federation, serial)
		results = append(results, serialResults...)
		if err != nil {
			return results, err
		}
		for _, result := range serialResults {
			if result.Status == "ok" {
				completed[result.StepID] = true
			} else if isRequired(p.Steps, result.StepID) {
				return results, fmt.Errorf("required step %s failed: %s", result.StepID, result.Error)
			}
		}

		parallelResults := make([]CapabilityResult, 0, len(parallel))
		if len(parallel) > 0 {
			parallelResults, err = executeStepsParallel(ctx, federation, parallel)
			results = append(results, parallelResults...)
			if err != nil {
				return results, err
			}
			for _, result := range parallelResults {
				if result.Status == "ok" {
					completed[result.StepID] = true
				} else if isRequired(p.Steps, result.StepID) {
					return results, fmt.Errorf("required step %s failed: %s", result.StepID, result.Error)
				}
			}
		}
		remaining = blocked
	}
	sort.Slice(results, func(i, j int) bool { return results[i].StepID < results[j].StepID })
	return results, nil
}

func executeSteps(ctx context.Context, federation *Federation, steps []CapabilityStep) ([]CapabilityResult, error) {
	results := make([]CapabilityResult, 0, len(steps))
	for _, step := range steps {
		result := executeStep(ctx, federation, step)
		results = append(results, result)
		if result.Status != "ok" && step.Required {
			return results, fmt.Errorf("required step %s failed: %s", step.ID, result.Error)
		}
	}
	return results, nil
}

func executeStepsParallel(ctx context.Context, federation *Federation, steps []CapabilityStep) ([]CapabilityResult, error) {
	results := make([]CapabilityResult, len(steps))
	var wg sync.WaitGroup
	for i, step := range steps {
		i, step := i, step
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = executeStep(ctx, federation, step)
		}()
	}
	wg.Wait()
	for _, result := range results {
		if result.Status != "ok" && isRequired(steps, result.StepID) {
			return results, fmt.Errorf("required parallel step %s failed: %s", result.StepID, result.Error)
		}
	}
	return results, nil
}

func executeStep(ctx context.Context, federation *Federation, step CapabilityStep) CapabilityResult {
	start := time.Now()
	result := CapabilityResult{ID: step.ID, StepID: step.ID, Capability: normalizeCapability(step.Capability), Status: "error"}
	response, err := federation.Delegate(ctx, protocol.NewTraceID(), step.Capability, step.Payload)
	result.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Status = "ok"
	result.Source = responseSource(response)
	result.Payload = responsePayload(response)
	return result
}

func responseSource(response map[string]any) string {
	if value, ok := response["source"].(string); ok {
		return value
	}
	return ""
}

func responsePayload(response map[string]any) map[string]any {
	if value, ok := response["payload"].(map[string]any); ok {
		return value
	}
	return response
}

func isRequired(steps []CapabilityStep, id string) bool {
	for _, step := range steps {
		if step.ID == id {
			return step.Required
		}
	}
	return false
}
