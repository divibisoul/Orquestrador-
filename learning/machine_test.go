package learning

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
)

type memoryStore struct {
	events []PersistedExperience
}

func (s *memoryStore) RecordLearning(_ context.Context, exp PersistedExperience) error {
	s.events = append(s.events, exp)
	return nil
}

func (s *memoryStore) LoadLearning(_ context.Context, _ int) ([]PersistedExperience, error) {
	return append([]PersistedExperience(nil), s.events...), nil
}

func TestMachineFeedbackFeedsPrefrontalAndRouteLearning(t *testing.T) {
	n, err := neural.New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	store := &memoryStore{}
	m, err := New(n, c, store)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Feedback(context.Background(), Experience{
		ID: "exp-1", TraceID: "trace-1", CorrelationID: "corr-1",
		Source: "N03", Target: "N07", Capability: "audio.analyze",
		Outcome: "success", Reward: 0.8, Confidence: 0.9,
		Provenance: "observed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Weight("N03", "N07", "audio.analyze"); got <= .5 {
		t.Fatalf("expected learned route weight > .5, got %v", got)
	}
	if len(c.Recall(1)) != 0 {
		t.Fatalf("feedback must not fabricate a committed action decision")
	}
	if len(store.events) != 1 || store.events[0].EventType != string(EventFeedback) {
		t.Fatalf("feedback was not persisted")
	}
	health := c.Health()
	if health["learning_observations"].(uint64) != 1 {
		t.Fatalf("expected one prefrontal learning observation")
	}
}

func TestMachineSupervisedLearningChangesNetwork(t *testing.T) {
	n, err := neural.New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(n, c, nil)
	if err != nil {
		t.Fatal(err)
	}
	before := n.Health()["learning_steps"].(uint64)
	err = m.Learn(context.Background(), Experience{
		ID: "exp-2", Source: "N05", Target: "N07", Capability: "neural.forward",
		Input: []float64{0.2, 0.3}, TargetVector: []float64{0.4, 0.5},
		Confidence: 1, Reward: 0.5, Provenance: "observed",
	})
	if err != nil {
		t.Fatal(err)
	}
	after := n.Health()["learning_steps"].(uint64)
	if after <= before {
		t.Fatalf("expected neural learning step to advance")
	}
}

func TestMachineObservesRealMeshRouteOutcome(t *testing.T) {
	n, err := neural.New(2, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(n, c, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.ObserveRoute(context.Background(), "N07", "N02", "ai.generate", "corr-route", true); err != nil {
		t.Fatal(err)
	}
	if got := m.Weight("N07", "N02", "ai.generate"); got <= .5 {
		t.Fatalf("successful route was not learned: %v", got)
	}

	if err := m.ObserveRoute(context.Background(), "N07", "N02", "ai.generate", "corr-route-2", false); err != nil {
		t.Fatal(err)
	}
	if got := m.Weight("N07", "N02", "ai.generate"); got >= .65 {
		t.Fatalf("failed route did not reduce learned weight: %v", got)
	}
}
