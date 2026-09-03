package fusion

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Kind identifies a runtime component without coupling the fusion engine to a concrete implementation.
type Kind string

const (
	KindAgent Kind = "agent"
	KindTool  Kind = "tool"
)

type Component struct {
	ID           string
	Nucleus      string
	Kind         Kind
	Capabilities []string
	Execute      func(context.Context, []float64) ([]float64, error)
}

type TraceStep struct {
	Component string        `json:"component"`
	Nucleus   string        `json:"nucleus"`
	Kind      Kind          `json:"kind"`
	Started   time.Time     `json:"started"`
	Duration  time.Duration `json:"duration"`
}

type Result struct {
	Output []float64
	Trace  []TraceStep
}

type Registry struct {
	mu         sync.RWMutex
	components map[string]Component
}

func NewRegistry() *Registry { return &Registry{components: make(map[string]Component)} }

func (r *Registry) Register(c Component) error {
	if r == nil {
		return errors.New("fusion registry is nil")
	}
	c.ID = strings.TrimSpace(c.ID)
	c.Nucleus = strings.TrimSpace(c.Nucleus)
	if c.ID == "" || c.Nucleus == "" || c.Execute == nil {
		return errors.New("component id, nucleus and executable function are required")
	}
	if c.Kind != KindAgent && c.Kind != KindTool {
		return errors.New("component kind must be agent or tool")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.components[c.ID]; exists {
		return fmt.Errorf("FUSION_COMPONENT_DUPLICATE: %s", c.ID)
	}
	c.Capabilities = append([]string(nil), c.Capabilities...)
	sort.Strings(c.Capabilities)
	r.components[c.ID] = c
	return nil
}

func (r *Registry) Get(id string) (Component, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.components[strings.TrimSpace(id)]
	return c, ok
}

// AdjacentPair enforces the SOUL chain: only neighboring nuclei may fuse directly.
func AdjacentPair(left, right string) bool {
	order := []string{"N01", "N02", "N03", "N04", "N05", "N06", "N07"}
	li, ri := -1, -1
	for i, id := range order {
		if id == left { li = i }
		if id == right { ri = i }
	}
	return li >= 0 && ri >= 0 && left != right && (li-ri == 1 || ri-li == 1)
}

func (r *Registry) Fuse(ctx context.Context, componentIDs []string, input []float64) (Result, error) {
	if ctx == nil { return Result{}, errors.New("context is nil") }
	if len(componentIDs) < 2 { return Result{}, errors.New("at least two components are required for fusion") }
	current := append([]float64(nil), input...)
	trace := make([]TraceStep, 0, len(componentIDs))
	var previousNucleus string
	for _, id := range componentIDs {
		component, ok := r.Get(id)
		if !ok { return Result{}, fmt.Errorf("FUSION_COMPONENT_NOT_FOUND: %s", id) }
		if previousNucleus != "" && !AdjacentPair(previousNucleus, component.Nucleus) {
			return Result{}, fmt.Errorf("FUSION_NON_ADJACENT_NUCLEI: %s->%s", previousNucleus, component.Nucleus)
		}
		start := time.Now()
		output, err := component.Execute(ctx, current)
		if err != nil { return Result{}, fmt.Errorf("FUSION_COMPONENT_FAILED:%s: %w", component.ID, err) }
		trace = append(trace, TraceStep{Component: component.ID, Nucleus: component.Nucleus, Kind: component.Kind, Started: start.UTC(), Duration: time.Since(start)})
		current = output
		previousNucleus = component.Nucleus
	}
	return Result{Output: current, Trace: trace}, nil
}

func (r *Registry) Snapshot() []Component {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Component, 0, len(r.components))
	for _, c := range r.components { out = append(out, c) }
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
