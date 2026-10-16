package grf

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestFlexiblePayloadPreservesRawBytesAndConvertsObject(t *testing.T) {
	raw := []byte(" { \"text\" : \"hello\", \"n\" : 2 } \n")
	fromBytes := NewFlexiblePayloadFromBytes(raw)
	roundTrip, err := fromBytes.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(roundTrip) != string(raw) {
		t.Fatalf("raw bytes changed: %q != %q", roundTrip, raw)
	}
	object, err := fromBytes.ToMap()
	if err != nil {
		t.Fatal(err)
	}
	if object["text"] != "hello" {
		t.Fatalf("text=%v", object["text"])
	}
	fromMap := NewFlexiblePayloadFromMap(map[string]any{"text": "hello"})
	encoded, err := fromMap.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil || decoded["text"] != "hello" {
		t.Fatalf("map conversion failed: %s (%v)", encoded, err)
	}
}

func TestFlexiblePayloadRejectsInvalidOrNonObjectJSON(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte("not-json"), []byte("[]"), []byte("null"), []byte("{} {}")} {
		if _, err := NewFlexiblePayloadFromBytes(raw).ToMap(); err == nil {
			t.Errorf("expected payload to be rejected: %q", raw)
		}
	}
	if !NewFlexiblePayloadFromMap(map[string]any{}).IsEmpty() {
		t.Fatal("empty map should be empty")
	}
	if !NewFlexiblePayloadFromBytes(nil).IsEmpty() {
		t.Fatal("nil bytes should be empty")
	}
}

func TestEpistemicNormalizationIsCaseInsensitiveAndFailClosed(t *testing.T) {
	for _, input := range []string{"REAL", "real", "Real"} {
		state, err := NormalizeEpistemicStrict(input)
		if err != nil || state != REAL {
			t.Fatalf("%q => %q, %v", input, state, err)
		}
	}
	if NormalizeEpistemic("unknown") != UNRESOLVED {
		t.Fatal("unknown state must normalize to UNRESOLVED")
	}
	if _, err := NormalizeEpistemicStrict("unknown"); err == nil {
		t.Fatal("strict normalization must reject an unknown state")
	}
}

func TestSOUL28AdapterLosslessRoundTripPreservesBytesAndMetadata(t *testing.T) {
	raw := []byte(" { \"text\" : \"hello\" } \n")
	legacy := SOUL28State{
		Version:       7,
		Payload:       raw,
		ParentHash:    "parent-original",
		InputHash:     "input-original",
		OutputHash:    "output-original",
		SequenceIndex: 17,
		Epistemic:     "pRoJeCtEd",
	}
	adapted, err := legacy.ToCanonical("state-test")
	if err != nil {
		t.Fatal(err)
	}
	if adapted.Canonical.EpistemicState != PROJECTED {
		t.Fatalf("epistemic=%s", adapted.Canonical.EpistemicState)
	}
	back, err := adapted.ToSOUL28()
	if err != nil {
		t.Fatal(err)
	}
	if string(back.Payload) != string(legacy.Payload) {
		t.Fatalf("payload bytes changed: %q != %q", back.Payload, legacy.Payload)
	}
	if back.Version != legacy.Version || back.ParentHash != legacy.ParentHash ||
		back.InputHash != legacy.InputHash || back.OutputHash != legacy.OutputHash ||
		back.SequenceIndex != legacy.SequenceIndex || back.Epistemic != legacy.Epistemic {
		t.Fatalf("legacy metadata changed: %#v != %#v", back, legacy)
	}
	if !reflect.DeepEqual(back.Payload, legacy.Payload) {
		t.Fatal("payload slice bytes must be preserved")
	}
}

func TestSOUL28AdapterSerializesModifiedCanonicalStateWithoutDroppingLineage(t *testing.T) {
	legacy := SOUL28State{
		Version: 2, Payload: []byte("{\"text\":\"before\"}"),
		ParentHash: "parent", InputHash: "input", OutputHash: "old-output",
		SequenceIndex: 4, Epistemic: "PROJECTED",
	}
	adapted, err := legacy.ToCanonical("state-modified")
	if err != nil {
		t.Fatal(err)
	}
	adapted.Canonical.Payload["text"] = "after"
	back, err := adapted.ToSOUL28()
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(back.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["text"] != "after" {
		t.Fatalf("payload=%v", payload)
	}
	if back.ParentHash != legacy.ParentHash || back.InputHash != legacy.InputHash ||
		back.Version != legacy.Version || back.SequenceIndex != legacy.SequenceIndex {
		t.Fatal("metadata lineage was not preserved")
	}
	if back.OutputHash == legacy.OutputHash || back.OutputHash != back.Hash() {
		t.Fatal("updated output hash was not refreshed")
	}
	if strings.TrimSpace(string(back.Payload)) == string(legacy.Payload) {
		t.Fatal("modified payload should be encoded from canonical form")
	}
}

func TestSOUL28AdapterRejectsInvalidPayloadAndUnknownEpistemicState(t *testing.T) {
	if _, err := (SOUL28State{Payload: []byte("not-json"), Epistemic: "PROJECTED"}).ToCanonical("invalid"); err == nil {
		t.Fatal("invalid JSON must fail closed")
	}
	if _, err := (SOUL28State{Payload: []byte("{}"), Epistemic: "SURPRISE"}).ToCanonical("invalid-state"); err == nil {
		t.Fatal("unknown epistemic state must fail closed")
	}
}

func TestSOUL28ContextHashIsDeterministic(t *testing.T) {
	left := SOUL28Context{CycleID: "cycle", ParentHash: "parent", Values: map[string]string{"b": "2", "a": "1"}}
	right := SOUL28Context{CycleID: "cycle", ParentHash: "parent", Values: map[string]string{"a": "1", "b": "2"}}
	if left.Hash() != right.Hash() {
		t.Fatal("context hash depends on map insertion order")
	}
}
