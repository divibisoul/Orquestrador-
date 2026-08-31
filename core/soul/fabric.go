package soul

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/core/orchestrator"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/mesh/transport"
)

const MeshProtocol = "soul-mesh/1"

type NucleusID string

const (
	N01 NucleusID = "N01"
	N02 NucleusID = "N02"
	N03 NucleusID = "N03"
	N04 NucleusID = "N04"
	N05 NucleusID = "N05"
	N06 NucleusID = "N06"
)

type Nucleus struct {
	ID           NucleusID
	Repository   string
	Endpoint     string
	Capabilities []string
	Transports   []string
}

type Fabric struct {
	mu         sync.RWMutex
	nodes      map[NucleusID]Nucleus
	mesh       *mesh.Registry
	engine     *orchestrator.Engine
	httpClient transport.Client
}

func NewFabric(engine *orchestrator.Engine, meshRegistry *mesh.Registry, client transport.Client) *Fabric {
	if engine == nil {
		engine = orchestrator.NewEngine(4)
	}
	if meshRegistry == nil {
		meshRegistry = mesh.NewRegistry()
	}
	return &Fabric{nodes: map[NucleusID]Nucleus{}, mesh: meshRegistry, engine: engine, httpClient: client}
}

func (f *Fabric) Register(n Nucleus) error {
	if f == nil {
		return errors.New("nil soul fabric")
	}
	if err := validateNucleus(n); err != nil {
		return err
	}
	f.mu.Lock()
	if f.nodes == nil {
		f.nodes = make(map[NucleusID]Nucleus)
	}
	f.nodes[n.ID] = Nucleus{ID: n.ID, Repository: n.Repository, Endpoint: n.Endpoint, Capabilities: append([]string(nil), n.Capabilities...), Transports: append([]string(nil), n.Transports...)}
	meshRegistry := f.mesh
	f.mu.Unlock()
	if meshRegistry == nil {
		return errors.New("mesh registry unavailable")
	}
	status := "registered"
	if strings.TrimSpace(n.Endpoint) != "" && len(n.Transports) > 0 {
		status = "ready"
	}
	return meshRegistry.Announce(mesh.Node{ID: string(n.ID), Status: status, Capabilities: append([]string(nil), n.Capabilities...), Transports: append([]string(nil), n.Transports...), LastHeartbeat: time.Now()})
}

func (f *Fabric) RegisterDefaults(endpoints map[NucleusID]string) error {
	repos := map[NucleusID]string{N01: "divibisoul/aeternum-core-29", N02: "divibisoul/Eternium-", N03: "divibisoul/nexus-aeternum-fusion", N04: "divibisoul/nextjs-ai-chatbots", N05: "divibisoul/nextjs-ai-chatbot", N06: "divibisoul/nextjs-ai-chatbot-2000"}
	for _, id := range []NucleusID{N01, N02, N03, N04, N05, N06} {
		if err := f.Register(Nucleus{ID: id, Repository: repos[id], Endpoint: strings.TrimSpace(endpoints[id]), Capabilities: []string{"inference", "mesh"}, Transports: []string{MeshProtocol}}); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fabric) Snapshot() []Nucleus {
	if f == nil {
		return nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]Nucleus, 0, len(f.nodes))
	for _, n := range f.nodes {
		n.Capabilities = append([]string(nil), n.Capabilities...)
		n.Transports = append([]string(nil), n.Transports...)
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *Fabric) Dispatch(ctx context.Context, target NucleusID, traceID, goal string, metadata map[string]interface{}) (transport.Envelope, error) {
	if f == nil {
		return transport.Envelope{}, errors.New("nil soul fabric")
	}
	if ctx == nil {
		return transport.Envelope{}, errors.New("nil context")
	}
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return transport.Envelope{}, errors.New("trace id required")
	}
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return transport.Envelope{}, errors.New("goal required")
	}
	f.mu.RLock()
	n, ok := f.nodes[target]
	client := f.httpClient
	f.mu.RUnlock()
	if !ok {
		return transport.Envelope{}, fmt.Errorf("nucleus %s not registered", target)
	}
	if strings.TrimSpace(n.Endpoint) == "" {
		return transport.Envelope{}, fmt.Errorf("nucleus %s has no endpoint", target)
	}
	if len(n.Transports) == 0 {
		return transport.Envelope{}, fmt.Errorf("nucleus %s has no transport", target)
	}
	request := transport.Envelope{EventID: fmt.Sprintf("%s-%d", target, time.Now().UnixNano()), EventType: "soul.task.request", TraceID: traceID, Source: "orquestrador", Target: string(target), Timestamp: time.Now().UnixMilli(), CompatibleSystems: []string{"SOUL", MeshProtocol}, Payload: map[string]interface{}{"goal": goal, "metadata": metadata}}
	response, err := client.Do(ctx, n.Endpoint, request)
	if err != nil {
		return transport.Envelope{}, fmt.Errorf("dispatch %s: %w", target, err)
	}
	if response.TraceID != "" && response.TraceID != traceID {
		return transport.Envelope{}, fmt.Errorf("trace mismatch for %s: sent %q received %q", target, traceID, response.TraceID)
	}
	if response.Source != "" && response.Source != string(target) {
		return transport.Envelope{}, fmt.Errorf("source mismatch for %s: received %q", target, response.Source)
	}
	if response.TraceID == "" {
		response.TraceID = traceID
	}
	return response, nil
}

func (f *Fabric) Heartbeat(id NucleusID) error {
	if f == nil || f.mesh == nil {
		return errors.New("mesh registry unavailable")
	}
	return f.mesh.Heartbeat(string(id))
}

func (f *Fabric) MarkStale(maxAge time.Duration) int {
	if f == nil || f.mesh == nil {
		return 0
	}
	return f.mesh.MarkStale(maxAge)
}

func validateNucleus(n Nucleus) error {
	switch n.ID {
	case N01, N02, N03, N04, N05, N06:
	default:
		return fmt.Errorf("unsupported nucleus %q", n.ID)
	}
	if strings.TrimSpace(n.Repository) == "" {
		return errors.New("repository required")
	}
	return nil
}
