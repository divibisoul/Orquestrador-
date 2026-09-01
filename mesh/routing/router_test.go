package routing

import "testing"

func TestSelectRejectsNonPositiveCapacity(t *testing.T) {
	got := Select([]Candidate{{NodeID: "zero", Capacity: 0, LatencyMS: 1}, {NodeID: "negative", Capacity: -1, LatencyMS: 0}})
	if got != (Candidate{}) {
		t.Fatalf("selected unusable candidate: %+v", got)
	}
}

func TestSelectPrefersCapacityThenLatency(t *testing.T) {
	got := Select([]Candidate{
		{NodeID: "slow", Capacity: 2, LatencyMS: 10},
		{NodeID: "fast", Capacity: 2, LatencyMS: 5},
		{NodeID: "largest", Capacity: 3, LatencyMS: 100},
	})
	if got.NodeID != "largest" {
		t.Fatalf("selected=%q want largest", got.NodeID)
	}
}
