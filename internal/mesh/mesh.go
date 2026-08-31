package mesh

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

type Node struct {
	ID           string   `json:"node_id"`
	Status       string   `json:"status"`
	Capabilities []string `json:"capabilities"`
	Transport    []string `json:"transport"`
	CPU          bool     `json:"cpu"`
	GPU          bool     `json:"gpu"`
	NPU          bool     `json:"npu"`
	RAMMB        int64    `json:"ram_mb"`
	VRAMMB       int64    `json:"vram_mb"`
	Workers      int      `json:"workers"`
}

type Message struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	TraceID   string          `json:"trace_id"`
	Source    string          `json:"source"`
	Target    string          `json:"target"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Node
}

func NewRegistry() *Registry { return &Registry{nodes: map[string]Node{}} }

func (r *Registry) Announce(n Node) error {
	if n.ID == "" {
		return errors.New("node id required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[n.ID] = n
	return nil
}

func (r *Registry) Remove(id string) { r.mu.Lock(); defer r.mu.Unlock(); delete(r.nodes, id) }

func (r *Registry) Get(id string) (Node, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.nodes[id]
	return n, ok
}

func (r *Registry) Discover(capability string) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Node{}
	for _, n := range r.nodes {
		if n.Status != "ready" {
			continue
		}
		for _, c := range n.Capabilities {
			if c == capability {
				out = append(out, n)
				break
			}
		}
	}
	return out
}

func (r *Registry) Snapshot() []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Node, 0, len(r.nodes))
	for _, n := range r.nodes {
		out = append(out, n)
	}
	return out
}

type Handler interface { Handle(context.Context, Message) (Message, error) }
type Router struct { mu sync.RWMutex; handlers map[string]Handler }

func NewRouter() *Router { return &Router{handlers: map[string]Handler{}} }
func (x *Router) Register(kind string, h Handler) { x.mu.Lock(); defer x.mu.Unlock(); x.handlers[kind] = h }
func (x *Router) Route(ctx context.Context, m Message) (Message, error) {
	x.mu.RLock()
	h := x.handlers[m.Type]
	x.mu.RUnlock()
	if h == nil { return Message{}, errors.New("no route") }
	return h.Handle(ctx, m)
}

func HealthHandler(reg *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "nodes": len(reg.Snapshot())})
	})
}
