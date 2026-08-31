package prefrontal

import (
	"sort"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type memoryItem struct {
	workload trinity.Workload
	score    float64
	seen     time.Time
}

type WorkingMemory struct {
	mu    sync.RWMutex
	limit int
	items map[string]memoryItem
}

func NewWorkingMemory(limit int) *WorkingMemory {
	if limit < 1 { limit = 16 }
	return &WorkingMemory{limit: limit, items: make(map[string]memoryItem)}
}

func (m *WorkingMemory) Put(w trinity.Workload, score float64) {
	if m == nil || w.ID == "" { return }
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[w.ID] = memoryItem{workload: w, score: score, seen: time.Now()}
	for len(m.items) > m.limit {
		var victim string
		var victimValue float64
		first := true
		for id, item := range m.items {
			value := item.score + recency(item.seen)
			if first || value < victimValue { victim, victimValue, first = id, value, false }
		}
		delete(m.items, victim)
	}
}

func recency(t time.Time) float64 {
	age := time.Since(t).Seconds()
	if age <= 0 { return 1 }
	return 1 / (1 + age)
}

func (m *WorkingMemory) Get(id string) (trinity.Workload, bool) {
	if m == nil { return trinity.Workload{}, false }
	m.mu.RLock(); item, ok := m.items[id]; m.mu.RUnlock()
	return item.workload, ok
}

func (m *WorkingMemory) Snapshot() []trinity.Workload {
	if m == nil { return nil }
	m.mu.RLock(); defer m.mu.RUnlock()
	out := make([]memoryItem, 0, len(m.items))
	for _, item := range m.items { out = append(out, item) }
	sort.Slice(out, func(i, j int) bool { return out[i].seen.After(out[j].seen) })
	result := make([]trinity.Workload, 0, len(out))
	for _, item := range out { result = append(result, item.workload) }
	return result
}
