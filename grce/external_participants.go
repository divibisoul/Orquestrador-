package grce

import (
	"context"
	"encoding/base64"
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
	BijuxDAGRuntimeID = "bijux-dag-runtime"
	OuroLoopID        = "ouro-loop"
	RecursID          = "recurs"
)

type ExternalRuntimeConfig struct {
	ID         string
	Source     string
	Revision   string
	CommandEnv string
	ArgsEnv    string
}

type ExternalParticipant struct {
	Config       ExternalRuntimeConfig
	LastEvidence grf.Evidence
	LastProv     grf.Provenance
	LastStatus   grf.EpistemicState
}

type externalRequest struct {
	Operation string
	State     string
	Context   grf.Context
}

type externalResponse struct {
	StateB64   string
	Provenance grf.Provenance
	Evidence   grf.Evidence
	Status     grf.EpistemicState
	Source     string
	Revision   string
}

func NewExternalParticipants() map[string]*ExternalParticipant {
	return map[string]*ExternalParticipant{
		BijuxDAGRuntimeID: {Config: ExternalRuntimeConfig{
			ID: BijuxDAGRuntimeID, Source: "bijux/bijux-core", Revision: "f0f5f56b49f380196918c4c6910b725b2b30ab7d",
			CommandEnv: "SOUL_BIJUX_DAG_COMMAND", ArgsEnv: "SOUL_BIJUX_DAG_ARGS",
		}},
		OuroLoopID: {Config: ExternalRuntimeConfig{
			ID: OuroLoopID, Source: "VictorVVedtion/ouro-loop", Revision: "52f97e22dc2a72a45ffa025d452ea02c58cb6b94",
			CommandEnv: "SOUL_OURO_LOOP_COMMAND", ArgsEnv: "SOUL_OURO_LOOP_ARGS",
		}},
		RecursID: {Config: ExternalRuntimeConfig{
			ID: RecursID, Source: "Gen-Verse/Recuris", Revision: "7d3745ab787b1206ccd981cb1511100476097266",
			CommandEnv: "SOUL_RECURS_COMMAND", ArgsEnv: "SOUL_RECURS_ARGS",
		}},
	}
}

func (p *ExternalParticipant) Ingest(input grf.State, c grf.Context) (grf.ParticipantResult, error) {
	if err := c.Validate(); err != nil {
		return grf.ParticipantResult{}, err
	}
	command := strings.TrimSpace(os.Getenv(p.Config.CommandEnv))
	if command == "" {
		return p.failClosed(input, c, "runtime-unconfigured", "external runtime command is not configured")
	}

	args := splitArgs(os.Getenv(p.Config.ArgsEnv))
	req := externalRequest{Operation: "grce.ingest", State: base64.StdEncoding.EncodeToString(input.Payload), Context: c}
	payload, err := json.Marshal(req)
	if err != nil {
		return p.failClosed(input, c, "request-marshal", "could not encode external request", err)
	}

	execCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(execCtx, command, args...)
	cmd.Stdin = strings.NewReader(string(payload))
	raw, err := cmd.Output()
	if err != nil {
		return p.failClosed(input, c, "runtime-exec", "external runtime command failed", err)
	}

	var resp externalResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return p.failClosed(input, c, "response-json", "external runtime returned invalid JSON", err)
	}
	if err := p.validateResponse(input, c, resp); err != nil {
		return p.failClosed(input, c, "response-contract", err.Error(), err)
	}

	decoded, err := base64.StdEncoding.DecodeString(resp.StateB64)
	if err != nil {
		return p.failClosed(input, c, "response-state", "external response state is not valid base64", err)
	}

	out := input.Clone()
	out.Payload = decoded
	out.Epistemic = resp.Status
	if out.Hash() != resp.Provenance.OutputHash {
		return p.failClosed(input, c, "response-state-hash", "external response output hash does not match decoded state")
	}
	p.LastEvidence = resp.Evidence
	p.LastStatus = resp.Status
	p.LastProv = resp.Provenance
	return grf.ParticipantResult{Output: out, Provenance: resp.Provenance, Evidence: resp.Evidence}, nil
}

