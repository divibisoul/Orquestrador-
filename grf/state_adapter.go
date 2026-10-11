package grf

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// SOUL28State is the additive wire DTO for the earlier SOUL-28 GRF contract.
// It intentionally lives beside, rather than replacing, the canonical N07 State.
type SOUL28State struct {
	Version       uint64
	Payload       []byte
	ParentHash    string
	InputHash     string
	OutputHash    string
	SequenceIndex uint64
	Epistemic     string
}

// Clone returns an independent copy of the legacy byte payload.
func (s SOUL28State) Clone() SOUL28State {
	s.Payload = append([]byte(nil), s.Payload...)
	return s
}

// Hash reproduces the SOUL-28 state hash contract. OutputHash and Version were
// not inputs to the original hash and intentionally remain excluded.
func (s SOUL28State) Hash() string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(s.Epistemic))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(s.Payload)
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(s.ParentHash))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(s.InputHash))
	return hex.EncodeToString(hash.Sum(nil))
}

// SOUL28Context retains the earlier participant-context wire contract.
type SOUL28Context struct {
	CycleID    string
	ParentHash string
	Values     map[string]string
}

// Hash reproduces the SOUL-28 context hash, including deterministic key order.
func (c SOUL28Context) Hash() string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(c.CycleID))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(c.ParentHash))
	keys := make([]string, 0, len(c.Values))
	for key := range c.Values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(key))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(c.Values[key]))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// SOUL28Provenance is the legacy participant response provenance envelope.
type SOUL28Provenance struct {
	ParentHash    string
	InputHash     string
	OutputHash    string
	SequenceIndex uint64
	Chain         []string
}

// SOUL28Evidence is the legacy participant response evidence envelope.
type SOUL28Evidence struct {
	ID          string
	FailureID   string
	Hash        string
	ContextHash string
	Source      string
	Detail      string
}

// AdaptedSOUL28State holds both representations during an adapter boundary so
// an unchanged round trip can retain exact source bytes and legacy metadata.
type AdaptedSOUL28State struct {
	Canonical State
	legacy    SOUL28State
	payloadHash string
	epistemic EpistemicState
	parentHash string
}

// ToCanonical converts a legacy state to the canonical State while retaining
// original bytes and metadata in the adapter envelope for lossless round trips.
func (s SOUL28State) ToCanonical(id string) (AdaptedSOUL28State, error) {
	if id == "" {
		return AdaptedSOUL28State{}, errors.New("SOUL28_STATE_ID_REQUIRED")
	}
	epistemic, err := NormalizeEpistemicStrict(s.Epistemic)
	if err != nil {
		return AdaptedSOUL28State{}, err
	}
	payload, err := NewFlexiblePayloadFromBytes(s.Payload).ToMap()
	if err != nil {
		return AdaptedSOUL28State{}, err
	}
	payloadHash, err := HashJSON(payload)
	if err != nil {
		return AdaptedSOUL28State{}, fmt.Errorf("SOUL28_PAYLOAD_HASH_FAILED: %w", err)
	}
	canonical := State{
		ID:             id,
		EpistemicState: epistemic,
		Payload:        payload,
		ParentHash:     s.ParentHash,
	}
	return AdaptedSOUL28State{
		Canonical:  canonical,
		legacy:     s.Clone(),
		payloadHash: payloadHash,
		epistemic:  epistemic,
		parentHash: s.ParentHash,
	}, nil
}

// ToSOUL28 converts the canonical value back to the earlier contract. When no
// canonical field changed, it returns the original bytes and metadata exactly.
func (a AdaptedSOUL28State) ToSOUL28() (SOUL28State, error) {
	if a.Canonical.Payload == nil {
		return SOUL28State{}, errors.New("SOUL28_CANONICAL_PAYLOAD_REQUIRED")
	}
	payloadHash, err := HashJSON(a.Canonical.Payload)
	if err != nil {
		return SOUL28State{}, fmt.Errorf("SOUL28_PAYLOAD_HASH_FAILED: %w", err)
	}
	if payloadHash == a.payloadHash &&
		a.Canonical.EpistemicState == a.epistemic &&
		a.Canonical.ParentHash == a.parentHash {
		return a.legacy.Clone(), nil
	}
	payload, err := NewFlexiblePayloadFromMap(a.Canonical.Payload).ToBytes()
	if err != nil {
		return SOUL28State{}, err
	}
	epistemic, err := NormalizeEpistemicStrict(string(a.Canonical.EpistemicState))
	if err != nil {
		return SOUL28State{}, err
	}
	out := a.legacy.Clone()
	out.Payload = payload
	out.ParentHash = a.Canonical.ParentHash
	out.Epistemic = string(epistemic)
	out.OutputHash = out.Hash()
	return out, nil
}

// SOUL28StateFromCanonical creates a wire state for an explicit participant
// invocation. Missing parent lineage is rooted at the canonical payload hash;
// the generated value remains PROJECTED until the GRCE validators promote it.
func SOUL28StateFromCanonical(state State, sequence, version uint64) (SOUL28State, error) {
	if state.Payload == nil {
		return SOUL28State{}, errors.New("SOUL28_CANONICAL_PAYLOAD_REQUIRED")
	}
	epistemic, err := NormalizeEpistemicStrict(string(state.EpistemicState))
	if err != nil {
		return SOUL28State{}, err
	}
	payload, err := NewFlexiblePayloadFromMap(state.Payload).ToBytes()
	if err != nil {
		return SOUL28State{}, err
	}
	payloadHash, err := state.Hash()
	if err != nil {
		return SOUL28State{}, fmt.Errorf("SOUL28_CANONICAL_HASH_FAILED: %w", err)
	}
	parentHash := state.ParentHash
	if parentHash == "" {
		parentHash = payloadHash
	}
	if version == 0 {
		version = 1
	}
	out := SOUL28State{
		Version:       version,
		Payload:       payload,
		ParentHash:    parentHash,
		InputHash:     payloadHash,
		SequenceIndex: sequence,
		Epistemic:     string(epistemic),
	}
	out.OutputHash = out.Hash()
	return out, nil
}

// CanonicalJSON returns an independent JSON object for a wire state payload.
func (s SOUL28State) CanonicalJSON() (map[string]any, error) {
	return NewFlexiblePayloadFromBytes(s.Payload).ToMap()
}

// VerifyLegacyOutputHash checks a participant response against the old state hash.
func (s SOUL28State) VerifyLegacyOutputHash(expected string) error {
	if expected == "" || s.Hash() != expected {
		return errors.New("SOUL28_OUTPUT_HASH_MISMATCH")
	}
	return nil
}

// SOUL28StateJSONSize returns the encoded byte length, surfacing encoding errors.
func SOUL28StateJSONSize(state State) (int, error) {
	raw, err := json.Marshal(state.Payload)
	if err != nil {
		return 0, err
	}
	return len(raw), nil
}
