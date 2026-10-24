package grf

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParticipantRegistryContainsRequestedPublicContracts(t *testing.T) {
	r, err := ParticipantRegistry()
	if err != nil {
		t.Fatal(err)
	}
	// The registry is additive: keep its historical entries and assert that all
	// ten newly requested public capability sources are present, not that the
	// entire registry still has exactly ten elements.
	expected := []string{
		"bijux-dag-runtime",
		"ouro-loop",
		"recuris",
		"fedml",
		"hivemind",
		"temporal",
		"cognitive-workspace",
		"ravana",
		"ray",
		"nats-go",
	}
	for _, id := range expected {
		p, ok := r[id]
		if !ok {
			t.Errorf("requested participant missing: %s", id)
			continue
		}
		if p.EpistemicState() != PROJECTED {
			t.Errorf("new participant %s state=%s, want PROJECTED", id, p.EpistemicState())
		}
	}
	if len(r) < len(expected) {
		t.Fatalf("registry has %d participants; expected at least %d", len(r), len(expected))
	}
}

func TestHistoricalParticipantSourcesRemainPreserved(t *testing.T) {
	r, err := ParticipantRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"autogenesis", "cognifold", "belel-protocol", "opensinn-bus", "functional-graph-agi"} {
		p, ok := r[id]
		if !ok {
			t.Errorf("historical participant evidence missing: %s", id)
			continue
		}
		if p.EpistemicState() != PRESERVED {
			t.Errorf("historical participant %s state=%s, want PRESERVED", id, p.EpistemicState())
		}
	}
}

func TestBlockedParticipantNeverClaimsActiveExecution(t *testing.T) {
	// Exercise BLOCKED semantics with an explicitly blocked fixture. Historical
	// candidates in the live registry are PRESERVED evidence, not active BLOCKED
	// adapters, so production classifications are not distorted for this test.
	p, err := NewBoundaryParticipant(ParticipantDescriptor{
		ID: "blocked-test-fixture", Source: "test/unavailable-external-runtime",
		Role: "test-only blocked fixture", State: BLOCKED,
	})
	if err != nil {
		t.Fatal(err)
	}
	in := State{ID: "test", EpistemicState: PRESERVED, Payload: map[string]any{"x": 1}}
	ctx := Context{TraceID: "t", CorrelationID: "c", SequenceIndex: 1}
	out, prov, ev := p.Ingest(in, ctx)
	if out.EpistemicState != BLOCKED || ev.State != BLOCKED || prov.Stage != "PARTICIPANT_INGEST_BLOCKED" {
		t.Fatalf("unexpected blocked participant state: out=%#v prov=%#v ev=%#v", out, prov, ev)
	}
	payload := out.Payload
	if payload["external_execution"] != false {
		t.Fatalf("blocked participant claimed external execution: %#v", out.Payload)
	}
	failures := p.FailuresAbsorbed()
	if len(failures) != 1 || failures[0].State != BLOCKED {
		t.Fatalf("blocked participant failure state is not preserved as BLOCKED: %#v", failures)
	}
}

func TestGRFProvenanceSerializesCanonicalFieldNames(t *testing.T) {
	raw, err := json.Marshal(Provenance{
		ParentHash: "parent", InputHash: "input", OutputHash: "output", SequenceIndex: 7, Stage: "TEST",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, forbidden := range []string{"ParentHash", "InputHash", "OutputHash", "SequenceIndex"} {
		if strings.Contains(s, forbidden) {
			t.Fatalf("non-canonical provenance field leaked: %s", forbidden)
		}
	}
	for _, required := range []string{"parent_hash", "input_hash", "output_hash", "sequence_index"} {
		if !strings.Contains(s, required) {
			t.Fatalf("missing canonical provenance field: %s", required)
		}
	}
}
