package learning

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
)

type EventType string

const (
	EventSupervised EventType = "supervised"
	EventFeedback   EventType = "feedback"
)

type Experience struct {
	ID            string
	TraceID       string
	CorrelationID string
	Source        string
	Target        string
	Capability    string
	EventType     EventType
	Outcome       string
	Reward        float64
	Confidence    float64
	Input         []float64
	TargetVector  []float64
	Provenance    string
	Timestamp     time.Time
	Metadata      map[string]string
}

type PersistedExperience struct {
	ID            string            `json:"id"`
	TraceID       string            `json:"trace_id"`
	CorrelationID string            `json:"correlation_id"`
	Source        string            `json:"source"`
	Target        string            `json:"target"`
	Capability    string            `json:"capability"`
	EventType     string            `json:"event_type"`
	Outcome       string            `json:"outcome"`
	Reward        float64           `json:"reward"`
	Confidence    float64           `json:"confidence"`
	Input         []float64         `json:"input"`
	TargetVector  []float64         `json:"target_vector"`
	Provenance    string            `json:"provenance"`
	Timestamp     time.Time         `json:"created_at"`
	Metadata      map[string]string `json:"metadata"`
}

type Store interface {
	RecordLearning(context.Context, PersistedExperience) error
	LoadLearning(context.Context, int) ([]PersistedExperience, error)
}

