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
