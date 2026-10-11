package grce

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/grf"
)

const (
	SOUL28BijuxID = "bijux-dag-runtime"
	SOUL28OuroLoopID = "ouro-loop"
	SOUL28RecursID = "recurs"
	soul28MaxRuntimeOutput = 1 << 20
)

type soul28RuntimeConfig struct {
	ID         string
	Source     string
	Revision   string
	CommandEnv string
	ArgsEnv    string
}

// SOUL28ParticipantBridge adapts the existing SOUL-28 subprocess contract to
// the canonical N07 GRF/GRCE state. It does not execute until explicitly wired
// into GRCE and configured through the component-specific command environment.
type SOUL28ParticipantBridge struct {
	configs map[string]soul28RuntimeConfig
	timeout time.Duration
}

// NewSOUL28ParticipantBridge registers the three SOUL-28 complementary runners
// with pinned identities. The executables remain caller-configured.
func NewSOUL28ParticipantBridge() *SOUL28ParticipantBridge {
	configs := []soul28RuntimeConfig{
		{ID: SOUL28BijuxID, Source: "bijux/bijux-core", Revision: "f0f5f56b49f380196918c4c6910b725b2b30ab7d", CommandEnv: "SOUL_BIJUX_DAG_COMMAND", ArgsEnv: "SOUL_BIJUX_DAG_ARGS"},
		{ID: SOUL28OuroLoopID, Source: "VictorVVedtion/ouro-loop", Revision: "52f97e22dc2a72a45ffa025d452ea02c58cb6b94", CommandEnv: "SOUL_OURO_LOOP_COMMAND", ArgsEnv: "SOUL_OURO_LOOP_ARGS"},
		{ID: SOUL28RecursID, Source: "Gen-Verse/Recuris", Revision: "7d3745ab787b1206ccd981cb1511100476097266", CommandEnv: "SOUL_RECURS_COMMAND", ArgsEnv: "SOUL_RECURS_ARGS"},
	}
	byID := make(map[string]soul28RuntimeConfig, len(configs))
	for _, config := range configs {
		byID[config.ID] = config
	}
	return &SOUL28ParticipantBridge{configs: byID, timeout: 90 * time.Second}
}

// Configured reports whether an explicit runtime command has been supplied for
// at least one participant. Partial configuration is not treated as success:
// once enabled, every participant must pass its own runtime and evidence gates.
func (b *SOUL28ParticipantBridge) Configured() bool {
	if b == nil {
		return false
	}
	for _, config := range b.configs {
		if strings.TrimSpace(os.Getenv(config.CommandEnv)) != "" {
			return true
		}
	}
	return false
}

// NewSOUL28ParticipantBridgeIfConfigured keeps the default runtime path
// unchanged when no external command was explicitly enabled.
func NewSOUL28ParticipantBridgeIfConfigured() *SOUL28ParticipantBridge {
	bridge := NewSOUL28ParticipantBridge()
	if !bridge.Configured() {
		return nil
	}
	return bridge
}

type soul28ExternalRequest struct {
	Operation      string
	State          string
	Context        grf.SOUL28Context
	StateContract  grf.SOUL28State
	ParticipantID  string
	Source         string
	Revision       string
}

type soul28ExternalResponse struct {
	StateB64   string
	Provenance grf.SOUL28Provenance
	Evidence   grf.SOUL28Evidence
	Status     string
	Source     string
	Revision   string
}

