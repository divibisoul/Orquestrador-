package grce

import (
    "strings"
    "testing"

    "github.com/divibisoul/Orquestrador-/grf"
)

func TestExternalParticipantsAreFailClosedWhenRuntimeIsUnconfigured(t *testing.T) {
    for id, p := range NewExternalParticipants() {
        res, err := p.Ingest(grf.State{Payload: []byte("preserve"), Epistemic: grf.EpistemicActive}, grf.Context{CycleID:"cycle-external"})
        if err == nil || !strings.Contains(err.Error(), "EXTERNAL_RUNTIME_BLOCKED:"+id) {
            t.Fatalf("%s: expected blocked runtime, got %v", id, err)
        }
        if string(res.Output.Payload) != "preserve" { t.Fatalf("%s: input payload was not preserved", id) }
        if res.Output.Epistemic != grf.EpistemicProjected { t.Fatalf("%s: expected PROJECTED, got %s", id, res.Output.Epistemic) }
        if res.Evidence.Hash == "" || res.Provenance.ParentHash == "" { t.Fatalf("%s: missing evidence/provenance", id) }
    }
}