package prefrontal

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

// workingMemoryEntry and TaskFrame are defined in cortex.go as the canonical
// shared runtime types. This file preserves the behavior built on them.

type OutcomeObservation struct {
	DecisionID string
	Outcome    string
	Value      float64
	At         time.Time
}

type LearningObservation struct {
	ExperienceID string
	Capability   string
	Outcome      string
	Reward       float64
	Confidence   float64
	At           time.Time
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
					delete(c.workingMemory, id)
					break
				}
			}
		}
	}
	return nil
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

func (c *Cortex) ObserveLearningOutcome(experienceID, capability, outcome string, reward, confidence float64) (LearningObservation, error) {
	if c == nil {
		return LearningObservation{}, errors.New("cortex unavailable")
	}
	experienceID = strings.TrimSpace(experienceID)
	capability = strings.TrimSpace(capability)
	outcome = strings.TrimSpace(outcome)
	if experienceID == "" || capability == "" || outcome == "" {
		return LearningObservation{}, errors.New("experience id, capability and outcome are required")
	}
	if math.IsNaN(reward) || math.IsInf(reward, 0) || reward < -1 || reward > 1 {
		return LearningObservation{}, errors.New("learning reward must be finite and within [-1,1]")
	}
	if math.IsNaN(confidence) || math.IsInf(confidence, 0) || confidence < 0 || confidence > 1 {
		return LearningObservation{}, errors.New("learning confidence must be finite and within [0,1]")
	}
	observation := LearningObservation{ExperienceID: experienceID, Capability: capability, Outcome: outcome, Reward: reward, Confidence: confidence, At: time.Now().UTC()}
	c.mu.Lock()
	c.learningObservations = append(c.learningObservations, observation)
	if len(c.learningObservations) > c.capacity {
		c.learningObservations = c.learningObservations[len(c.learningObservations)-c.capacity:]
	}
	c.mu.Unlock()
	return observation, nil
}

func (c *Cortex) LearningObservations(limit int) []LearningObservation {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if limit <= 0 || limit > len(c.learningObservations) {
		limit = len(c.learningObservations)
	}
	out := make([]LearningObservation, limit)
	copy(out, c.learningObservations[len(c.learningObservations)-limit:])
	return out
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