func (p *ExternalParticipant) failClosed(input grf.State, c grf.Context, reason, detail string, cause ...error) (grf.ParticipantResult, error) {
	if detail == "" {
		detail = reason
	}
	if len(cause) > 0 && cause[0] != nil {
		detail = detail + ": " + cause[0].Error()
	}
	e := grf.Evidence{
		ID:          "external-" + p.Config.ID + "-" + reason,
		FailureID:   p.Config.ID + ":" + reason,
		ContextHash: c.Hash(),
		Source:      p.Config.ID,
		Detail:      detail,
	}
	e.Hash = hashExternalEvidence(e)
	out := input.Clone()
	out.Epistemic = grf.EpistemicProjected
	prov := grf.Provenance{
		ParentHash:    input.Hash(),
		InputHash:     input.Hash(),
		OutputHash:    out.Hash(),
		SequenceIndex: input.SequenceIndex + 1,
		Chain:         []string{p.Config.ID, "PROJECTED", "BLOCKED_ENV", reason},
	}
	p.LastEvidence = e
	p.LastStatus = grf.EpistemicProjected
	p.LastProv = prov
	return grf.ParticipantResult{Output: out, Provenance: prov, Evidence: e}, fmt.Errorf("EXTERNAL_RUNTIME_BLOCKED:%s:%s", p.Config.ID, reason)
}

func (p *ExternalParticipant) validateResponse(input grf.State, c grf.Context, resp externalResponse) error {
	if resp.Source != p.Config.Source {
		return fmt.Errorf("EXTERNAL_SOURCE_MISMATCH:%s", resp.Source)
	}
	if resp.Revision != p.Config.Revision {
		return fmt.Errorf("EXTERNAL_REVISION_MISMATCH:%s", resp.Revision)
	}
	if !validEpistemicState(resp.Status) {
		return fmt.Errorf("EXTERNAL_EPISTEMIC_INVALID:%s", resp.Status)
	}
	if resp.StateB64 == "" {
		return errors.New("EXTERNAL_STATE_EMPTY")
	}
	if resp.Evidence.ID == "" {
		return errors.New("EXTERNAL_EVIDENCE_ID_EMPTY")
	}
	if resp.Evidence.ContextHash != c.Hash() {
		return errors.New("EXTERNAL_EVIDENCE_CONTEXT_MISMATCH")
	}
	if resp.Evidence.Source != p.Config.ID {
		return errors.New("EXTERNAL_EVIDENCE_SOURCE_MISMATCH")
	}
	if resp.Evidence.Hash == "" || resp.Evidence.Hash != hashExternalEvidence(resp.Evidence) {
		return errors.New("EXTERNAL_EVIDENCE_HASH_MISMATCH")
	}
	if resp.Provenance.ParentHash != input.Hash() || resp.Provenance.InputHash != input.Hash() {
		return errors.New("EXTERNAL_PROVENANCE_INPUT_MISMATCH")
	}
	if resp.Provenance.OutputHash == "" || resp.Provenance.SequenceIndex != input.SequenceIndex+1 {
		return errors.New("EXTERNAL_PROVENANCE_OUTPUT_INCOMPLETE")
	}
	if len(resp.Provenance.Chain) == 0 {
		return errors.New("EXTERNAL_PROVENANCE_CHAIN_EMPTY")
	}
	return nil
}

func validEpistemicState(s grf.EpistemicState) bool {
	switch s {
	case grf.EpistemicReal, grf.EpistemicProjected, grf.EpistemicBlocked, grf.EpistemicUnresolved, grf.EpistemicPreserved, grf.EpistemicActive, grf.EpistemicIdle:
		return true
	default:
		return false
	}
}

func (p *ExternalParticipant) EpistemicState() grf.EpistemicState {
	if p.LastStatus != "" {
		return p.LastStatus
	}
	return grf.EpistemicProjected
}

func (p *ExternalParticipant) Invariants() []grf.Invariant { return grf.CanonicalInvariants }

func (p *ExternalParticipant) Capabilities() []grf.Capability {
	return []grf.Capability{{ID: p.Config.ID, Description: "External GRCE complement behind an explicit fail-closed adapter.", Genealogy: []string{p.Config.ID}, Epistemic: p.EpistemicState()}}
}

func (p *ExternalParticipant) FailuresAbsorbed() []grf.Failure {
	if p.LastEvidence.FailureID == "" {
		return nil
	}
	return []grf.Failure{{ID: p.LastEvidence.FailureID, Property: "external-runtime-boundary", Observed: p.LastEvidence.Detail, EvidenceID: p.LastEvidence.ID}}
}

func (p *ExternalParticipant) Provenance() []grf.Provenance {
	if p.LastProv.ParentHash == "" {
		return nil
	}
	return []grf.Provenance{p.LastProv}
}

func splitArgs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Fields(raw)
}

func hashExternalEvidence(e grf.Evidence) string {
	return sha256Bytes([]byte(e.ID + "|" + e.FailureID + "|" + e.ContextHash + "|" + e.Source + "|" + e.Detail))
}