// Run invokes the three participants in deterministic order after the ETR
// gate. Any missing command, invalid response, or unverifiable evidence aborts
// the sequence and returns the original state plus all evidence observed so far.
func (b *SOUL28ParticipantBridge) Run(ctx context.Context, input grf.State, c grf.Context) (grf.State, []grf.Provenance, []grf.Evidence, error) {
	ids := []string{SOUL28BijuxID, SOUL28OuroLoopID, SOUL28RecursID}
	maxUint := ^uint64(0)
	if c.SequenceIndex > maxUint-33 {
		_, provenance, evidence, err := b.blocked(input, c, SOUL28BijuxID, "sequence-overflow", errors.New("GRCE sequence index cannot safely reserve participant stages"))
		return input, []grf.Provenance{provenance}, []grf.Evidence{evidence}, err
	}
	current := input
	provenance := make([]grf.Provenance, 0, len(ids))
	evidence := make([]grf.Evidence, 0, len(ids))
	for index, id := range ids {
		participantContext := c
		participantContext.SequenceIndex = c.SequenceIndex + 30 + uint64(index)
		next, itemProvenance, itemEvidence, err := b.InvokeParticipant(ctx, id, current, participantContext)
		provenance = append(provenance, itemProvenance)
		evidence = append(evidence, itemEvidence)
		if err != nil {
			return input, provenance, evidence, err
		}
		current = next
	}
	return current, provenance, evidence, nil
}

// InvokeParticipant performs one explicit invocation and validates the pinned
// source, revision, output hash, sequence, and evidence contract before returning
// any candidate state to the canonical GRCE pipeline.
func (b *SOUL28ParticipantBridge) InvokeParticipant(ctx context.Context, id string, input grf.State, c grf.Context) (grf.State, grf.Provenance, grf.Evidence, error) {
	if b == nil {
		return b.blocked(input, c, id, "participant-not-registered", errors.New("SOUL-28 participant bridge is nil"))
	}
	config, exists := b.configs[id]
	if !exists {
		return b.blocked(input, c, id, "participant-not-registered", errors.New("SOUL-28 participant is not registered"))
	}
	command := strings.TrimSpace(os.Getenv(config.CommandEnv))
	if command == "" {
		return b.blocked(input, c, id, "runtime-unconfigured", errors.New("external participant command is not configured"))
	}
	if ctx == nil {
		return b.blocked(input, c, id, "context-nil", errors.New("runtime context is nil"))
	}
	if err := c.Validate(); err != nil {
		return b.blocked(input, c, id, "context-invalid", err)
	}
	if c.SequenceIndex == ^uint64(0) {
		return b.blocked(input, c, id, "sequence-overflow", errors.New("sequence index cannot be incremented"))
	}

	legacyInput, err := grf.SOUL28StateFromCanonical(input, c.SequenceIndex, 1)
	if err != nil {
		return b.blocked(input, c, id, "state-adaptation", err)
	}
	legacyContext, err := soul28ContextFromCanonical(c, legacyInput.ParentHash)
	if err != nil {
		return b.blocked(input, c, id, "context-adaptation", err)
	}
	request := soul28ExternalRequest{
		Operation:     "grce.ingest",
		State:         base64.StdEncoding.EncodeToString(legacyInput.Payload),
		Context:       legacyContext,
		StateContract: legacyInput,
		ParticipantID: config.ID,
		Source:        config.Source,
		Revision:      config.Revision,
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return b.blocked(input, c, id, "request-encode", err)
	}

	timeout := b.timeout
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	executionContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(executionContext, command, strings.Fields(os.Getenv(config.ArgsEnv))...)
	cmd.Stdin = bytes.NewReader(requestBytes)
	stdout := &limitedBuffer{max: soul28MaxRuntimeOutput}
	stderr := &limitedBuffer{max: soul28MaxRuntimeOutput}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		reason := "runtime-exec"
		if executionContext.Err() == context.DeadlineExceeded {
			reason = "runtime-timeout"
		} else if stdout.exceeded || stderr.exceeded {
			reason = "runtime-output-limit"
		}
		return b.blocked(input, c, id, reason, err)
	}

	var response soul28ExternalResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return b.blocked(input, c, id, "response-json", err)
	}
	if err := validateSOUL28Response(config, legacyInput, legacyContext, response); err != nil {
		return b.blocked(input, c, id, "response-contract", err)
	}
	payloadBytes, err := base64.StdEncoding.DecodeString(response.StateB64)
	if err != nil {
		return b.blocked(input, c, id, "response-base64", err)
	}
	payload, err := grf.NewFlexiblePayloadFromBytes(payloadBytes).ToMap()
	if err != nil {
		return b.blocked(input, c, id, "response-payload", err)
	}

	// External status is preserved as evidence, not trusted as N07 authority.
	// The canonical candidate remains PROJECTED until ETR/ITR/RGO validate it.
	inputHash, hashErr := input.Hash()
	if hashErr != nil {
		return b.blocked(input, c, id, "input-hash", hashErr)
	}
	output := grf.State{
		ID:             input.ID + ":" + id,
		EpistemicState: grf.PROJECTED,
		Payload:        payload,
		ParentHash:     inputHash,
	}
	sequence := c.SequenceIndex + 1
	prov, err := grf.SealProvenance(inputHash, input, output, sequence, "SOUL28_PARTICIPANT:"+id, response.Evidence.FailureID)
	if err != nil {
		return b.blocked(input, c, id, "provenance-seal", err)
	}
	evidencePayload := map[string]any{
		"participant":       id,
		"source":            response.Source,
		"pinned_revision":   response.Revision,
		"external_state":    response.Status,
		"external_evidence": response.Evidence,
		"external_provenance": response.Provenance,
		"activation":        "PROJECTED",
	}
	evidenceHash, err := grf.HashJSON(evidencePayload)
	if err != nil {
		return b.blocked(input, c, id, "evidence-hash", err)
	}
	ev := grf.Evidence{
		ID:            "evidence:soul28:" + id + ":" + response.Evidence.FailureID,
		FailureID:     response.Evidence.FailureID,
		State:         grf.PROJECTED,
		Hash:          evidenceHash,
		InputHash:     prov.InputHash,
		OutputHash:    prov.OutputHash,
		SequenceIndex: sequence,
		Payload:       evidencePayload,
	}
	return output, prov, ev, nil
}

