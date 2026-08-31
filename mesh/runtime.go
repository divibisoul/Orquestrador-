package mesh

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Node struct {
	ID            string
	Status        string
	Capabilities  []string
	Transports    []string
	CPU, GPU, NPU bool
	RAMMB, VRAMMB int64
	LatencyMS     float64
	LastHeartbeat time.Time
}

type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Node
}

func NewRegistry() *Registry { return &Registry{nodes: map[string]Node{}} }

func (r *Registry) Announce(n Node) error {
	if r == nil {
		return errors.New("nil mesh registry")
	}
	if n.ID == "" {
		return errors.New("node id required")
	}
	if n.LastHeartbeat.IsZero() {
		n.LastHeartbeat = time.Now()
	}
	n = cloneNode(n)
	r.mu.Lock()
	if r.nodes == nil {
		r.nodes = make(map[string]Node)
	}
	r.nodes[n.ID] = n
	r.mu.Unlock()
	return nil
}

func (r *Registry) Remove(id string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	delete(r.nodes, id)
	r.mu.Unlock()
}

func (r *Registry) Discover(ctx context.Context, capability string) []Node {
	if r == nil {
		return nil
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
	r.mu.RLock()
	out := make([]Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		if capability == "" || hasCapability(n, capability) {
			out = append(out, cloneNode(n))
		}
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *Registry) Heartbeat(id string) error {
	if r == nil {
		return errors.New("nil mesh registry")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.nodes[id]
	if !ok {
		return errors.New("unknown node")
	}
	n.Status = "ready"
	n.LastHeartbeat = time.Now()
	r.nodes[id] = n
	return nil
}

func (r *Registry) MarkStale(maxAge time.Duration) int {
	if r == nil || maxAge <= 0 {
		return 0
	}
	now := time.Now()
	count := 0
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, n := range r.nodes {
		if !n.LastHeartbeat.IsZero() && now.Sub(n.LastHeartbeat) > maxAge && n.Status != "stale" {
			n.Status = "stale"
			r.nodes[id] = n
			count++
		}
	}
	return count
}

func (r *Registry) Snapshot() []Node { return r.Discover(context.Background(), "") }

func hasCapability(n Node, capability string) bool {
	for _, c := range n.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

func cloneNode(n Node) Node {
	n.Capabilities = append([]string(nil), n.Capabilities...)
	n.Transports = append([]string(nil), n.Transports...)
	return n
}
