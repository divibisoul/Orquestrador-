package cognitive

import "testing"

func TestExternalPlanningSourcesHaveStableProvenance(t *testing.T) {
	if len(ExternalPlanningSources) < 4 {
		t.Fatalf("expected multiple planning sources, got %d", len(ExternalPlanningSources))
	}
	seen := make(map[string]struct{}, len(ExternalPlanningSources))
	for _, source := range ExternalPlanningSources {
		if source.ID == "" || source.Repository == "" || source.SourcePath == "" || source.SourceRevision == "" {
			t.Fatalf("incomplete provenance for %+v", source)
		}
		if _, exists := seen[source.ID]; exists {
			t.Fatalf("duplicate planning source ID %q", source.ID)
		}
		seen[source.ID] = struct{}{}
	}
}

func TestExternalPlanningSourceIDsReturnsStableOrder(t *testing.T) {
	ids := ExternalPlanningSourceIDs()
	if len(ids) != len(ExternalPlanningSources) {
		t.Fatalf("IDs length mismatch: %d != %d", len(ids), len(ExternalPlanningSources))
	}
	for i, id := range ids {
		if id != ExternalPlanningSources[i].ID {
			t.Fatalf("ID order mismatch at %d: %q != %q", i, id, ExternalPlanningSources[i].ID)
		}
	}
}
