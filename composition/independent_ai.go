package composition

import (
	"errors"
	"sort"
	"strings"
)

// IndependenceStatus distinguishes a structural nucleus profile from runtime proof.
type IndependenceStatus string

const (
	StructuralIndependent IndependenceStatus = "STRUCTURAL_INDEPENDENT"
	InsufficientEvidence  IndependenceStatus = "INSUFFICIENT_EVIDENCE"
)

// AIProfile is the additive structural profile of one native SOUL nucleus/service.
// It deliberately starts from the primordial ledger and does not replace any
// existing runtime capability, registry or ownership declaration.
type AIProfile struct {
	ID                       string             `json:"id"`
	Repository               string             `json:"repository"`
	NativeRole               string             `json:"nativeRole"`
	PrimordialEssence        string             `json:"primordialEssence"`
	Entrypoints              []string           `json:"entrypoints"`
	CapabilityOwnership      string             `json:"capabilityOwnership"`
	Evidence                 []string           `json:"evidence"`
	IndependentRuntime       bool               `json:"independentRuntime"`
	NativeOwnershipPreserved bool               `json:"nativeOwnershipPreserved"`
	Status                   IndependenceStatus `json:"status"`
	Basis                    []string           `json:"basis"`
}

// CompositionPlan generalizes the existing pair composition without removing
// PairComposition. It supports pair, triple and higher-order compositions.
type CompositionPlan struct {
	Participants        []string   `json:"participants"`
	Status              PairStatus `json:"status"`
	Mode                string     `json:"mode,omitempty"`
	DerivedFunction     string     `json:"derivedFunction,omitempty"`
	NativeRoles         []string   `json:"nativeRoles"`
	EssenceSummaries    []string   `json:"essenceSummaries"`
	ExistingEvidence    []string   `json:"existingEvidence"`
	OwnershipPreserved  bool       `json:"ownershipPreserved"`
	CorrelationRequired bool       `json:"correlationRequired"`
	ProvenanceRequired  bool       `json:"provenanceRequired"`
	IndependentMembers  bool       `json:"independentMembers"`
	NextGate            string     `json:"nextGate"`
	SeedStatus          string     `json:"seedStatus,omitempty"`
}

func ResolveAIProfile(ledger Ledger, id string) (AIProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return AIProfile{}, errors.New("composition profile participant is required")
	}

	var member Member
	if m, ok := ledger.Nuclei[id]; ok {
		member = m
	} else if m, ok := ledger.Transversal[id]; ok {
		member = m
	} else {
		return AIProfile{}, errors.New("unknown composition profile participant")
	}

	essence, ok := memberEssence(ledger, id)
	if !ok {
		return AIProfile{}, errors.New("participant primordial essence missing")
	}

	evidence := append([]string(nil), essence.Evidence...)
	sort.Strings(evidence)

	status := StructuralIndependent
	basis := []string{
		"native repository is explicitly identified",
		"native role is explicitly identified",
		"native entrypoint is explicitly identified",
		"capability ownership is explicitly identified",
		"primordial essence has source evidence",
	}

	// SARA is deliberately represented as a transversal independent
	// service profile, not as a forged N08 nucleus.
	independentRuntime := len(member.Entrypoints) > 0 && strings.TrimSpace(member.Ownership) != ""
	if !independentRuntime || len(evidence) == 0 {
		status = InsufficientEvidence
		basis = append(basis, "runtime independence cannot be established from ledger-only evidence")
	}

	if id == "SARA" {
		basis = append(basis, "transversal service; not a Soul Mesh nucleus")
	}

	return AIProfile{
		ID:                       id,
		Repository:               member.Repository,
		NativeRole:               essence.NativeRole,
		PrimordialEssence:        essence.Essence,
		Entrypoints:              append([]string(nil), member.Entrypoints...),
		CapabilityOwnership:      member.Ownership,
		Evidence:                 evidence,
		IndependentRuntime:       independentRuntime,
		NativeOwnershipPreserved: true,
		Status:                   status,
		Basis:                    basis,
	}, nil
}

func AIProfiles(ledger Ledger) ([]AIProfile, error) {
	ids := make([]string, 0, len(ledger.Nuclei)+len(ledger.Transversal))
	for id := range ledger.Nuclei {
		ids = append(ids, id)
	}
	for id := range ledger.Transversal {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	profiles := make([]AIProfile, 0, len(ids))
	for _, id := range ids {
		profile, err := ResolveAIProfile(ledger, id)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func seedForParticipants(ledger Ledger, participants []string) *Seed {
	want := append([]string(nil), participants...)
	sort.Strings(want)
	for i := range ledger.PrimordialEssence.CompositionSeeds {
		candidate := append([]string(nil), ledger.PrimordialEssence.CompositionSeeds[i].Participants...)
		sort.Strings(candidate)
		if len(candidate) == len(want) {
			match := true
			for j := range candidate {
				if candidate[j] != want[j] {
					match = false
					break
				}
			}
			if match {
				return &ledger.PrimordialEssence.CompositionSeeds[i]
			}
		}
	}
	return nil
}

// ResolveComposition preserves ResolvePair and Matrix while providing the
// missing higher-order composition primitive required by the current
// primordial-composition directive.
func ResolveComposition(ledger Ledger, participants ...string) (CompositionPlan, error) {
	normalized := make([]string, 0, len(participants))
	seen := map[string]struct{}{}

	for _, raw := range participants {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			return CompositionPlan{}, errors.New("composition participants must be distinct")
		}
		if !memberExists(ledger, id) {
			return CompositionPlan{}, errors.New("unknown composition participant: " + id)
		}
		if _, ok := memberEssence(ledger, id); !ok {
			return CompositionPlan{}, errors.New("participant primordial essence missing: " + id)
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}

	if len(normalized) < 2 {
		return CompositionPlan{}, errors.New("at least two composition participants are required")
	}

	sort.Strings(normalized)

	roles := make([]string, 0, len(normalized))
	essences := make([]string, 0, len(normalized))
	independentMembers := true
	for _, id := range normalized {
		profile, err := ResolveAIProfile(ledger, id)
		if err != nil {
			return CompositionPlan{}, err
		}
		roles = append(roles, profile.NativeRole)
		essences = append(essences, profile.PrimordialEssence)
		independentMembers = independentMembers && profile.Status == StructuralIndependent
	}

	plan := CompositionPlan{
		Participants:        normalized,
		Status:              Unseeded,
		NativeRoles:         roles,
		EssenceSummaries:    essences,
		ExistingEvidence:    nil,
		OwnershipPreserved:  true,
		CorrelationRequired: true,
		ProvenanceRequired:  true,
		IndependentMembers:  independentMembers,
		NextGate:            "DEFINE_AND_EXECUTE_CORRELATED_COMPOSITION",
	}

	if seed := seedForParticipants(ledger, normalized); seed != nil {
		plan.SeedStatus = seed.Status
		plan.Status = Seeded
		plan.Mode = seed.Mode
		plan.DerivedFunction = seed.DerivedFunction
		plan.ExistingEvidence = append([]string(nil), seed.ExistingEvidence...)
		sort.Strings(plan.ExistingEvidence)
		plan.NextGate = "REAL_EXECUTION_TRACE"
	}

	return plan, nil
}
