package prefrontal

import (
	"fmt"
	"testing"
	"time"
)

func TestExecutiveControlFunctions(t *testing.T) {
	c, err := New(.1, 4)
	if err != nil { t.Fatal(err) }
	cands := []Candidate{
		{ID:"a", Utility:.8, Cost:.1, Risk:.1, Impact:.4},
		{ID:"b", Utility:.7, Cost:.05, Risk:.05, Urgency:.5},
	}
	if err := c.UpdateWorkingMemory(cands); err != nil { t.Fatal(err) }
	if got := len(c.WorkingMemory(0)); got != 2 { t.Fatalf("working memory size=%d",got) }
	if err := c.SwitchTask("task-1"); err != nil { t.Fatal(err) }
	if c.CurrentTask() != "task-1" { t.Fatal("task switch not recorded") }
	selected, err := c.Select(cands)
	if err != nil { t.Fatal(err) }
	decision, err := c.Commit(selected, "executive-test")
	if err != nil { t.Fatal(err) }
	if _, err := c.ObserveOutcome(decision.ID, "completed", 1); err != nil { t.Fatal(err) }
	if len(c.PendingWithinHorizon(time.Now().UTC())) != 0 { t.Fatal("completed decision cannot be pending") }
	if c.Monitor(time.Now().UTC())["status"] != "ready" { t.Fatal("monitor unhealthy") }
}

func TestPrefrontalPreservesDecisionHistoryBeyondWorkingCapacity(t *testing.T) {
	c, err := New(.01, 1)
	if err != nil {
		t.Fatal(err)
	}
	candidate := Candidate{ID: "history-action", Utility: .9, Cost: .01, Risk: .01}
	for i := 0; i < 3; i++ {
		candidate.ID = fmt.Sprintf("history-action-%d", i)
		if _, err := c.Commit(candidate, "history-test"); err != nil {
			t.Fatal(err)
		}
	}
	if got := c.HistorySize(); got != 3 {
		t.Fatalf("decision history was truncated: %d", got)
	}

	candidate.ID = "wm-a"
	if err := c.UpdateWorkingMemory([]Candidate{candidate}); err != nil {
		t.Fatal(err)
	}
	candidate.ID = "wm-b"
	if err := c.UpdateWorkingMemory([]Candidate{candidate}); err != nil {
		t.Fatal(err)
	}
	if got := len(c.WorkingMemoryHistory()); got == 0 {
		t.Fatal("working-memory overflow was discarded without archive")
	}
}
