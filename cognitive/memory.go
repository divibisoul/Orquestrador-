package cognitive

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type WorkingMemoryItem struct {
	Key        string         `json:"key"`
	Value      map[string]any `json:"value"`
	Relevance  float64        `json:"relevance"`
	ObservedAt time.Time      `json:"observed_at"`
	ExpiresAt  time.Time      `json:"expires_at"`
}

type WorkingMemory struct {
	mu    sync.Mutex
	items map[string]WorkingMemoryItem
	cfg   Config
}

func NewWorkingMemory(cfg Config) *WorkingMemory {
	if cfg.WorkingMemoryItems <= 0 {
		cfg.WorkingMemoryItems = 64
	}
	if cfg.WorkingMemoryTTL <= 0 {
		cfg.WorkingMemoryTTL = 15 * time.Minute
	}
	return &WorkingMemory{items: map[string]WorkingMemoryItem{}, cfg: cfg}
}

func (m *WorkingMemory) Put(key string, value map[string]any, relevance float64) error {
	if m == nil {
		return errors.New("working memory unavailable")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("working memory key is required")
	}
	if value == nil {
		return errors.New("working memory value is required")
	}
	if relevance < 0 || relevance > 1 {
		return errors.New("working memory relevance must be between 0 and 1")
	}
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evict(now)
	m.items[key] = WorkingMemoryItem{
		Key: key, Value: cloneMap(value), Relevance: relevance,
		ObservedAt: now, ExpiresAt: now.Add(m.cfg.WorkingMemoryTTL),
	}
	for len(m.items) > m.cfg.WorkingMemoryItems {
		victim := ""
		score := 2.0
		for k, v := range m.items {
			if v.Relevance < score {
				victim, score = k, v.Relevance
			}
		}
		if victim == "" {
			break
		}
		delete(m.items, victim)
	}
	return nil
}

func (m *WorkingMemory) Get(key string) (WorkingMemoryItem, bool) {
	if m == nil {
		return WorkingMemoryItem{}, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evict(time.Now().UTC())
	v, ok := m.items[strings.TrimSpace(key)]
	if !ok {
		return WorkingMemoryItem{}, false
	}
	v.Value = cloneMap(v.Value)
	return v, true
}

func (m *WorkingMemory) Snapshot() []WorkingMemoryItem {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evict(time.Now().UTC())
	out := make([]WorkingMemoryItem, 0, len(m.items))
	for _, v := range m.items {
		v.Value = cloneMap(v.Value)
		out = append(out, v)
	}
	return out
}

func (m *WorkingMemory) evict(now time.Time) {
	for k, v := range m.items {
		if !v.ExpiresAt.After(now) {
			delete(m.items, k)
		}
	}
}

type GoalStore struct {
	mu    sync.RWMutex
	goals map[string]Goal
}

func NewGoalStore() *GoalStore {
	return &GoalStore{goals: map[string]Goal{}}
}

func (s *GoalStore) Put(g Goal) error {
	if s == nil {
		return errors.New("goal store unavailable")
	}
	if strings.TrimSpace(g.ID) == "" {
		return errors.New("goal id is required")
	}
	if strings.TrimSpace(g.Objective) == "" {
		return errors.New("goal objective is required")
	}
	if strings.TrimSpace(g.CorrelationID) == "" {
		return errors.New("goal correlation_id is required")
	}
	if g.ExpiresAt.IsZero() {
		return errors.New("goal expiry is required")
	}
	if !g.ExpiresAt.After(time.Now().UTC()) {
		return errors.New("goal expiry must be in the future")
	}
	for name, value := range map[string]float64{
		"risk": g.Risk, "cost": g.Cost, "urgency": g.Urgency, "impact": g.Impact,
	} {
		if value < 0 || value != value {
			return errors.New("goal " + name + " must be finite and non-negative")
		}
	}
	if g.Risk > 1 {
		return errors.New("goal risk must be <= 1")
	}
	s.mu.Lock()
	s.goals[g.ID] = g
	s.mu.Unlock()
	return nil
}

func (s *GoalStore) Get(id string) (Goal, bool) {
	if s == nil {
		return Goal{}, false
	}
	id = strings.TrimSpace(id)
	s.mu.RLock()
	g, ok := s.goals[id]
	s.mu.RUnlock()
	if ok && !g.ExpiresAt.After(time.Now().UTC()) {
		s.mu.Lock()
		delete(s.goals, id)
		s.mu.Unlock()
		return Goal{}, false
	}
	return g, ok
}

func cloneMap(v map[string]any) map[string]any {
	if v == nil {
		return nil
	}
	out := make(map[string]any, len(v))
	for k, x := range v {
		out[k] = x
	}
	return out
}
