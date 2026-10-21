package prefrontal

import "testing"

func TestCortexRuntime(t *testing.T) {
	c, err := New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	in := []Candidate{{ID: "a", Utility: .8, Cost: .1, Risk: .1}, {ID: "b", Utility: .7, Cost: .05, Risk: .05}}
	if _, err = c.Evaluate(in); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Plan(in); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Prioritize(in); err != nil {
		t.Fatal(err)
	}
	if c.Inhibit(Candidate{ID: "x", Utility: .1, Risk: .9}) != true {
		t.Fatal("risk policy failed")
	}
	v, err := c.Select(in)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.ValidateAction(v); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Commit(v, "test"); err != nil {
		t.Fatal(err)
	}
	if len(c.Recall(1)) != 1 {
		t.Fatal("recall failed")
	}
	if c.Health()["status"] != "ready" {
		t.Fatal("cortex unhealthy")
	}
}

func TestCortexPreservesInvalidCandidateEvaluationEvidence(t *testing.T) {
	c, err := New(.1, 2)
	if err != nil {
		t.Fatal(err)
	}
	validCandidate := Candidate{ID: "valid", Utility: .9, Cost: .1, Risk: .1}
	invalidCandidate := Candidate{ID: "invalid", Utility: .8, Cost: -1, Risk: .1}

	if _, err := c.Evaluate([]Candidate{invalidCandidate, validCandidate}); err != nil {
		t.Fatal(err)
	}
	issues := c.EvaluationIssues(0)
	if len(issues) != 1 {
		t.Fatalf("expected one retained evaluation issue, got %d", len(issues))
	}
	if issues[0].CandidateID != "invalid" || issues[0].Reason != "cost, risk, uncertainty, urgency and impact cannot be negative" {
		t.Fatalf("invalid candidate evidence was not preserved: %+v", issues[0])
	}
	health := c.Health()
	if health["evaluation_issue_count"] != uint64(1) || health["evaluation_issue_retained"] != 1 {
		t.Fatalf("evaluation issue telemetry missing: %#v", health)
	}
}
