package grce

import "testing"

func TestSOUL28TypedParticipants(t *testing.T) {
    p := NewSOUL28Participants()
    if len(p) != 3 { t.Fatalf("expected 3 typed participants, got %d", len(p)) }
    if _, ok := p[BijuxDAGRuntimeID].(*BijuxDAGParticipant); !ok { t.Fatal("bijux participant type missing") }
    if _, ok := p[OuroLoopID].(*OuroLoopParticipant); !ok { t.Fatal("ouro-loop participant type missing") }
    if _, ok := p[RecursID].(*RecursParticipant); !ok { t.Fatal("recurs participant type missing") }
}

func TestSOUL28Bindings(t *testing.T) {
    b := SOUL28Bindings()
    if len(b) != 3 { t.Fatalf("expected 3 bindings, got %d", len(b)) }
    if b[0].Role != ComplementExecutionKernel || b[1].Role != ComplementSelfHealing || b[2].Role != ComplementExperientialMemory { t.Fatal("unexpected complement role mapping") }
    for _, x := range b { if x.State != "PROJECTED" { t.Fatalf("%s must remain PROJECTED before runtime evidence", x.ID) } }
}