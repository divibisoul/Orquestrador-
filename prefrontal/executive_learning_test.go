package prefrontal

import (
	"testing"
	"time"
)

func TestExecutiveControlsCoexistWithLearningOutcomeState(t *testing.T) {
	c, err := New(.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	cands := []Candidate{
		{ID: "exec-a", Utility: .8, Cost: .1, Risk: .1, Impact: .4, Context: map[string]any{"capability": "cooperative.exec"}},
		{ID: "exec-b", Utility: .7, Cost: .05, Risk: .05, Urgency: .5},
	}
	if err := c.UpdateWorkingMemory(cands); err != nil {
		t.Fatal(err)
	}
	if got := len(c.WorkingMemory(0)); got != 2 {
		t.Fatalf("working memory size=%d", got)
	}
	if err := c.SwitchTask("task-exec"); err != nil {
		t.Fatal(err)
	}
	if c.CurrentTask() != "task-exec" {
		t.Fatal("task switch not recorded")
	}
	selected, err := c.Select(cands)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := c.Commit(selected, "executive-cooperative-test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.RecordDecisionOutcome(decision.ID, "completed", 1); err != nil {
		t.Fatal(err)
	}
	if err := c.ObserveOutcome("cooperative.exec", 1, 1); err != nil {
		t.Fatal(err)
	}
	if got, ok := c.Outcome("cooperative.exec"); !ok || got.Successes != 1 {
		t.Fatalf("learning outcome was not retained: %#v %v", got, ok)
	}
	if len(c.PendingWithinHorizon(time.Now().UTC().Add(30*time.Minute))) != 0 {
		t.Fatal("completed decision cannot remain pending")
	}
	if c.Monitor(time.Now().UTC())["status"] != "ready" {
		t.Fatal("monitor unhealthy")
	}
}
