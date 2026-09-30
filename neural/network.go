package neural

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type Edge struct {
	From, To int
	Weight   float64
	Name     string
}

type Layer struct {
	Activation  string
	DropoutRate float64
}

type Config struct {
	Layers         []Layer
	Optimizer      string
	Regularization float64
	GradientClip   float64
	Heads          int
	BatchCache     int
}

type NetworkStats struct {
	LearningSteps uint64
	LastUpdate    time.Time
	LastGradient  float64
	LastLoss      float64
	Activations   uint64
	CacheHits     uint64
}

type Network struct {
	mu           sync.RWMutex
	size         int
	edges        map[int][]Edge
	bias         []float64
	learningRate float64
	config       Config
	adamM        []float64
	adamV        []float64
	edgeM        map[string]float64
	edgeV        map[string]float64
	stats        NetworkStats
	cache        map[string][]float64
}

func New(size int, learningRate float64) (*Network, error) {
	if size < 1 {
		return nil, errors.New("network size must be positive")
	}
	if learningRate < 1e-6 || learningRate > 0.1 || math.IsNaN(learningRate) || math.IsInf(learningRate, 0) {
		return nil, errors.New("learning rate must be finite and in [1e-6,0.1]")
	}
	n := &Network{
		size:         size,
		edges:        make(map[int][]Edge),
		bias:         make([]float64, size),
		learningRate: learningRate,
		adamM:        make([]float64, size),
		adamV:        make([]float64, size),
		edgeM:        make(map[string]float64),
		edgeV:        make(map[string]float64),
		cache:        make(map[string][]float64),
	}
	if err := n.Configure(Config{
		Layers:         []Layer{{Activation: "tanh"}},
		Optimizer:      "adam",
		Regularization: 1e-6,
		GradientClip:   1.0,
		Heads:          1,
		BatchCache:     128,
	}); err != nil {
		return nil, err
	}
	return n, nil
}

func validateConfig(config Config) error {
	if len(config.Layers) == 0 {
		return errors.New("at least one neural layer is required")
	}
	if config.Optimizer != "sgd" && config.Optimizer != "rmsprop" && config.Optimizer != "adam" {
		return errors.New("optimizer must be one of sgd, rmsprop, adam")
	}
	if config.Regularization < 0 || math.IsNaN(config.Regularization) || math.IsInf(config.Regularization, 0) {
		return errors.New("regularization must be finite and non-negative")
	}
	if config.GradientClip < 0 || math.IsNaN(config.GradientClip) || math.IsInf(config.GradientClip, 0) {
		return errors.New("gradient clip must be finite and non-negative")
	}
	if config.Heads < 1 {
		return errors.New("attention heads must be positive")
	}
	if config.BatchCache < 1 {
		return errors.New("batch cache capacity must be positive")
	}
	for _, layer := range config.Layers {
		switch layer.Activation {
		case "relu", "sigmoid", "linear", "tanh", "":
		default:
			return errors.New("unsupported activation: " + layer.Activation)
		}
		if layer.DropoutRate < 0 || layer.DropoutRate >= 1 || math.IsNaN(layer.DropoutRate) || math.IsInf(layer.DropoutRate, 0) {
			return errors.New("dropout rate must be finite and in [0,1)")
		}
	}
	return nil
}

// Configure makes the previously declared layer/optimizer configuration executable
// instead of leaving it as passive metadata. Parameters are preserved; changing the
// graph configuration only invalidates stale forward-cache entries.
func (n *Network) Configure(config Config) error {
	if n == nil {
		return errors.New("network unavailable")
	}
	if err := validateConfig(config); err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	config.Layers = append([]Layer(nil), config.Layers...)
	n.config = config
	n.cache = make(map[string][]float64)
	return nil
}

