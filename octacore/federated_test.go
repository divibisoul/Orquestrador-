package octacore

import (
	"context"
	"testing"
)

func TestBuildFederatedPreJobsCreatesG6ToG4AndG3ParallelBarrier(t *testing.T) {
	jobs := BuildFederatedPreJobs(FederatedContextInput{
		CorrelationID:     "corr-flow-1",
		ResearchPayload:   map[string]any{"query": "real research"},
		PerceptionPayload: map[string]any{"capability": "mesh.describe"},
	})
	if len(jobs) != 2 {
		t.Fatalf("expected G4 and G3 jobs, got %d", len(jobs))
	}
	for _, job := range jobs {
		if job.Source != G6 {
			t.Fatalf("expected G6 source, got %s", job.Source)
		}
		if job.ParallelGroup == nil || *job.ParallelGroup != "pre" {
			t.Fatalf("missing pre parallel group for %+v", job)
		}
		if job.Barrier == nil || *job.Barrier != "pre" {
			t.Fatalf("missing pre barrier for %+v", job)
		}
	}
	if jobs[0].Target != "G4" || jobs[1].Target != "G3" {
		t.Fatalf("unexpected pre targets: %s, %s", jobs[0].Target, jobs[1].Target)
	}
}

func TestFederatedContextCycleFailsClosedWithoutRealG4OrG3MeshEndpoints(t *testing.T) {
	processor, err := NewProcessor(DefaultSchedulerConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	result := processor.ExecuteFederatedContextCycle(
		context.Background(),
		FederatedContextInput{
			CorrelationID:     "corr-flow-2",
			Input:             "real federated cycle",
			ResearchPayload:   map[string]any{"query": "no mock"},
			PerceptionPayload: map[string]any{"capability": "mesh.describe"},
		},
	)
	if result.Cycle.OK {
		t.Fatal("cycle must not report success without real configured upstreams")
	}
	if result.Audit.OK {
		t.Fatal("audit must not report success without real SARA")
	}
}
