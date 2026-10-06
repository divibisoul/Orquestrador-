package grce

import "testing"

func TestSOUL28ComplementTopology(t *testing.T) {
    placements := SOUL28ComplementPlacements()
    if len(placements) != 3 { t.Fatalf("expected 3 complements, got %d", len(placements)) }
    if err := ValidateComplementTopology(); err != nil { t.Fatal(err) }
}