func (n *Network) AddEdge(from, to int, weight float64) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if from < 0 || from >= n.size || to < 0 || to >= n.size {
		return errors.New("edge index out of range")
	}
	if from == to {
		return errors.New("self-cycle not allowed")
	}
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		return errors.New("invalid edge weight")
	}
	if n.pathExistsLocked(to, from) {
		return errors.New("edge would create cycle")
	}
	for i, e := range n.edges[from] {
		if e.To == to {
			n.edges[from][i].Weight = weight
			n.cache = make(map[string][]float64)
			return nil
		}
	}
	n.edges[from] = append(n.edges[from], Edge{From: from, To: to, Weight: weight, Name: fmt.Sprintf("edge-%d-%d", from, to)})
	n.cache = make(map[string][]float64)
	return nil
}

func (n *Network) pathExistsLocked(from, to int) bool {
	seen := map[int]bool{}
	var dfs func(int) bool
	dfs = func(v int) bool {
		if v == to {
			return true
		}
		if seen[v] {
			return false
		}
		seen[v] = true
		for _, e := range n.edges[v] {
			if dfs(e.To) {
				return true
			}
		}
		return false
	}
	return dfs(from)
}

func (n *Network) RemoveEdge(from, to int) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if from < 0 || from >= n.size || to < 0 || to >= n.size {
		return errors.New("edge index out of range")
	}
	edges := n.edges[from]
	for i, e := range edges {
		if e.To == to {
			copy(edges[i:], edges[i+1:])
			n.edges[from] = edges[:len(edges)-1]
			n.cache = make(map[string][]float64)
			delete(n.edgeM, fmt.Sprintf("%d:%d", from, to))
			delete(n.edgeV, fmt.Sprintf("%d:%d", from, to))
			return nil
		}
	}
	return errors.New("edge not found")
}

func (n *Network) Activate(inputs []float64) ([]float64, error) {
	if len(inputs) == 0 {
		return nil, errors.New("empty input")
	}
	if len(inputs)%n.size != 0 {
		return nil, errors.New("input size must equal network size or an exact batch multiple")
	}
	n.mu.RLock()
	bias := append([]float64(nil), n.bias...)
	activation := n.config.Layers[0].Activation
	n.mu.RUnlock()
	out := make([]float64, len(inputs))
	for off := 0; off < len(inputs); off += n.size {
		for i := 0; i < n.size; i++ {
			v := inputs[off+i]
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, errors.New("input contains non-finite value")
			}
			out[off+i] = activateValue(v+bias[i], activation)
		}
	}
	n.mu.Lock()
	n.stats.Activations += uint64(len(inputs) / n.size)
	n.mu.Unlock()
	return out, nil
}

func activateValue(v float64, a string) float64 {
	switch a {
	case "relu":
		if v > 0 {
			return v
		}
		return 0
	case "sigmoid":
		// Stable sigmoid evaluation for large negative/positive inputs.
		if v >= 0 {
			z := math.Exp(-v)
			return 1 / (1 + z)
		}
		z := math.Exp(v)
		return z / (1 + z)
	case "linear":
		return v
	default:
		return math.Tanh(v)
	}
}

func (n *Network) forwardTrace(ctx context.Context, inputs []float64) ([][]float64, [][]float64, error) {
	if ctx == nil {
		return nil, nil, errors.New("context is nil")
	}
	if len(inputs) != n.size {
		return nil, nil, errors.New("input size mismatch")
	}
	for _, v := range inputs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, nil, errors.New("input contains non-finite value")
		}
	}

	n.mu.RLock()
	bias := append([]float64(nil), n.bias...)
	edges := make(map[int][]Edge, len(n.edges))
	for from, es := range n.edges {
		edges[from] = append([]Edge(nil), es...)
	}
	cfg := n.config
	n.mu.RUnlock()

	states := make([][]float64, len(cfg.Layers)+1)
	preActivations := make([][]float64, len(cfg.Layers))
	states[0] = append([]float64(nil), inputs...)

	for pass, layer := range cfg.Layers {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}
		z := append([]float64(nil), bias...)
		for i, value := range states[pass] {
			z[i] += value
		}
		for from, es := range edges {
			for _, e := range es {
				z[e.To] += states[pass][from] * e.Weight
			}
		}
		next := make([]float64, n.size)
		for i := range z {
			next[i] = activateValue(z[i], layer.Activation)
		}
		preActivations[pass] = z
		states[pass+1] = next
	}
	return states, preActivations, nil
}

