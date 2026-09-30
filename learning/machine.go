package learning

import (
	"context"
	"errors"
	"fmt"
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
	ID            string `json:"id"`
	TraceID       string `json:"trace_id"`
	CorrelationID string `json:"correlation_id"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Capability    string `json:"capability"`
	EventType     EventType
	Outcome       string `json:"outcome"`
	Reward        float64 `json:"reward"`
	Confidence    float64 `json:"confidence"`
	Input         []float64 `json:"input"`
	TargetVector  []float64 `json:"target_vector"`
	Provenance    string `json:"provenance"`
	Timestamp     time.Time `json:"created_at"`
	Metadata      map[string]string `json:"metadata"`
}

type PersistedExperience struct {
	ID            string `json:"id"`
	TraceID       string `json:"trace_id"`
	CorrelationID string `json:"correlation_id"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Capability    string `json:"capability"`
	EventType     string `json:"event_type"`
	Outcome       string `json:"outcome"`
	Reward        float64 `json:"reward"`
	Confidence    float64 `json:"confidence"`
	Input         []float64 `json:"input"`
	TargetVector  []float64 `json:"target_vector"`
	Provenance    string `json:"provenance"`
	Timestamp     time.Time `json:"created_at"`
	Metadata      map[string]string `json:"metadata"`
}

type Store interface {
	RecordLearning(context.Context, PersistedExperience) error
	LoadLearning(context.Context, int) ([]PersistedExperience, error)
}

