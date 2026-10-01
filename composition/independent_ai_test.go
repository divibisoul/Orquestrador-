package composition

import "testing"

func testPrimordialLedger() Ledger {
	return Ledger{
		Nuclei: map[string]Member{
			"N01": {Repository: "N01", Role: "core", Entrypoints: []string{"n01"}, Ownership: "native"},
			"N02": {Repository: "N02", Role: "language", Entrypoints: []string{"n02"}, Ownership: "native"},
			"N03": {Repository: "N03", Role: "perception", Entrypoints: []string{"n03"}, Ownership: "native"},
			"N04": {Repository: "N04", Role: "tools", Entrypoints: []string{"n04"}, Ownership: "native"},
			"N05": {Repository: "N05", Role: "inference", Entrypoints: []string{"n05"}, Ownership: "native"},
			"N06": {Repository: "N06", Role: "cognition", Entrypoints: []string{"n06"}, Ownership: "native"},
			"N07": {Repository: "N07", Role: "orchestration", Entrypoints: []string{"n07"}, Ownership: "system"},
		},
		Transversal: map[string]Member{
			"SARA": {Repository: "SARA", Role: "regeneration", Entrypoints: []string{"sara"}, Ownership: "transversal"},
		},
	}
}

func addEssence(ledger *Ledger, id, role, essence string) {
	if ledger.PrimordialEssence.Essences == nil {
		ledger.PrimordialEssence.Essences = map[string]Essence{}
	}
	ledger.PrimordialEssence.Essences[id] = Essence{NativeRole: role, Essence: essence, Evidence: []string{"evidence/" + id}}
}

func TestResolveCompositionPreservesPrimordialEssenceAndSupportsHigherOrderComposition(t *testing.T) {
	ledger := testPrimordialLedger()
	for id, member := range ledger.Nuclei {
		addEssence(&ledger, id, member.Role, "native-"+id)
	}
	addEssence(&ledger, "SARA", "regeneration", "native-SARA")
	ledger.PrimordialEssence.CompositionSeeds = []Seed{
		{Participants: []string{"N02", "N03", "N04"}, Mode: "higher-order", DerivedFunction: "perceive -> reason -> act", ExistingEvidence: []string{"evidence/N02-N03-N04"}, Status: string(Seeded)},
		{Participants: []string{"N01", "N05"}, Mode: "orchestrated", DerivedFunction: "context -> inference", ExistingEvidence: []string{"evidence/N01-N05"}, Status: Seeded},
		{Participants: []string{"N04", "SARA"}, Mode: "transversal", DerivedFunction: "tool execution under governance", ExistingEvidence: []string{"evidence/N04-SARA"}, Status: Seeded},
	}

	plan, err := ResolveComposition(ledger, "N03", "N02", "N04")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Status != Seeded {
		t.Fatalf("expected seeded composition, got %s", plan.Status)
	}
	if plan.DerivedFunction != "perceive -> reason -> act" {
		t.Fatalf("unexpected derived function: %q", plan.DerivedFunction)
	}
	if len(plan.Participants) != 3 || plan.Participants[0] != "N02" || plan.Participants[2] != "N04" {
		t.Fatalf("unexpected participants: %#v", plan.Participants)
	}
	if !plan.OwnershipPreserved || !plan.CorrelationRequired || !plan.ProvenanceRequired {
		t.Fatalf("composition invariants were not preserved: %#v", plan)
	}
}

func TestAIProfilesAreStructuralAndNonTransmuting(t *testing.T) {
	ledger := testPrimordialLedger()
	for id, member := range ledger.Nuclei {
		addEssence(&ledger, id, member.Role, "native-"+id)
	}
	addEssence(&ledger, "SARA", "regeneration", "native-SARA")

	profiles, err := AIProfiles(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 8 {
		t.Fatalf("expected seven nuclei plus SARA, got %d", len(profiles))
	}
	for _, profile := range profiles {
		if profile.Status != StructuralIndependent {
			t.Fatalf("expected structural profile for %s, got %s", profile.ID, profile.Status)
		}
		if !profile.NativeOwnershipPreserved {
			t.Fatalf("ownership preservation lost for %s", profile.ID)
		}
	}
}
