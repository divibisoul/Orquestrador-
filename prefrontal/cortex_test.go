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

func TestCortexTelemetryCountersBecomeEvidence(t *testing.T) {
	c, err := New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	candidate := Candidate{ID: "telemetry", Utility: .8, Cost: .1, Risk: .1}

	if _, err = c.Evaluate([]Candidate{candidate}); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Select([]Candidate{candidate}); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Commit(candidate, "telemetry-test"); err != nil {
		t.Fatal(err)
	}
	if !c.Inhibit(Candidate{ID: "blocked", Utility: .1, Risk: .9}) {
		t.Fatal("expected blocked candidate")
	}

	health := c.Health()
	if health["decision_count"] != uint64(1) {
		t.Fatalf("expected one evaluation decision, got %#v", health["decision_count"])
	}
	if health["inhibition_checks"] != uint64(3) {
		t.Fatalf("expected three inhibition checks, got %#v", health["inhibition_checks"])
	}
	if health["commits"] != uint64(1) {
		t.Fatalf("expected one commit, got %#v", health["commits"])
	}
	if health["evaluation_nanos"].(uint64) == 0 || health["commit_nanos"].(uint64) == 0 {
		t.Fatalf("expected non-zero timing evidence: %#v", health)
	}
	if rate, _ := health["inhibition_rate"].(float64); rate <= 0 || rate >= 1 {
		t.Fatalf("inhibition rate must use inhibition checks, got %v from %#v", rate, health)
	}
	if avg, _ := health["avg_decision_ms"].(float64); avg <= 0 {
		t.Fatalf("expected evaluation latency evidence in avg_decision_ms: %#v", health)
	}
	if avg, _ := health["avg_commit_ms"].(float64); avg <= 0 {
		t.Fatalf("expected commit latency evidence in avg_commit_ms: %#v", health)
	}
}