type RouteState struct {
	Attempts    uint64    `json:"attempts"`
	Successes   uint64    `json:"successes"`
	RewardSum   float64   `json:"reward_sum"`
	Confidence  float64   `json:"confidence_sum"`
	Weight      float64   `json:"weight"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Snapshot struct {
	Status          string
	Experiences     uint64
	Supervised      uint64
	Feedback        uint64
	PersistFailures uint64
	Restored        uint64
	LearnedRoutes   int
	Neural          map[string]any
	Prefrontal      map[string]any
}

type Machine struct {
	mu               sync.RWMutex
	neural           *neural.Network
	cortex           *prefrontal.Cortex
	store            Store
	learningRate     float64
	routes           map[string]RouteState
	experiences      atomic.Uint64
	supervised       atomic.Uint64
	feedback         atomic.Uint64
	persistFailures  atomic.Uint64
	restored         atomic.Uint64
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

func clampReward(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Max(-1, math.Min(1, value))
}

func clampConfidence(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Max(0, math.Min(1, value))
}

func routeKey(source, target, capability string) string {
	return strings.TrimSpace(source) + "->" + strings.TrimSpace(target) + ":" + strings.TrimSpace(capability)
}

func (m *Machine) admit(exp Experience) error {
	confidence := clampConfidence(exp.Confidence)
	reward := clampReward(exp.Reward)
	candidate := prefrontal.Candidate{
		ID:          strings.TrimSpace(exp.ID),
		Utility:     0.5 + 0.5*confidence,
		Cost:        0.01,
		Risk:        0,
		Uncertainty: 1 - confidence,
		Urgency:     0,
		Impact:      math.Abs(reward) * 0.2,
		Context: map[string]any{
			"capability": strings.TrimSpace(exp.Capability),
			"learning":   true,
			"event_type": string(exp.EventType),
		},
	}
	if candidate.ID == "" {
		return errors.New("learning experience id is required")
	}
	return m.cortex.ValidateAction(candidate)
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
	for _, v := range append(append([]float64{}, exp.Input...), exp.TargetVector...) {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("learning vectors must contain only finite values")
		}
	}
	if exp.EventType == EventSupervised {
		if len(exp.Input) == 0 || len(exp.Input) != len(exp.TargetVector) {
			return errors.New("supervised learning requires equal non-empty input and target vectors")
		}
	}
	return nil
}

func (m *Machine) persist(ctx context.Context, exp Experience) error {
	if m.store == nil {
		return nil
	}
	row := PersistedExperience{
		ID: exp.ID, TraceID: exp.TraceID, CorrelationID: exp.CorrelationID,
		Source: exp.Source, Target: exp.Target, Capability: exp.Capability,
		EventType: string(exp.EventType), Outcome: exp.Outcome,
		Reward: exp.Reward, Confidence: exp.Confidence,
		Input: append([]float64(nil), exp.Input...),
		TargetVector: append([]float64(nil), exp.TargetVector...),
		Provenance: exp.Provenance, Timestamp: exp.Timestamp.UTC(),
		Metadata: cloneMetadata(exp.Metadata),
	}
	if err := m.store.RecordLearning(ctx, row); err != nil {
		m.persistFailures.Add(1)
		return fmt.Errorf("persist learning experience: %w", err)
	}
	return nil
}

func (m *Machine) updateRoute(exp Experience) {
	key := routeKey(exp.Source, exp.Target, exp.Capability)
	reward := clampReward(exp.Reward)
	conf := clampConfidence(exp.Confidence)

	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.routes[key]
	state.Attempts++
	if reward > 0 {
		state.Successes++
	}
	state.RewardSum += reward * conf
	state.Confidence += conf
	current := state.Weight
	if state.Attempts == 1 && current == 0 {
		current = 0.5
	}
	state.Weight = math.Max(0, math.Min(1, current+m.learningRate*reward*conf))
	state.UpdatedAt = time.Now().UTC()
	m.routes[key] = state
}

func (m *Machine) Learn(ctx context.Context, exp Experience) error {
	if ctx == nil {
		return errors.New("learning context is required")
	}
	if exp.EventType == "" {
		exp.EventType = EventSupervised
	}
	if exp.Timestamp.IsZero() {
		exp.Timestamp = time.Now().UTC()
	}
	if exp.TraceID == "" {
		exp.TraceID = exp.ID
	}
	if exp.CorrelationID == "" {
		exp.CorrelationID = exp.TraceID
	}
	if err := validateExperience(exp); err != nil {
		return err
	}
	if err := m.admit(exp); err != nil {
		return fmt.Errorf("prefrontal learning admission rejected: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := m.neural.Learn(exp.Input, exp.TargetVector); err != nil {
		return fmt.Errorf("neural learning failed: %w", err)
	}
	m.updateRoute(exp)
	m.cortex.ObserveOutcome(exp.Capability, exp.Reward, exp.Confidence)
	m.experiences.Add(1)
	m.supervised.Add(1)

	return m.persist(ctx, exp)
}

func (m *Machine) Feedback(ctx context.Context, exp Experience) error {
	if ctx == nil {
		return errors.New("learning context is required")
	}
	exp.EventType = EventFeedback
	if exp.Timestamp.IsZero() {
		exp.Timestamp = time.Now().UTC()
	}
	if exp.TraceID == "" {
		exp.TraceID = exp.ID
	}
	if exp.CorrelationID == "" {
		exp.CorrelationID = exp.TraceID
	}
	if err := validateExperience(exp); err != nil {
		return err
	}
	if err := m.admit(exp); err != nil {
		return fmt.Errorf("prefrontal feedback admission rejected: %w", err)
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
	if limit <= 0 {
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
			EventType: EventType(item.EventType), Outcome: item.Outcome,
			Reward: item.Reward, Confidence: item.Confidence,
			Input: append([]float64(nil), item.Input...),
			TargetVector: append([]float64(nil), item.TargetVector...),
			Provenance: item.Provenance, Timestamp: item.Timestamp,
			Metadata: cloneMetadata(item.Metadata),
		}
		switch exp.EventType {
		case EventSupervised:
			if err := validateExperience(exp); err != nil {
				return fmt.Errorf("invalid persisted supervised experience %s: %w", exp.ID, err)
			}
			if err := m.neural.Learn(exp.Input, exp.TargetVector); err != nil {
				return fmt.Errorf("failed replay of supervised experience %s: %w", exp.ID, err)
			}
			m.updateRoute(exp)
			m.cortex.ObserveOutcome(exp.Capability, exp.Reward, exp.Confidence)
			m.supervised.Add(1)
		case EventFeedback:
			if err := validateExperience(exp); err != nil {
				return fmt.Errorf("invalid persisted feedback experience %s: %w", exp.ID, err)
			}
			m.updateRoute(exp)
			m.cortex.ObserveOutcome(exp.Capability, exp.Reward, exp.Confidence)
			m.feedback.Add(1)
		default:
			continue
		}
		m.experiences.Add(1)
		m.restored.Add(1)
	}
	return nil
}

func (m *Machine) Snapshot() Snapshot {
	m.mu.RLock()
	routeCount := len(m.routes)
	m.mu.RUnlock()
	return Snapshot{
		Status:          "ready",
		Experiences:     m.experiences.Load(),
		Supervised:      m.supervised.Load(),
		Feedback:        m.feedback.Load(),
		PersistFailures: m.persistFailures.Load(),
		Restored:        m.restored.Load(),
		LearnedRoutes:   routeCount,
		Neural:          m.neural.Health(),
		Prefrontal:      m.cortex.Health(),
	}
}

func cloneMetadata(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
