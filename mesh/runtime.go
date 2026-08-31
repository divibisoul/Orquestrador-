package mesh

import("context";"errors";"sync";"time")

type Node struct{ID string; Status string; Capabilities []string; Transports []string; CPU,GPU,NPU bool; RAMMB,VRAMMB int64; LatencyMS float64}
type Registry struct{mu sync.RWMutex; nodes map[string]Node}
func NewRegistry()*Registry{return &Registry{nodes:map[string]Node{}}}
func(r *Registry)Announce(n Node)error{if n.ID==""{return errors.New("node id required")};r.mu.Lock();r.nodes[n.ID]=n;r.mu.Unlock();return nil}
func(r *Registry)Remove(id string){r.mu.Lock();delete(r.nodes,id);r.mu.Unlock()}
func(r *Registry)Discover(_ context.Context,capability string)[]Node{r.mu.RLock();defer r.mu.RUnlock();out:=[]Node{};for _,n:=range r.nodes{for _,c:=range n.Capabilities{if capability==""||c==capability{out=append(out,n);break}}};return out}
func(r *Registry)Heartbeat(id string)error{r.mu.Lock();defer r.mu.Unlock();n,ok:=r.nodes[id];if !ok{return errors.New("unknown node")};n.Status="ready";r.nodes[id]=n;return nil}
func(r *Registry)MarkStale(maxAge time.Duration)int{_ = maxAge;return 0}
func(r *Registry)Snapshot()[]Node{return r.Discover(context.Background(),"")}
