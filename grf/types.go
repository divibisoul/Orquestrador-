package grf

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
)

type EpistemicState string

const (
	EpistemicReal       EpistemicState = "REAL"
	EpistemicProjected  EpistemicState = "PROJECTED"
	EpistemicBlocked    EpistemicState = "BLOCKED"
	EpistemicUnresolved EpistemicState = "UNRESOLVED"
	EpistemicPreserved  EpistemicState = "PRESERVED"
	EpistemicActive     EpistemicState = "ACTIVE"
	EpistemicIdle       EpistemicState = "IDLE"
)

type State struct {
	Version       uint64
	Payload       []byte
	ParentHash    string
	InputHash     string
	OutputHash    string
	SequenceIndex uint64
	Epistemic     EpistemicState
}

func (s State) Hash() string {
	h := sha256.New()
	h.Write([]byte(s.Epistemic))
	h.Write([]byte{0})
	h.Write(s.Payload)
	h.Write([]byte{0})
	h.Write([]byte(s.ParentHash))
	h.Write([]byte{0})
	h.Write([]byte(s.InputHash))
	return hex.EncodeToString(h.Sum(nil))
}

func (s State) Size() int { return len(s.Payload) }

func (s State) Clone() State {
	s.Payload = append([]byte(nil), s.Payload...)
	return s
}

type Context struct {
	CycleID    string
	ParentHash string
	Values     map[string]string
}

func (c Context) Hash() string {
	h := sha256.New()
	h.Write([]byte(c.CycleID))
	h.Write([]byte{0})
	h.Write([]byte(c.ParentHash))
	keys := make([]string, 0, len(c.Values))
	for k := range c.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		h.Write([]byte{0})
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write([]byte(c.Values[k]))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (c Context) Validate() error {
	if c.CycleID == "" {
		return errors.New("GRF_CONTEXT_MISSING_CYCLE_ID")
	}
	return nil
}

type Evidence struct {
	ID          string
	FailureID   string
	Hash        string
	ContextHash string
	Source      string
	Detail      string
}

type Provenance struct {
	ParentHash    string
	InputHash     string
	OutputHash    string
	SequenceIndex uint64
	Chain         []string
}

type Capability struct {
	ID          string
	Description string
	Genealogy   []string
	Epistemic   EpistemicState
}

type Failure struct {
	ID         string
	Property   string
	Observed   string
	EvidenceID string
}

type Characterization struct {
	FailureID    string
	Properties   []string
	RequiredDual string
	Epistemic    EpistemicState
}

type Opposition struct {
	FailureID  string
	Property   string
	RequiredBy string
	Epistemic  EpistemicState
}

type Analysis struct {
	FailureID string
	Payload   []byte
	Source    string
	Epistemic EpistemicState
}

type Integration struct {
	FailureID  string
	Payload    []byte
	Provenance Provenance
	Epistemic  EpistemicState
}

type Transformation struct {
	FailureID  string
	State      State
	Provenance Provenance
	Epistemic  EpistemicState
}

type ParticipantResult struct {
	Output     State
	Provenance Provenance
	Evidence   Evidence
}

type Invariant struct {
	ID          string
	Description string
}

type GoldenRuleParticipant interface {
	Ingest(input State, context Context) (ParticipantResult, error)
	EpistemicState() EpistemicState
	Invariants() []Invariant
	Capabilities() []Capability
	FailuresAbsorbed() []Failure
	Provenance() []Provenance
}
