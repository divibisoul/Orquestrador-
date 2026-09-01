package routing

import "sort"

type Candidate struct {
	NodeID    string
	Capacity  float64
	LatencyMS float64
}

// Select chooses the highest-capacity candidate among nodes with positive capacity;
// latency breaks ties so an unusable capacity can never become the selected route.
func Select(nodes []Candidate) Candidate {
	eligible := make([]Candidate, 0, len(nodes))
	for _, node := range nodes {
		if node.Capacity > 0 {
			eligible = append(eligible, node)
		}
	}
	if len(eligible) == 0 {
		return Candidate{}
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Capacity == eligible[j].Capacity {
			return eligible[i].LatencyMS < eligible[j].LatencyMS
		}
		return eligible[i].Capacity > eligible[j].Capacity
	})
	return eligible[0]
}
