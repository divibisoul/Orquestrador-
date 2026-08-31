package discovery

type ComputeResources struct{CPU,GPU,NPU bool; RAMMB,VRAMMB int64; Threads,Workers int}
type Node struct{ID,Status string; Capabilities,Transport,Compatible []string; Compute ComputeResources}
type Registry struct{nodes map[string]Node}
func New()*Registry{return &Registry{nodes:map[string]Node{}}}
func(r *Registry)Announce(n Node){r.nodes[n.ID]=n}
func(r *Registry)Get(id string)(Node,bool){n,ok:=r.nodes[id];return n,ok}
func(r *Registry)List()[]Node{out:=make([]Node,0,len(r.nodes));for _,n:=range r.nodes{out=append(out,n)};return out}