func (n *Network) Forward(ctx context.Context, inputs []float64) ([]float64, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if len(inputs) != n.size {
		return nil, errors.New("input size mismatch")
	}
	for _, v := range inputs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("input contains non-finite value")
		}
	}
	key := cacheKey(inputs)
	n.mu.RLock()
	if cached, ok := n.cache[key]; ok {
		out := append([]float64(nil), cached...)
		n.mu.RUnlock()
		n.mu.Lock()
		n.stats.CacheHits++
		n.mu.Unlock()
		return out, nil
	}
	n.mu.RUnlock()

	states, _, err := n.forwardTrace(ctx, inputs)
	if err != nil {
		return nil, err
	}
	out := states[len(states)-1]

	n.mu.Lock()
	n.cache[key] = append([]float64(nil), out...)
	cfg := n.config
	if len(n.cache) > cfg.BatchCache {
		for k := range n.cache {
			delete(n.cache, k)
			break
		}
	}
	n.stats.Activations++
	n.mu.Unlock()
	return append([]float64(nil), out...), nil
}

func cacheKey(v []float64) string {
	buf := make([]byte, len(v)*8)
	for i, x := range v {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(x))
	}
	return string(buf)
}

func finiteVector(values []float64) error {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("vector contains non-finite value")
		}
	}
	return nil
}

func mseLoss(pred, target []float64) (float64, error) {
	if len(pred) == 0 || len(pred) != len(target) {
		return 0, errors.New("loss vectors must match and be non-empty")
	}
	sum := 0.0
	for i := range pred {
		delta := pred[i] - target[i]
		sum += .5 * delta * delta
	}
	return sum / float64(len(pred)), nil
}