type RouteState struct {
	Attempts   uint64    `json:"attempts"`
	Successes  uint64    `json:"successes"`
	RewardSum  float64   `json:"reward_sum"`
	Confidence float64   `json:"confidence_sum"`
	Weight     float64   `json:"weight"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Snapshot struct {
	Status          string                `json:"status"`
	Experiences     uint64                `json:"experiences"`
	Supervised      uint64                `json:"supervised"`
	Feedback        uint64                `json:"feedback"`
	PersistFailures uint64                `json:"persist_failures"`
	Restored        uint64                `json:"restored"`
	LearnedRoutes   int                   `json:"learned_routes"`
	Routes          map[string]RouteState `json:"routes"`
}

type Machine struct {
	mu              sync.RWMutex
	neural          *neural.Network
	cortex          *prefrontal.Cortex
	store           Store
	learningRate    float64
	routes          map[string]RouteState
	experiences     atomic.Uint64
	supervised      atomic.Uint64
	feedback        atomic.Uint64
	persistFailures atomic.Uint64
	restored        atomic.Uint64
}

func New(n *neural.Network, c *prefrontal.Cortex, store Store) (*Machine, error) {
	if n == nil {
		return nil, errors.New("learning neural network is required")
	}
	if c == nil {
		return nil, errors.New("learning prefrontal cortex is required")
	}
	return &Machine{
		neural:       n,
		cortex:       c,
		store:        store,
		learningRate: 0.15,
		routes:       make(map[string]RouteState),
	}, nil
}

func validateExperience(exp Experience) error {
	if strings.TrimSpace(exp.ID) == "" {
		return errors.New("learning experience id is required")
	}
	if strings.TrimSpace(exp.Source) == "" {
		return errors.New("learning experience source is required")
	}
	if strings.TrimSpace(exp.Capability) == "" {
		return errors.New("learning capability is required")
	}
	if exp.EventType != EventFeedback && exp.EventType != EventSupervised {
		return errors.New("unsupported learning event type")
	}
	if exp.Confidence < 0 || exp.Confidence > 1 || math.IsNaN(exp.Confidence) || math.IsInf(exp.Confidence, 0) {
		return errors.New("learning confidence must be finite and within [0,1]")
	}
	if exp.Reward < -1 || exp.Reward > 1 || math.IsNaN(exp.Reward) || math.IsInf(exp.Reward, 0) {
		return errors.New("learning reward must be finite and within [-1,1]")
	}
	for _, value := range append(append([]float64{}, exp.Input...), exp.TargetVector...) {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("learning vectors must contain only finite values")
		}
	}
	if exp.EventType == EventSupervised && (len(exp.Input) == 0 || len(exp.Input) != len(exp.TargetVector)) {
		return errors.New("supervised learning requires equal non-empty input and target vectors")
	}
	return nil
}

func routeKey(source, target, capability string) string {
	return strings.TrimSpace(source) + "->" + strings.TrimSpace(target) + ":" + strings.TrimSpace(capability)
}

func clampReward(v float64) float64 {
	return math.Max(-1, math.Min(1, v))
}

func clampConfidence(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func (m *Machine) admit(exp Experience) error {
	candidate := prefrontal.Candidate{
		ID:          strings.TrimSpace(exp.ID),
		Utility:     0.5 + 0.5*clampConfidence(exp.Confidence),
		Cost:        0,
		Risk:        0,
		Uncertainty: 1 - clampConfidence(exp.Confidence),
		Urgency:     0,
		Impact:      0.2 * math.Abs(clampReward(exp.Reward)),
		Context:     map[string]any{"capability": exp.Capability, "learning": true, "event_type": string(exp.EventType)},
	}
	return m.cortex.ValidateAction(candidate)
}

func (m *Machine) persist(ctx context.Context, exp Experience) error {
	if m.store == nil {
		return nil
	}
	row := PersistedExperience{
		ID: exp.ID, TraceID: exp.TraceID, CorrelationID: exp.CorrelationID,
		Source: exp.Source, Target: exp.Target, Capability: exp.Capability,
		EventType: string(exp.EventType), Outcome: exp.Outcome, Reward: exp.Reward,
		Confidence: exp.Confidence, Input: append([]float64(nil), exp.Input...),
		TargetVector: append([]float64(nil), exp.TargetVector...), Provenance: exp.Provenance,
		Timestamp: exp.Timestamp.UTC(), Metadata: cloneMetadata(exp.Metadata),
	}
	if err := m.store.RecordLearning(ctx, row); err != nil {
		m.persistFailures.Add(1)
		return err
	}
	return nil
}

func (m *Machine) updateRoute(exp Experience) {
	key := routeKey(exp.Source, exp.Target, exp.Capability)
	reward := clampReward(exp.Reward)
	confidence := clampConfidence(exp.Confidence)
	m.mu.Lock()
	defer m.mu.Unlock()
	state := m.routes[key]
	state.Attempts++
	if reward > 0 {
		state.Successes++
	}
	state.RewardSum += reward * confidence
	state.Confidence += confidence
	if state.Attempts == 1 && state.Weight == 0 {
		state.Weight = 0.5
	}
	state.Weight = math.Max(0, math.Min(1, state.Weight+m.learningRate*reward*confidence))
	state.UpdatedAt = time.Now().UTC()
	m.routes[key] = state
}

func (m *Machine) prepare(exp Experience) Experience {
	if exp.Timestamp.IsZero() {
		exp.Timestamp = time.Now().UTC()
	}
	if strings.TrimSpace(exp.TraceID) == "" {
		exp.TraceID = exp.ID
	}
	if strings.TrimSpace(exp.CorrelationID) == "" {
		exp.CorrelationID = exp.TraceID
	}
	if strings.TrimSpace(exp.Target) == "" {
		exp.Target = "N07"
	}
	if strings.TrimSpace(exp.Outcome) == "" {
		exp.Outcome = "observed"
	}
	if strings.TrimSpace(exp.Provenance) == "" {
		exp.Provenance = "mesh-observed"
	}
	return exp
}

func (m *Machine) Learn(ctx context.Context, exp Experience) error {
	if ctx == nil {
		return errors.New("learning context is required")
	}
	exp = m.prepare(exp)
	if exp.EventType == "" {
		exp.EventType = EventSupervised
	}
	if err := validateExperience(exp); err != nil {
		return err
	}
	if err := m.admit(exp); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := m.neural.Learn(exp.Input, exp.TargetVector); err != nil {
		return err
	}
	m.updateRoute(exp)
	m.experiences.Add(1)
	m.supervised.Add(1)
	return m.persist(ctx, exp)
}

func (m *Machine) Feedback(ctx context.Context, exp Experience) error {
	if ctx == nil {
		return errors.New("learning context is required")
	}
	exp.EventType = EventFeedback
	exp = m.prepare(exp)
	if err := validateExperience(exp); err != nil {
		return err
	}
	if err := m.admit(exp); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.updateRoute(exp)
	m.cortex.ObserveOutcome(exp.Capability, exp.Reward, exp.Confidence)
	m.experiences.Add(1)
	m.feedback.Add(1)
	return m.persist(ctx, exp)
}

func (m *Machine) ObserveRoute(ctx context.Context, source, target, capability, correlation string, success bool) error {
	reward := -1.0
	outcome := "route_failure"
	if success {
		reward = 1
		outcome = "route_success"
	}
	id := correlation
	if strings.TrimSpace(id) == "" {
		id = time.Now().UTC().Format("20060102T150405.000000000Z07:00")
	}
	return m.Feedback(ctx, Experience{
		ID:            id + "-" + strings.TrimSpace(capability),
		TraceID:       correlation,
		CorrelationID: correlation,
		Source:        source,
		Target:        target,
		Capability:    capability,
		Outcome:       outcome,
		Reward:        reward,
		Confidence:    1,
		Provenance:    "mesh-observed-route",
		Metadata:      map[string]string{"learning_origin": "mesh.call_best"},
	})
}

func (m *Machine) Weight(source, target, capability string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.routes[routeKey(source, target, capability)]
	if !ok {
		return 0.5
	}
	return state.Weight
}

func (m *Machine) Routes() map[string]RouteState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]RouteState, len(m.routes))
	for key, value := range m.routes {
		out[key] = value
	}
	return out
}

func (m *Machine) Restore(ctx context.Context, limit int) error {
	if ctx == nil {
		return errors.New("learning restore context is required")
	}
	if m.store == nil {
		return nil
	}
	if limit <= 0 || limit > 10000 {
		limit = 5000
	}
	events, err := m.store.LoadLearning(ctx, limit)
	if err != nil {
		return err
	}
	for _, item := range events {
		exp := Experience{
			ID: item.ID, TraceID: item.TraceID, CorrelationID: item.CorrelationID,
			Source: item.Source, Target: item.Target, Capability: item.Capability,
			EventType: EventType(item.EventType), Outcome: item.Outcome, Reward: item.Reward,
			Confidence: item.Confidence, Input: append([]float64(nil), item.Input...),
			TargetVector: append([]float64(nil), item.TargetVector...), Provenance: item.Provenance,
			Timestamp: item.Timestamp, Metadata: cloneMetadata(item.Metadata),
		}
		if err := validateExperience(exp); err != nil {
			return err
		}
		switch exp.EventType {
		case EventSupervised:
			if err := m.neural.Learn(exp.Input, exp.TargetVector); err != nil {
				return err
			}
			m.supervised.Add(1)
		case EventFeedback:
			m.feedback.Add(1)
		default:
			continue
		}
		m.updateRoute(exp)
		m.experiences.Add(1)
		m.restored.Add(1)
	}
	return nil
}

func (m *Machine) Snapshot() Snapshot {
	return Snapshot{
		Status:          "ready",
		Experiences:     m.experiences.Load(),
		Supervised:      m.supervised.Load(),
		Feedback:        m.feedback.Load(),
		PersistFailures: m.persistFailures.Load(),
		Restored:        m.restored.Load(),
		LearnedRoutes:   len(m.Routes()),
		Routes:          m.Routes(),
	}
}

func cloneMetadata(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
