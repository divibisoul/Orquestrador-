package neural

import (
	"context"
	"errors"
	"math"
	"sync"
)

type Edge struct { From, To int; Weight float64 }
type Network struct { mu sync.RWMutex; size int; edges map[int][]Edge; bias []float64; learningRate float64; steps uint64 }

func New(size int, learningRate float64) (*Network, error) {
	if size < 1 { return nil, errors.New("network size must be positive") }
	if learningRate <= 0 || learningRate > 1 { return nil, errors.New("learning rate must be in (0,1]") }
	return &Network{size:size, edges:make(map[int][]Edge), bias:make([]float64,size), learningRate:learningRate}, nil
}

func (n *Network) AddEdge(from, to int, weight float64) error {
	n.mu.Lock(); defer n.mu.Unlock()
	if from < 0 || from >= n.size || to < 0 || to >= n.size { return errors.New("edge index out of range") }
	if math.IsNaN(weight) || math.IsInf(weight,0) { return errors.New("invalid edge weight") }
	for i, e := range n.edges[from] { if e.To == to { n.edges[from][i].Weight = weight; return nil } }
	n.edges[from] = append(n.edges[from], Edge{From:from, To:to, Weight:weight})
	return nil
}

func (n *Network) RemoveEdge(from, to int) error {
	n.mu.Lock(); defer n.mu.Unlock()
	if from < 0 || from >= n.size || to < 0 || to >= n.size { return errors.New("edge index out of range") }
	edges := n.edges[from]
	for i,e := range edges { if e.To == to { n.edges[from] = append(edges[:i], edges[i+1:]...); return nil } }
	return errors.New("edge not found")
}

func (n *Network) Activate(inputs []float64) ([]float64, error) {
	n.mu.RLock(); defer n.mu.RUnlock()
	if len(inputs) != n.size { return nil, errors.New("input size mismatch") }
	out := append([]float64(nil), inputs...)
	for i := range out { out[i] = math.Tanh(out[i] + n.bias[i]) }
	return out, nil
}

func (n *Network) Forward(ctx context.Context, inputs []float64) ([]float64, error) {
	if ctx == nil { return nil, errors.New("context is nil") }
	select { case <-ctx.Done(): return nil, ctx.Err(); default: }
	if len(inputs) != n.size { return nil, errors.New("input size mismatch") }
	n.mu.RLock(); defer n.mu.RUnlock()
	state := append([]float64(nil), inputs...)
	for layer := 0; layer < n.size; layer++ {
		select { case <-ctx.Done(): return nil, ctx.Err(); default: }
		next := make([]float64,n.size)
		for i := 0; i < n.size; i++ { next[i] = state[i] + n.bias[i] }
		for from, edges := range n.edges { for _,e := range edges { next[e.To] += state[from]*e.Weight } }
		for i := range next { next[i] = math.Tanh(next[i]) }
		state = next
	}
	return state,nil
}

func (n *Network) Learn(inputs, target []float64) error {
	n.mu.Lock(); defer n.mu.Unlock()
	if len(inputs)!=n.size || len(target)!=n.size { return errors.New("training vector size mismatch") }
	pred := make([]float64,n.size)
	for i := range pred { pred[i] = math.Tanh(inputs[i]+n.bias[i]) }
	for i := range pred { err := target[i]-pred[i]; n.bias[i] += n.learningRate*err*(1-pred[i]*pred[i]) }
	n.steps++
	return nil
}

func (n *Network) Normalize(values []float64) ([]float64,error) {
	if len(values)==0 { return nil, errors.New("empty vector") }
	mean:=0.0; for _,v:=range values { if math.IsNaN(v)||math.IsInf(v,0){return nil,errors.New("invalid vector")}; mean+=v }; mean/=float64(len(values))
	variance:=0.0; for _,v:=range values { d:=v-mean; variance+=d*d }; variance/=float64(len(values)); std:=math.Sqrt(variance+1e-12)
	out:=make([]float64,len(values)); for i,v:=range values { out[i]=(v-mean)/std }; return out,nil
}

func (n *Network) Attention(query, keys, values []float64) ([]float64,error) {
	if len(keys)==0 || len(keys)!=len(values) { return nil, errors.New("attention inputs must have equal non-zero length") }
	scores:=make([]float64,len(keys)); maxScore:=-math.MaxFloat64
	for i,k:=range keys { scores[i]=query[0]*k; if scores[i]>maxScore{maxScore=scores[i]} }
	sum:=0.0; for i:=range scores { scores[i]=math.Exp(scores[i]-maxScore); sum+=scores[i] }; if sum==0{return nil,errors.New("attention normalization failed")}
	out:=make([]float64,len(values)); for i,v:=range values { out[i]=v*scores[i]/sum }; return out,nil
}

func (n *Network) Backprop(inputs, target []float64) ([]float64,error) {
	if len(inputs)!=len(target) || len(inputs)==0{return nil,errors.New("backprop vectors must match")}
	grad:=make([]float64,len(inputs)); for i:=range inputs { pred:=math.Tanh(inputs[i]); grad[i]=(pred-target[i])*(1-pred*pred) }; return grad,nil
}

func (n *Network) Health() map[string]any {
	n.mu.RLock(); defer n.mu.RUnlock(); edges:=0; for _,e:=range n.edges{edges+=len(e)}
	return map[string]any{"status":"ready","size":n.size,"edges":edges,"learning_steps":n.steps,"learning_rate":n.learningRate}
}