func (b *SOUL28ParticipantBridge) blocked(input grf.State, c grf.Context, id, reason string, cause error) (grf.State, grf.Provenance, grf.Evidence, error) {
	detail := map[string]any{
		"participant": id,
		"reason":      reason,
		"state":       "BLOCKED",
		"preserved":   true,
	}
	if cause != nil {
		detail["cause"] = cause.Error()
	}
	inputHash, hashErr := input.Hash()
	sequence := c.SequenceIndex
	if sequence < ^uint64(0) {
		sequence++
	}
	prov := grf.Provenance{SequenceIndex: sequence, Stage: "SOUL28_PARTICIPANT_BLOCKED:" + id, CausalFailureID: "soul28:" + id + ":" + reason}
	if hashErr == nil {
		prov.ParentHash = inputHash
		prov.InputHash = inputHash
		prov.OutputHash = inputHash
		detail["input_hash"] = inputHash
	} else {
		detail["input_hash_error"] = hashErr.Error()
	}
	evidenceHash, _ := grf.HashJSON(detail)
	ev := grf.Evidence{
		ID:            "evidence:soul28:blocked:" + id + ":" + reason,
		FailureID:     "soul28:" + id + ":" + reason,
		State:         grf.BLOCKED,
		Hash:          evidenceHash,
		InputHash:     inputHash,
		OutputHash:    inputHash,
		SequenceIndex: sequence,
		Payload:       detail,
	}
	return input, prov, ev, fmt.Errorf("SOUL28_RUNTIME_BLOCKED:%s:%s", id, reason)
}

