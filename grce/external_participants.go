package grce

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
		e := grf.Evidence{
			ID:          "external-" + p.Config.ID + "-blocked",
			FailureID:   p.Config.ID + ":runtime-unconfigured",
			ContextHash: c.Hash(),
			Source:      p.Config.ID,
			Detail:      "external runtime command is not configured; preserving input and remaining PROJECTED",
		}
		e.Hash = hashExternalEvidence(e)
		p.LastEvidence = e
		p.LastStatus = grf.EpistemicProjected
		p.LastProv = grf.Provenance{
			ParentHash: input.Hash(), InputHash: input.Hash(), OutputHash: input.Hash(),
			SequenceIndex: input.SequenceIndex + 1, Chain: []string{p.Config.ID, "PROJECTED", "BLOCKED_ENV"},
		}
		out := input.Clone()
		out.Epistemic = grf.EpistemicProjected
		return grf.ParticipantResult{Output: out, Provenance: p.LastProv, Evidence: e}, errors.New("EXTERNAL_RUNTIME_BLOCKED:" + p.Config.ID)
	}

	args := splitArgs(os.Getenv(p.Config.ArgsEnv))
	req := externalRequest{Operation: "grce.ingest", State: base64.StdEncoding.EncodeToString(input.Payload), Context: c}
	payload, err := json.Marshal(req)
	if err != nil {
		return grf.ParticipantResult{}, err
	}

	execCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(execCtx, command, args...)
	cmd.Stdin = strings.NewReader(string(payload))
	raw, err := cmd.Output()
	if err != nil {
		return grf.ParticipantResult{}, err
	}

	var resp externalResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return grf.ParticipantResult{}, err
	}
	decoded, err := base64.StdEncoding.DecodeString(resp.StateB64)
	if err != nil {
		return grf.ParticipantResult{}, err
	}

	out := input.Clone()
	out.Payload = decoded
	out.Epistemic = resp.Status
	if out.Epistemic == "" {
		out.Epistemic = grf.EpistemicProjected
	}
	if resp.Provenance.ParentHash == "" {
		resp.Provenance.ParentHash = input.Hash()
	}
	if resp.Provenance.InputHash == "" {
		resp.Provenance.InputHash = input.Hash()
	}
	if resp.Provenance.OutputHash == "" {
		resp.Provenance.OutputHash = out.Hash()
	}
	if resp.Provenance.SequenceIndex == 0 {
		resp.Provenance.SequenceIndex = input.SequenceIndex + 1
	}
	if resp.Evidence.ID == "" {
		resp.Evidence.ID = p.Config.ID + "-external-evidence"
	}
	p.LastEvidence = resp.Evidence
	p.LastStatus = resp.Status
	if p.LastStatus == "" {
		p.LastStatus = grf.EpistemicProjected
	}
	p.LastProv = resp.Provenance
	return grf.ParticipantResult{Output: out, Provenance: resp.Provenance, Evidence: resp.Evidence}, nil
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