// Learn performs actual backpropagation through every configured graph layer.
// The graph update is synchronous: z_i^l = b_i + h_i^l + Σ_j w_{ji}h_j^l,
// h_i^{l+1}=φ_l(z_i^l). Parameters are shared across layers, therefore the
// parameter gradients are accumulated over all passes before the optimizer step.
// The objective is mean half-squared error plus L2 regularization.
func (n *Network) Learn(inputs, target []float64) error {
	if len(inputs) != n.size || len(target) != n.size {
		return errors.New("training vector size mismatch")
	}
	if err := finiteVector(inputs); err != nil {
		return errors.New("training data contains non-finite value")
	}
	if err := finiteVector(target); err != nil {
		return errors.New("training data contains non-finite value")
	}

	states, pre, err := n.forwardTrace(context.Background(), inputs)
	if err != nil {
		return err
	}
	pred := states[len(states)-1]
	loss, err := mseLoss(pred, target)
	if err != nil {
		return err
	}

	n.mu.RLock()
	cfg := n.config
	edges := make(map[int][]Edge, len(n.edges))
	for from, es := range n.edges {
		edges[from] = append(edges[from], es...)
	}
	n.mu.RUnlock()

	gradBias := make([]float64, n.size)
	gradEdges := make(map[string]float64)
	gNext := make([]float64, n.size)
	for i := range pred {
		gNext[i] = (pred[i] - target[i]) / float64(n.size)
	}

	for layer := len(cfg.Layers) - 1; layer >= 0; layer-- {
		delta := make([]float64, n.size)
		for i := 0; i < n.size; i++ {
			delta[i] = gNext[i] * activationDerivative(pre[layer][i], states[layer+1][i], cfg.Layers[layer].Activation)
			gradBias[i] += delta[i]
		}
		for from, es := range edges {
			for _, e := range es {
				gradEdges[fmt.Sprintf("%d:%d", from, e.To)] += delta[e.To] * states[layer][from]
			}
		}
		gPrev := make([]float64, n.size)
		copy(gPrev, delta) // residual/self path h_i^l -> z_i^l
		for from, es := range edges {
			for _, e := range es {
				gPrev[from] += delta[e.To] * e.Weight
			}
		}
		gNext = gPrev
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	for i := range gradBias {
		gradBias[i] += cfg.Regularization * n.bias[i]
	}
	for from, es := range n.edges {
		for _, e := range es {
			key := fmt.Sprintf("%d:%d", from, e.To)
			gradEdges[key] += cfg.Regularization * e.Weight
		}
	}

	normSq := 0.0
	for _, g := range gradBias {
		normSq += g * g
	}
	for _, g := range gradEdges {
		normSq += g * g
	}
	norm := math.Sqrt(normSq)
	if cfg.GradientClip > 0 && norm > cfg.GradientClip {
		scale := cfg.GradientClip / norm
		for i := range gradBias {
			gradBias[i] *= scale
		}
		for key, g := range gradEdges {
			gradEdges[key] = g * scale
		}
		norm = cfg.GradientClip
	}

	step := n.stats.LearningSteps + 1
	for i, grad := range gradBias {
		n.bias[i] = n.updateParameter(n.bias[i], grad, &n.adamM[i], &n.adamV[i], fmt.Sprintf("bias:%d", i), step)
	}
	for from, es := range n.edges {
		for idx, e := range es {
			key := fmt.Sprintf("%d:%d", from, e.To)
			grad := gradEdges[key]
			m := n.edgeM[key]
			v := n.edgeV[key]
			weight := n.updateParameter(e.Weight, grad, &m, &v, key, step)
			n.edgeM[key], n.edgeV[key] = m, v
			n.edges[from][idx].Weight = weight
		}
	}
	reg := 0.0
	for _, b := range n.bias {
		reg += b * b
	}
	for _, es := range n.edges {
		for _, e := range es {
			reg += e.Weight * e.Weight
		}
	}
	n.stats.LearningSteps = step
	n.stats.LastUpdate = time.Now().UTC()
	n.stats.LastGradient = math.Sqrt(normSq) / float64(n.size)
	n.stats.LastLoss = loss + .5*cfg.Regularization*reg
	n.cache = make(map[string][]float64)
	return nil
}

func (n *Network) updateParameter(value, grad float64, m, v *float64, key string, step uint64) float64 {
	switch n.config.Optimizer {
	case "sgd":
		return value - n.learningRate*grad
	case "rmsprop":
		*v = .9*(*v) + .1*grad*grad
		return value - n.learningRate*grad/(math.Sqrt(*v)+1e-8)
	default:
		*m = .9*(*m) + .1*grad
		*v = .999*(*v) + .001*grad*grad
		t := float64(step)
		mh := *m / (1 - math.Pow(.9, t))
		vh := *v / (1 - math.Pow(.999, t))
		return value - n.learningRate*mh/(math.Sqrt(vh)+1e-8)
	}
}

func activationDerivative(x, y float64, a string) float64 {
	switch a {
	case "relu":
		if x > 0 {
			return 1
		}
		return 0
	case "sigmoid":
		return y * (1 - y)
	case "linear":
		return 1
	default:
		return 1 - y*y
	}
}

// Backprop computes the true gradient of the network loss with respect to the
// input vector without mutating network parameters.
func (n *Network) Backprop(inputs, target []float64) ([]float64, error) {
	if len(inputs) != n.size || len(target) != n.size || len(inputs) == 0 {
		return nil, errors.New("backprop vectors must match network size")
	}
	if err := finiteVector(inputs); err != nil {
		return nil, errors.New("backprop data contains non-finite input")
	}
	if err := finiteVector(target); err != nil {
		return nil, errors.New("backprop data contains non-finite target")
	}

	states, pre, err := n.forwardTrace(context.Background(), inputs)
	if err != nil {
		return nil, err
	}
	pred := states[len(states)-1]
	gNext := make([]float64, n.size)
	for i := range pred {
		gNext[i] = (pred[i] - target[i]) / float64(n.size)
	}

	n.mu.RLock()
	cfg := n.config
	edges := make(map[int][]Edge, len(n.edges))
	for from, es := range n.edges {
		edges[from] = append(edges[from], es...)
	}
	n.mu.RUnlock()

	for layer := len(cfg.Layers) - 1; layer >= 0; layer-- {
		delta := make([]float64, n.size)
		for i := range delta {
			delta[i] = gNext[i] * activationDerivative(pre[layer][i], states[layer+1][i], cfg.Layers[layer].Activation)
		}
		gPrev := make([]float64, n.size)
		copy(gPrev, delta)
		for from, es := range edges {
			for _, e := range es {
				gPrev[from] += delta[e.To] * e.Weight
			}
		}
		gNext = gPrev
	}

	if err := finiteVector(gNext); err != nil {
		return nil, errors.New("backprop produced non-finite gradient")
	}
	sumAbs := 0.0
	for _, g := range gNext {
		sumAbs += math.Abs(g)
	}
	n.mu.Lock()
	n.stats.LastGradient = sumAbs / float64(len(gNext))
	n.mu.Unlock()
	return gNext, nil
}

func (n *Network) Normalize(values []float64) ([]float64, error) {
	if len(values) == 0 {
		return nil, errors.New("empty vector")
	}
	mean := 0.0
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("invalid vector")
		}
		mean += v
	}
	mean /= float64(len(values))
	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	std := math.Sqrt(variance/float64(len(values)) + 1e-12)
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = (v - mean) / std
	}
	return out, nil
}

