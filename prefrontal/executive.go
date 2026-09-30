package prefrontal

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

type workingMemoryEntry struct {
	Candidate Candidate
	UpdatedAt time.Time
}

type TaskFrame struct {
	TaskID      string
	ActivatedAt time.Time
}

type OutcomeObservation struct {
	DecisionID string
	Outcome    string
	Value      float64
	At         time.Time
}

func cloneCandidate(candidate Candidate) Candidate {
	copyCandidate := candidate
	if candidate.Context != nil {
		copyCandidate.Context = make(map[string]any, len(candidate.Context))
		for key, value := range candidate.Context {
			copyCandidate.Context[key] = value
		}
	}
	return copyCandidate
}

func (c *Cortex) UpdateWorkingMemory(candidates []Candidate) error {
	if c == nil {
		return errors.New("cortex unavailable")
	}
	now := time.Now().UTC()
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, candidate := range candidates {
		if err := valid(candidate); err != nil {
			return err
		}
		c.workingMemory[candidate.ID] = workingMemoryEntry{Candidate: cloneCandidate(candidate), UpdatedAt: now}
	}
	if len(c.workingMemory) > c.capacity {
		entries := make([]workingMemoryEntry, 0, len(c.workingMemory))
		for _, entry := range c.workingMemory {
			entries = append(entries, entry)
		}
		sort.SliceStable(entries, func(i, j int) bool { return entries[i].UpdatedAt.Before(entries[j].UpdatedAt) })
		for len(c.workingMemory) > c.capacity && len(entries) > 0 {
			oldest := entries[0]
			entries = entries[1:]
			for id, entry := range c.workingMemory {
				if entry.UpdatedAt.Equal(oldest.UpdatedAt) && entry.Candidate.ID == oldest.Candidate.ID {
					c.workingMemoryArchive = append(c.workingMemoryArchive, entry)
					delete(c.workingMemory, id)
					break
				}
			}
		}
	}
	return nil
}

func (c *Cortex) WorkingMemoryHistory() []Candidate {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Candidate, 0, len(c.workingMemoryArchive))
	for _, entry := range c.workingMemoryArchive {
		out = append(out, cloneCandidate(entry.Candidate))
	}
	return out
}

func (c *Cortex) WorkingMemory(limit int) []Candidate {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entries := make([]workingMemoryEntry, 0, len(c.workingMemory))
	for _, entry := range c.workingMemory {
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].UpdatedAt.After(entries[j].UpdatedAt) })
	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}
	out := make([]Candidate, 0, limit)
	for _, entry := range entries[:limit] {
		out = append(out, cloneCandidate(entry.Candidate))
	}
	return out
}

// SwitchTask provides an explicit set-shifting boundary while preserving task frames.
func (c *Cortex) SwitchTask(taskID string) error {
	if c == nil {
		return errors.New("cortex unavailable")
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return errors.New("task id is required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentTask = taskID
	c.taskFrames = append(c.taskFrames, TaskFrame{TaskID: taskID, ActivatedAt: time.Now().UTC()})
	return nil
}

func (c *Cortex) CurrentTask() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentTask
}

func (c *Cortex) ObserveOutcome(decisionID, outcome string, value float64) (OutcomeObservation, error) {
	if c == nil {
		return OutcomeObservation{}, errors.New("cortex unavailable")
	}
	decisionID = strings.TrimSpace(decisionID)
	outcome = strings.TrimSpace(outcome)
	if decisionID == "" || outcome == "" {
		return OutcomeObservation{}, errors.New("decision id and outcome are required")
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return OutcomeObservation{}, errors.New("outcome value must be finite")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.decisions)-1; i >= 0; i-- {
		if c.decisions[i].ID == decisionID {
			c.decisions[i].Outcome = outcome
			return OutcomeObservation{DecisionID: decisionID, Outcome: outcome, Value: value, At: time.Now().UTC()}, nil
		}
	}
	return OutcomeObservation{}, errors.New("decision not found")
}

func (c *Cortex) PendingWithinHorizon(now time.Time) []Decision {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	horizon := c.policy.Horizon
	if horizon <= 0 {
		return nil
	}
	out := make([]Decision, 0)
	for _, decision := range c.decisions {
		if decision.Outcome == "pending" && now.After(decision.At.Add(horizon)) {
			out = append(out, decision)
		}
	}
	return out
}

func (c *Cortex) Monitor(now time.Time) map[string]any {
	if c == nil {
		return map[string]any{"status": "degraded", "error": "cortex unavailable"}
	}
	h := c.Health()
	h["overdue_pending_decisions"] = len(c.PendingWithinHorizon(now))
	h["working_memory"] = c.WorkingMemory(0)
	return h
}