func soul28ContextFromCanonical(c grf.Context, parentHash string) (grf.SOUL28Context, error) {
	if err := c.Validate(); err != nil {
		return grf.SOUL28Context{}, err
	}
	legacy := grf.SOUL28Context{CycleID: c.TraceID, ParentHash: parentHash, Values: make(map[string]string, len(c.Values))}
	for key, value := range c.Values {
		if text, ok := value.(string); ok {
			legacy.Values[key] = text
			continue
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return grf.SOUL28Context{}, fmt.Errorf("SOUL28_CONTEXT_VALUE_ENCODE_FAILED:%s: %w", key, err)
		}
		legacy.Values[key] = string(raw)
	}
	return legacy, nil
}

func validateSOUL28Response(config soul28RuntimeConfig, input grf.SOUL28State, c grf.SOUL28Context, response soul28ExternalResponse) error {
	if response.Source != config.Source {
		return fmt.Errorf("SOUL28_EXTERNAL_SOURCE_MISMATCH:%s", response.Source)
	}
	if response.Revision != config.Revision {
		return fmt.Errorf("SOUL28_EXTERNAL_REVISION_MISMATCH:%s", response.Revision)
	}
	status, err := grf.NormalizeEpistemicStrict(response.Status)
	if err != nil || string(status) != response.Status {
		return fmt.Errorf("SOUL28_EXTERNAL_EPISTEMIC_INVALID:%s", response.Status)
	}
	if response.StateB64 == "" {
		return errors.New("SOUL28_EXTERNAL_STATE_EMPTY")
	}
	if response.Evidence.ID == "" || response.Evidence.FailureID == "" {
		return errors.New("SOUL28_EXTERNAL_EVIDENCE_ID_OR_FAILURE_EMPTY")
	}
	if response.Evidence.ContextHash != c.Hash() {
		return errors.New("SOUL28_EXTERNAL_EVIDENCE_CONTEXT_MISMATCH")
	}
	if response.Evidence.Source != config.ID {
		return errors.New("SOUL28_EXTERNAL_EVIDENCE_SOURCE_MISMATCH")
	}
	if response.Evidence.Hash == "" || response.Evidence.Hash != hashSOUL28ExternalEvidence(response.Evidence) {
		return errors.New("SOUL28_EXTERNAL_EVIDENCE_HASH_MISMATCH")
	}
	inputHash := input.Hash()
	if response.Provenance.ParentHash != inputHash || response.Provenance.InputHash != inputHash {
		return errors.New("SOUL28_EXTERNAL_PROVENANCE_INPUT_MISMATCH")
	}
	if response.Provenance.OutputHash == "" || input.SequenceIndex == ^uint64(0) || response.Provenance.SequenceIndex != input.SequenceIndex+1 {
		return errors.New("SOUL28_EXTERNAL_PROVENANCE_OUTPUT_INCOMPLETE")
	}
	if len(response.Provenance.Chain) == 0 {
		return errors.New("SOUL28_EXTERNAL_PROVENANCE_CHAIN_EMPTY")
	}
	raw, err := base64.StdEncoding.DecodeString(response.StateB64)
	if err != nil {
		return fmt.Errorf("SOUL28_EXTERNAL_STATE_BASE64_INVALID: %w", err)
	}
	output := input.Clone()
	output.Payload = raw
	output.Epistemic = response.Status
	if err := output.VerifyLegacyOutputHash(response.Provenance.OutputHash); err != nil {
		return err
	}
	return nil
}

func hashSOUL28ExternalEvidence(evidence grf.SOUL28Evidence) string {
	raw := []byte(evidence.ID + "|" + evidence.FailureID + "|" + evidence.ContextHash + "|" + evidence.Source + "|" + evidence.Detail)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

type limitedBuffer struct {
	buffer   bytes.Buffer
	max      int
	exceeded bool
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	if b.max > 0 && len(data) > b.max-b.buffer.Len() {
		b.exceeded = true
		return 0, errors.New("SOUL28_RUNTIME_OUTPUT_LIMIT")
	}
	return b.buffer.Write(data)
}

func (b *limitedBuffer) Bytes() []byte {
	return b.buffer.Bytes()
}
