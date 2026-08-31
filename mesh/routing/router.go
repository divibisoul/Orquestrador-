package routing

import "sort"
type Candidate struct{NodeID string; Capacity,LatencyMS float64}
// Select chooses the lowest-latency candidate among nodes with positive capacity.
func Select(nodes []Candidate)Candidate{xs:=append([]Candidate(nil),nodes...);sort.Slice(xs,func(i,j int)bool{if xs[i].Capacity==xs[j].Capacity{return xs[i].LatencyMS<xs[j].LatencyMS};return xs[i].Capacity>xs[j].Capacity});if len(xs)==0{return Candidate{}};return xs[0]}
