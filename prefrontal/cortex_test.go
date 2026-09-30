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

func TestScoreNormalizationIsContinuousAtOne(t *testing.T) {
	values := []float64{.999999, 1, 1.000001, 2, 10}
	prev := -1.0
	for _, value := range values {
		got := boundedPositive(value)
		if got < 0 || got >= 1 {
			t.Fatalf("boundedPositive(%g)=%g outside [0,1)", value, got)
		}
		if got < prev {
			t.Fatalf("boundedPositive is not monotonic: prev=%g current=%g", prev, got)
		}
		prev = got
	}
	if diff := math.Abs(boundedPositive(1.000001)-boundedPositive(.999999)); diff > 2e-6 {
		t.Fatalf("normalization has a discontinuity near 1: diff=%g", diff)
	}
}
