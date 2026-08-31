package fractal

import("sort";"sync";"sync/atomic")
type SubOrchestrator struct{ID string; Queue []string; Active bool}
type Manager struct{mu sync.Mutex; nodes map[string]*SubOrchestrator; seq uint64}
func New()*Manager{return &Manager{nodes:map[string]*SubOrchestrator{}}}
// 13 SpawnSubOrchestrator creates a child when demand exceeds the threshold.
func(m *Manager)SpawnSubOrchestrator(queue []string)*SubOrchestrator{if len(queue)<=100{return nil};m.mu.Lock();defer m.mu.Unlock();id:=atomic.AddUint64(&m.seq,1);n:=&SubOrchestrator{ID:"sub-"+itoa(id),Queue:append([]string(nil),queue[100:]...),Active:true};m.nodes[n.ID]=n;return n}
// 14 KillSubOrchestrator safely retires a child.
func(m *Manager)KillSubOrchestrator(id string)bool{m.mu.Lock();defer m.mu.Unlock();n:=m.nodes[id];if n==nil{return false};n.Active=false;delete(m.nodes,id);return true}
// 15 RebalanceTasks performs deterministic work-stealing by moving work to the least loaded child.
func(m *Manager)RebalanceTasks(){m.mu.Lock();defer m.mu.Unlock();nodes:=[]*SubOrchestrator{};for _,n:=range m.nodes{if n.Active{nodes=append(nodes,n)}};sort.Slice(nodes,func(i,j int)bool{return len(nodes[i].Queue)<len(nodes[j].Queue)});if len(nodes)<2{return};for len(nodes[len(nodes)-1].Queue)>len(nodes[0].Queue)+1{from,to:=nodes[len(nodes)-1],nodes[0];task:=from.Queue[len(from.Queue)-1];from.Queue=from.Queue[:len(from.Queue)-1];to.Queue=append(to.Queue,task);sort.Slice(nodes,func(i,j int)bool{return len(nodes[i].Queue)<len(nodes[j].Queue)})}}
func itoa(v uint64)string{if v==0{return "0"};b:=make([]byte,0,20);for v>0{b=append([]byte{byte('0'+v%10)},b...);v/=10};return string(b)}
