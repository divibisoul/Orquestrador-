package composition

import (
    "path/filepath"
    "testing"
)

func TestMatrixPreservesAllEightFunctionalMembersAndUnseededPairs(t *testing.T) {
    ledger, err := LoadLedger(filepath.Join("..", "soul-nuclei.json"))
    if err != nil { t.Fatal(err) }
    matrix, err := Matrix(ledger)
    if err != nil { t.Fatal(err) }
    if len(matrix) != 28 { t.Fatalf("expected 28 possible pairs for 8 functional members, got %d", len(matrix)) }
    seeded := 0
    unseeded := 0
    for _, pair := range matrix {
        if !pair.OwnershipPreserved || !pair.CorrelationRequired || !pair.ProvenanceRequired {
            t.Fatalf("continuity contract weakened for pair %v", pair.Participants)
        }
        switch pair.Status {
        case Seeded:
            seeded++
            if pair.DerivedFunction == "" { t.Fatalf("seeded pair %v has no derived function", pair.Participants) }
        case Unseeded:
            unseeded++
            if pair.NextGate != "DEFINE_AND_EXECUTE_CORRELATED_COMPOSITION" { t.Fatalf("bad unseeded gate for %v", pair.Participants) }
        default:
            t.Fatalf("unexpected status %q", pair.Status)
        }
    }
    binarySeeds := 0
    for _, seed := range ledger.PrimordialEssence.CompositionSeeds {
        if len(seed.Participants) == 2 {
            binarySeeds++
        }
    }
    if seeded != binarySeeds {
        t.Fatalf("expected one matrix entry for each binary seed: binary-seeds=%d matrix-seeded=%d", binarySeeds, seeded)
    }
    if unseeded == 0 { t.Fatal("expected explicit unseeded pairs to remain visible") }
}

func TestResolvePairDoesNotInventMissingComposition(t *testing.T) {
    ledger, err := LoadLedger(filepath.Join("..", "soul-nuclei.json"))
    if err != nil { t.Fatal(err) }
    ledger.PrimordialEssence.CompositionSeeds = nil
    plan, err := ResolvePair(ledger, "N02", "N03")
    if err != nil { t.Fatal(err) }
    if plan.Status != Unseeded { t.Fatalf("expected unseeded, got %s", plan.Status) }
    if plan.DerivedFunction != "" { t.Fatalf("invented derived function: %q", plan.DerivedFunction) }
}
