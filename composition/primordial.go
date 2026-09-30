package composition

import (
    "encoding/json"
    "errors"
    "os"
    "sort"
)

type Member struct {
    Repository string `json:"repository"`
    Role string `json:"role"`
    Entrypoints []string `json:"entrypoints"`
    Ownership string `json:"capabilityOwnership"`
}

type Essence struct {
    NativeRole string `json:"nativeRole"`
    Essence string `json:"essence"`
    Evidence []string `json:"evidence"`
}

type Seed struct {
    Participants []string `json:"participants"`
    Mode string `json:"mode"`
    DerivedFunction string `json:"derivedFunction"`
    ExistingEvidence []string `json:"existingEvidence"`
    Status string `json:"status"`
}

type Ledger struct {
    Nuclei map[string]Member `json:"nuclei"`
    Transversal map[string]Member `json:"transversal"`
    Fusion struct {
        Topology string `json:"topology"`
        Policy string `json:"policy"`
    } `json:"fusion"`
    PrimordialEssence struct {
        Definition string `json:"definition"`
        PreservationRule string `json:"preservationRule"`
        Essences map[string]Essence `json:"essences"`
        CompositionSeeds []Seed `json:"compositionSeeds"`
    } `json:"primordialEssence"`
}

type PairStatus string

const (
    Seeded PairStatus = "SEEDED"
    Unseeded PairStatus = "UNSEEDED"
    Invalid PairStatus = "INVALID"
)

type PairComposition struct {
    Participants [2]string `json:"participants"`
    Status PairStatus `json:"status"`
    Mode string `json:"mode,omitempty"`
    DerivedFunction string `json:"derivedFunction,omitempty"`
    NativeRoles [2]string `json:"nativeRoles"`
    EssenceSummaries [2]string `json:"essenceSummaries"`
    ExistingEvidence []string `json:"existingEvidence"`
    OwnershipPreserved bool `json:"ownershipPreserved"`
    CorrelationRequired bool `json:"correlationRequired"`
    ProvenanceRequired bool `json:"provenanceRequired"`
    NextGate string `json:"nextGate"`
}

func LoadLedger(path string) (Ledger, error) {
    raw, err := os.ReadFile(path)
    if err != nil { return Ledger{}, err }
    var ledger Ledger
    if err := json.Unmarshal(raw, &ledger); err != nil { return Ledger{}, err }
    if len(ledger.PrimordialEssence.Essences) == 0 { return Ledger{}, errors.New("primordial essence ledger is empty") }
    return ledger, nil
}

func memberExists(ledger Ledger, id string) bool {
    if _, ok := ledger.Nuclei[id]; ok { return true }
    _, ok := ledger.Transversal[id]
    return ok
}

func memberEssence(ledger Ledger, id string) (Essence, bool) {
    essence, ok := ledger.PrimordialEssence.Essences[id]
    return essence, ok
}

func seedFor(ledger Ledger, a, b string) *Seed {
    want := []string{a, b}
    sort.Strings(want)
    for i := range ledger.PrimordialEssence.CompositionSeeds {
        participants := append([]string(nil), ledger.PrimordialEssence.CompositionSeeds[i].Participants...)
        sort.Strings(participants)
        if len(participants) == 2 && participants[0] == want[0] && participants[1] == want[1] {
            return &ledger.PrimordialEssence.CompositionSeeds[i]
        }
    }
    return nil
}

func ResolvePair(ledger Ledger, a, b string) (PairComposition, error) {
    if a == b {
        return PairComposition{Participants: [2]string{a, b}, Status: Invalid, NextGate: "DISTINCT_PARTICIPANTS_REQUIRED"}, errors.New("participants must be distinct")
    }
    if !memberExists(ledger, a) || !memberExists(ledger, b) {
        return PairComposition{Participants: [2]string{a, b}, Status: Invalid, NextGate: "MEMBER_DECLARATION_REQUIRED"}, errors.New("unknown composition participant")
    }
    ea, oka := memberEssence(ledger, a)
    eb, okb := memberEssence(ledger, b)
    if !oka || !okb {
        return PairComposition{Participants: [2]string{a, b}, Status: Invalid, NextGate: "PRIMORDIAL_ESSENCE_REQUIRED"}, errors.New("participant essence missing")
    }
    plan := PairComposition{
        Participants: [2]string{a, b},
        Status: Unseeded,
        NativeRoles: [2]string{ea.NativeRole, eb.NativeRole},
        EssenceSummaries: [2]string{ea.Essence, eb.Essence},
        OwnershipPreserved: true,
        CorrelationRequired: true,
        ProvenanceRequired: true,
        NextGate: "DEFINE_AND_EXECUTE_CORRELATED_COMPOSITION",
    }
    if seed := seedFor(ledger, a, b); seed != nil {
        plan.Status = Seeded
        plan.Mode = seed.Mode
        plan.DerivedFunction = seed.DerivedFunction
        plan.ExistingEvidence = append([]string(nil), seed.ExistingEvidence...)
        plan.NextGate = "REAL_EXECUTION_TRACE"
    }
    sort.Strings(plan.ExistingEvidence)
    return plan, nil
}

func Matrix(ledger Ledger) ([]PairComposition, error) {
    ids := make([]string, 0, len(ledger.Nuclei)+len(ledger.Transversal))
    for id := range ledger.Nuclei { ids = append(ids, id) }
    for id := range ledger.Transversal { ids = append(ids, id) }
    sort.Strings(ids)
    out := make([]PairComposition, 0, len(ids)*(len(ids)-1)/2)
    for i := 0; i < len(ids); i++ {
        for j := i + 1; j < len(ids); j++ {
            plan, err := ResolvePair(ledger, ids[i], ids[j])
            if err != nil { return nil, err }
            out = append(out, plan)
        }
    }
    return out, nil
}
