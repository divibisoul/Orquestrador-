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
    if engine == nil { engine = orchestrator.NewEngine(4) }
    if meshRegistry == nil { meshRegistry = mesh.NewRegistry() }
    return &Fabric{nodes: map[NucleusID]Nucleus{}, mesh: meshRegistry, engine: engine, httpClient: client}
}

func (f *Fabric) Register(n Nucleus) error {
    if err := validateNucleus(n); err != nil { return err }
    f.mu.Lock()
    f.nodes[n.ID] = n
    f.mu.Unlock()
    return f.mesh.Announce(mesh.Node{ID: string(n.ID), Status: "ready", Capabilities: append([]string(nil), n.Capabilities...), Transports: append([]string(nil), n.Transports...), LastHeartbeat: time.Now()})
}

func (f *Fabric) RegisterDefaults(endpoints map[NucleusID]string) error {
    repos := map[NucleusID]string{
        N01: "divibisoul/aeternum-core-29",
        N02: "divibisoul/Eternium-",
        N03: "divibisoul/nexus-aeternum-fusion",
        N04: "divibisoul/nextjs-ai-chatbots",
        N05: "divibisoul/nextjs-ai-chatbot",
        N06: "divibisoul/nextjs-ai-chatbot-2000",
    }
    for _, id := range []NucleusID{N01, N02, N03, N04, N05, N06} {
        if err := f.Register(Nucleus{ID: id, Repository: repos[id], Endpoint: strings.TrimSpace(endpoints[id]), Capabilities: []string{"inference", "mesh"}, Transports: []string{MeshProtocol}}); err != nil { return err }
    }
    return nil
}

func (f *Fabric) Snapshot() []Nucleus {
    f.mu.RLock(); defer f.mu.RUnlock()
    out := make([]Nucleus, 0, len(f.nodes))
    for _, n := range f.nodes { out = append(out, n) }
    sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
    return out
}

func (f *Fabric) Dispatch(ctx context.Context, target NucleusID, traceID, goal string, metadata map[string]interface{}) (transport.Envelope, error) {
    if strings.TrimSpace(traceID) == "" { return transport.Envelope{}, errors.New("trace id required") }
    f.mu.RLock()
    n, ok := f.nodes[target]
    f.mu.RUnlock()
    if !ok { return transport.Envelope{}, fmt.Errorf("nucleus %s not registered", target) }
    if strings.TrimSpace(n.Endpoint) == "" { return transport.Envelope{}, fmt.Errorf("nucleus %s has no endpoint", target) }
    if len(n.Transports) == 0 { return transport.Envelope{}, fmt.Errorf("nucleus %s has no transport", target) }

    return f.httpClient.Do(ctx, n.Endpoint, transport.Envelope{
        EventID: fmt.Sprintf("%s-%d", target, time.Now().UnixNano()),
        EventType: "soul.task.request",
        TraceID: traceID,
        Source: "orquestrador",
        Target: string(target),
        Timestamp: time.Now().UnixMilli(),
        CompatibleSystems: []string{"SOUL", MeshProtocol},
        Payload: map[string]interface{}{"goal": goal, "metadata": metadata},
    })
}

func (f *Fabric) Heartbeat(id NucleusID) error { return f.mesh.Heartbeat(string(id)) }
func (f *Fabric) MarkStale(maxAge time.Duration) int { return f.mesh.MarkStale(maxAge) }

func validateNucleus(n Nucleus) error {
    switch n.ID { case N01, N02, N03, N04, N05, N06: default: return fmt.Errorf("unsupported nucleus %q", n.ID) }
    if strings.TrimSpace(n.Repository) == "" { return errors.New("repository required") }
    return nil
}
