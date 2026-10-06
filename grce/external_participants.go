package grce

import (
	"context"
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
	Operation string     `json:"operation"`
	State     grf.State  `json:"state"`
	Context   grf.Context `json:"context"`
}

type externalResponse struct {
	State      grf.State       `json:"state"`
	Provenance grf.Provenance  `json:"provenance"`
	Evidence   grf.Evidence    `json:"evidence"`
	Status     grf.EpistemicState `json:"status"`
	Source     string          `json:"source"`
	Revision   string          `json:"revision"`
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

func (p *ExternalParticipant) Ingest(input grf.State, c grf.Context) (grf.State, grf.Provenance, grf.Evidence) {
	if err := c.Validate(); err != nil {
		return input, grf.Provenance{ParentHash: safeStateHash(input), SequenceIndex: c.SequenceIndex, Stage: "PARTICIPANT_INGEST_INVALID_CONTEXT"}, grf.Evidence{
			ID: "evidence:" + p.Config.ID + ":invalid-context",
			FailureID: p.Config.ID + ":invalid-context", State: grf.PRESERVED, SequenceIndex: c.SequenceIndex,
			Payload: map[string]any{"error": err.Error()},
		}
	}
	command := strings.TrimSpace(os.Getenv(p.Config.CommandEnv))
	if command == "" {
		return p.failClosed(input, c, "runtime-unconfigured", "external runtime command is not configured")
	}

	payload, err := json.Marshal(externalRequest{Operation: "grce.ingest", State: input, Context: c})
	if err != nil {
		return p.failClosed(input, c, "request-marshal", "could not encode external request: "+err.Error())
	}

	execCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(execCtx, command, splitArgs(os.Getenv(p.Config.ArgsEnv))...)
	cmd.Stdin = strings.NewReader(string(payload))
	raw, err := cmd.Output()
	if err != nil {
		return p.failClosed(input, c, "runtime-exec", "external runtime command failed: "+err.Error())
	}

	var resp externalResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return p.failClosed(input, c, "response-json", "external runtime returned invalid JSON: "+err.Error())
	}
	if err := p.validateResponse(input, c, resp); err != nil {
		return p.failClosed(input, c, "response-contract", err.Error())
	}

	p.LastEvidence = resp.Evidence
	p.LastStatus = resp.Status
	p.LastProv = resp.Provenance
	return resp.State, resp.Provenance, resp.Evidence
}

func (p *ExternalParticipant) failClosed(input grf.State, c grf.Context, reason, detail string) (grf.State, grf.Provenance, grf.Evidence) {
	parent := safeStateHash(input)
	output := input
	output.EpistemicState = grf.PROJECTED
	output.ParentHash = parent
	prov, err := grf.SealProvenance(parent, input, output, c.SequenceIndex+1, "EXTERNAL_"+p.Config.ID+"_"+strings.ToUpper(reason), p.Config.ID+":"+reason)
	if err != nil {
		prov = grf.Provenance{ParentHash: parent, InputHash: parent, OutputHash: parent, SequenceIndex: c.SequenceIndex + 1, Stage: "EXTERNAL_"+p.Config.ID+"_"+strings.ToUpper(reason), CausalFailureID: p.Config.ID+":"+reason}
	}
	ev := grf.Evidence{
		ID: "evidence:" + p.Config.ID + ":" + reason,
		FailureID: p.Config.ID + ":" + reason,
		State: grf.BLOCKED,
		Hash: prov.OutputHash,
		InputHash: prov.InputHash,
		OutputHash: prov.OutputHash,
		SequenceIndex: c.SequenceIndex + 1,
		Payload: map[string]any{"source": p.Config.Source, "revision": p.Config.Revision, "reason": reason, "detail": detail},
	}
	p.LastEvidence, p.LastProv, p.LastStatus = ev, prov, grf.BLOCKED
	return output, prov, ev
}

func (p *ExternalParticipant) validateResponse(input grf.State, c grf.Context, resp externalResponse) error {
	if resp.Source != p.Config.Source {
		return fmt.Errorf("EXTERNAL_SOURCE_MISMATCH:%s", resp.Source)
	}
	if resp.Revision != p.Config.Revision {
		return fmt.Errorf("EXTERNAL_REVISION_MISMATCH:%s", resp.Revision)
	}
	if resp.Evidence.ID == "" || resp.Evidence.Hash == "" {
		return errors.New("EXTERNAL_EVIDENCE_INCOMPLETE")
	}
	parent := safeStateHash(input)
	output := safeStateHash(resp.State)
	if resp.Provenance.ParentHash != parent || resp.Provenance.InputHash != parent {
		return errors.New("EXTERNAL_PROVENANCE_INPUT_MISMATCH")
	}
	if resp.Provenance.OutputHash != output {
		return errors.New("EXTERNAL_PROVENANCE_OUTPUT_MISMATCH")
	}
	if resp.Provenance.SequenceIndex != c.SequenceIndex+1 {
		return errors.New("EXTERNAL_PROVENANCE_SEQUENCE_MISMATCH")
	}
	return nil
}

func (p *ExternalParticipant) EpistemicState() grf.EpistemicState {
	if p.LastStatus != "" { return p.LastStatus }
	return grf.PROJECTED
}

func (p *ExternalParticipant) Invariants() []string {
	return grf.InvariantIDs(grf.CanonicalInvariantSet())
}

func (p *ExternalParticipant) Capabilities() []grf.Capability {
	return []grf.Capability{{
		ID: p.Config.ID,
		Description: "External GRCE complement behind an explicit fail-closed adapter.",
		State: p.EpistemicState(),
		Provenance: grf.Provenance{Stage: "EXTERNAL_PARTICIPANT_REGISTRATION"},
	}}
}

func (p *ExternalParticipant) FailuresAbsorbed() []grf.Failure {
	if p.LastEvidence.FailureID == "" { return nil }
	return []grf.Failure{{
		ID: p.LastEvidence.FailureID,
		Source: p.Config.Source,
		Description: "external runtime availability or contract failure preserved at GRF boundary",
		State: p.LastEvidence.State,
	}}
}

func (p *ExternalParticipant) Provenance() []grf.Provenance {
	if p.LastProv.Stage == "" { return nil }
	return []grf.Provenance{p.LastProv}
}

func splitArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" { return nil }
	return strings.Fields(raw)
}

func safeStateHash(s grf.State) string {
	h, err := s.Hash()
	if err != nil { return "" }
	return h
}