func (n *Network) Attention(query, keys, values []float64) ([]float64, error) {
	if len(query) == 0 || len(keys) == 0 || len(keys) != len(values) {
		return nil, errors.New("attention requires non-empty query, equal non-zero keys and values")
	}
	all := append(append(append([]float64{}, query...), keys...), values...)
	if err := finiteVector(all); err != nil {
		return nil, errors.New("attention input contains non-finite value")
	}
	n.mu.RLock()
	heads := n.config.Heads
	n.mu.RUnlock()
	scores := make([]float64, len(keys))
	maxScore := -math.MaxFloat64
	scale := math.Sqrt(float64(len(query)))
	for i, k := range keys {
		q := query[i%len(query)]
		if heads > 1 {
			q *= 1 + float64(i%heads)/float64(heads)
		}
		scores[i] = q * k / scale
		if scores[i] > maxScore {
			maxScore = scores[i]
		}
	}
	sum := 0.0
	for i := range scores {
		scores[i] = math.Exp(scores[i] - maxScore)
		sum += scores[i]
	}
	if math.IsNaN(sum) || math.IsInf(sum, 0) || sum <= 0 {
		return nil, errors.New("attention normalization failed")
	}
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = v * scores[i] / sum
	}
	return out, nil
}

func (n *Network) Health() map[string]any {
	n.mu.RLock()
	defer n.mu.RUnlock()
	edges := 0
	for _, e := range n.edges {
		edges += len(e)
	}
	density := 0.0
	if n.size > 1 {
		density = float64(edges) / float64(n.size*(n.size-1))
	}
	return map[string]any{
		"status":         "ready",
		"size":           n.size,
		"edges":          edges,
		"density":        density,
		"learning_steps": n.stats.LearningSteps,
		"last_update":    n.stats.LastUpdate,
		"last_gradient":  n.stats.LastGradient,
		"last_loss":      n.stats.LastLoss,
		"activations":    n.stats.Activations,
		"cache_hits":     n.stats.CacheHits,
		"optimizer":      n.config.Optimizer,
		"heads":           n.config.Heads,
		"layers":          len(n.config.Layers),
		"gradient_clip":   n.config.GradientClip,
		"regularization": n.config.Regularization,
	}
}
