package grf

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type EpistemicState string

const (
	REAL EpistemicState = "REAL"
	PROJECTED EpistemicState = "PROJECTED"
	BLOCKED EpistemicState = "BLOCKED"
	UNRESOLVED EpistemicState = "UNRESOLVED"
	PRESERVED EpistemicState = "PRESERVED"
	ACTIVE EpistemicState = "ACTIVE"
	IDLE EpistemicState = "IDLE"
)

type Context struct {
	TraceID string `json:"trace_id"`
	CorrelationID string `json:"correlation_id"`
	SequenceIndex uint64 `json:"sequence_index"`
	Values map[string]any `json:"values,omitempty"`
}

func (c Context) Validate() error {
	if c.TraceID == "" { return errors.New("grf context trace_id is required") }
	if c.CorrelationID == "" { return errors.New("grf context correlation_id is required") }
	if c.SequenceIndex == 0 { return errors.New("grf context sequence_index must be non-zero") }
	return nil
}

type Provenance struct {
	ParentHash string `json:"parent_hash"`
	InputHash string `json:"input_hash"`
	OutputHash string `json:"output_hash"`
	SequenceIndex uint64 `json:"sequence_index"`
	Stage string `json:"stage"`
	CausalFailureID string `json:"causal_failure_id,omitempty"`
}

type Evidence struct {
	ID string `json:"id"`
	FailureID string `json:"failure_id"`
	State EpistemicState `json:"state"`
	Hash string `json:"hash"`
	InputHash string `json:"input_hash"`
	OutputHash string `json:"output_hash"`
	SequenceIndex uint64 `json:"sequence_index"`
	Payload map[string]any `json:"payload,omitempty"`
}

type Failure struct {
	ID string `json:"id"`
	Source string `json:"source"`
	Description string `json:"description"`
	RequiredOpposition string `json:"required_opposition,omitempty"`
	PropertyNecessary string `json:"property_necessary,omitempty"`
	PropertyDeclared bool `json:"property_declared"`
	State EpistemicState `json:"state"`
}

type Characterization struct {
	FailureID string `json:"failure_id"`
	Description string `json:"description"`
	Properties map[string]any `json:"properties,omitempty"`
	State EpistemicState `json:"state"`
}

type Opposition struct {
	FailureID string `json:"failure_id"`
	Description string `json:"description"`
	NecessaryProperty string `json:"necessary_property"`
	PropertyDeclared bool `json:"property_declared"`
	State EpistemicState `json:"state"`
}

type Artifact struct {
	ID string `json:"id"`
	Stage string `json:"stage"`
	State EpistemicState `json:"state"`
	Payload map[string]any `json:"payload,omitempty"`
	Provenance Provenance `json:"provenance"`
}

func (a Artifact) Size() (int, error) {
	raw, err := json.Marshal(a.Payload)
	if err != nil { return 0, err }
	return len(raw), nil
}

type State struct {
	ID string `json:"id"`
	EpistemicState EpistemicState `json:"epistemic_state"`
	Payload map[string]any `json:"payload,omitempty"`
	ParentHash string `json:"parent_hash,omitempty"`
}

func (s State) Hash() (string, error) { return HashJSON(s.Payload) }

func (s State) Size() (int, error) {
	raw, err := json.Marshal(s.Payload)
	if err != nil { return 0, err }
	return len(raw), nil
}

type Capability struct {
	ID string `json:"id"`
	Description string `json:"description"`
	State EpistemicState `json:"state"`
	Provenance Provenance `json:"provenance"`
}

type GoldenRuleParticipant interface {
	Ingest(input State, context Context) (State, Provenance, Evidence)
	EpistemicState() EpistemicState
	Invariants() []string
	Capabilities() []Capability
	FailuresAbsorbed() []Failure
	Provenance() []Provenance
}

func HashJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil { return "", err }
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func NewEvidence(f Failure, input State, output Artifact, sequence uint64) (Evidence, error) {
	inputHash, err := input.Hash()
	if err != nil { return Evidence{}, err }
	outputHash, err := HashJSON(output)
	if err != nil { return Evidence{}, err }
	payload := map[string]any{"failure": f, "input": input, "output": output}
	hash, err := HashJSON(payload)
	if err != nil { return Evidence{}, err }
	return Evidence{
		ID: "evidence:" + hash[:16],
		FailureID: f.ID,
		State: f.State,
		Hash: hash,
		InputHash: inputHash,
		OutputHash: outputHash,
		SequenceIndex: sequence,
		Payload: payload,
	}, nil
}

func SealProvenance(parentHash string, input any, output any, sequence uint64, stage, failureID string) (Provenance, error) {
	inputHash, err := HashJSON(input)
	if err != nil { return Provenance{}, fmt.Errorf("%s input hash: %w", stage, err) }
	outputHash, err := HashJSON(output)
	if err != nil { return Provenance{}, fmt.Errorf("%s output hash: %w", stage, err) }
	return Provenance{ParentHash: parentHash, InputHash: inputHash, OutputHash: outputHash, SequenceIndex: sequence, Stage: stage, CausalFailureID: failureID}, nil
}
