package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/divibisoul/Orquestrador-/grce"
	"github.com/divibisoul/Orquestrador-/grf"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type GRCEFeedback struct {
	Horta func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error
	Vagus func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error
	Mesh  func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error
}

type GRCEExecutorRuntime struct {
	SARA     *SARAProxy
	Compute  *supergpu.Runtime
	Peers    any
	Feedback GRCEFeedback
}

func NewGRCEExecutorRuntime(sara *SARAProxy, compute *supergpu.Runtime, peers any, feedback GRCEFeedback) (*grce.Executor, error) {
	if sara == nil || !sara.Configured() {
		return nil, errors.New("GRCE requires configured SARA authoritative boundary")
	}
	if compute == nil {
		return nil, errors.New("GRCE requires existing SuperGPU runtime")
	}
	if peers == nil {
		return nil, errors.New("GRCE requires existing canonical Mesh peer client")
	}
	r := &GRCEExecutorRuntime{SARA: sara, Compute: compute, Peers: peers, Feedback: feedback}
	hooks := grce.Hooks{
		Snapshot:            r.snapshot,
		DetectGPU:           r.detect,
		DetectCPU:           r.detectCPU,
		Characterize:       r.characterize,
		Dualize:            r.dualize,
		EthicalGate:        r.ethicalGate,
		AnalyzeGPU:         r.analyzeGPU,
		AnalyzeCPU:         r.analyzeCPU,
		Mediate:             r.mediate,
		Regenerate:          r.regenerate,
		EthicalTransform:    r.ethicalTransform,
		Innovate:            r.innovate,
		Form:                r.form,
		ValidateETR:         r.validateETR,
		ValidateITR:         r.validateITR,
		ValidateRGO:         r.validateRGO,
		Rollback:            r.rollback,
		Horta:               r.feedback("horta"),
		Vagus:               r.feedback("vagus"),
		Mesh:                r.feedback("mesh"),
		ExtractCapabilities: r.extractCapabilities,
	}
	return grce.New(grf.CanonicalInvariantSet(), hooks)
}

func stateText(state grf.State) string {
	if state.Payload != nil {
		if value, ok := state.Payload["text"].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	raw, err := json.Marshal(state.Payload)
	if err != nil {
		return fmt.Sprintf("state:%s", state.ID)
	}
	return string(raw)
}

func stateHash(state grf.State) (string, error) {
	return state.Hash()
}

func makeArtifact(stage string, state grf.State, c grf.Context, payload map[string]any, failureID string, sequence uint64) (grf.Artifact, error) {
	parent, err := stateHash(state)
	if err != nil {
		return grf.Artifact{}, err
	}
	output, err := grf.HashJSON(payload)
	if err != nil {
		return grf.Artifact{}, err
	}
	return grf.Artifact{
		ID:      stage + ":" + c.TraceID,
		Stage:   stage,
		State:   grf.REAL,
		Payload: payload,
		Provenance: grf.Provenance{
			ParentHash:      parent,
			InputHash:       parent,
			OutputHash:      output,
			SequenceIndex:   sequence,
			Stage:           stage,
			CausalFailureID: failureID,
		},
	}, nil
}

func (r *GRCEExecutorRuntime) snapshot(ctx context.Context, state grf.State, c grf.Context) (grf.Artifact, error) {
	if err := c.Validate(); err != nil {
		return grf.Artifact{}, err
	}
	if _, err := r.SARA.State(ctx, c.CorrelationID); err != nil {
		return grf.Artifact{}, fmt.Errorf("GRCE_ERU_SNAPSHOT_UNAVAILABLE:%w", err)
	}
	return makeArtifact("PREFLIGHT", state, c, map[string]any{
		"state_id":         state.ID,
		"state_hash":       mustStateHash(state),
		"epistemic_state": string(state.EpistemicState),
		"authority":        "SARA/ERU",
	}, "", c.SequenceIndex+1)
}

func mustStateHash(state grf.State) string {
	value, err := stateHash(state)
	if err != nil {
		return ""
	}
	return value
}

func (r *GRCEExecutorRuntime) detect(ctx context.Context, state grf.State, c grf.Context) ([]grf.Failure, error) {
	audit, err := r.SARA.Audit(ctx, stateText(state), c.CorrelationID)
	if err != nil {
		return nil, fmt.Errorf("GRCE_ARA_AUDIT_FAILED:%w", err)
	}
	items, ok := audit["flaws"].([]any)
	if !ok {
		return nil, nil
	}
	failures := make([]grf.Failure, 0, len(items))
	for index, item := range items {
		detail := fmt.Sprint(item)
		kind := "SARA/ARA"
		if record, ok := item.(map[string]any); ok {
			if value, exists := record["detail"]; exists {
				detail = fmt.Sprint(value)
			}
			if value, exists := record["kind"]; exists {
				detail = fmt.Sprint(value) + ": " + detail
			}
		}
		hash, err := grf.HashJSON(map[string]any{"index": index, "detail": detail, "correlation": c.CorrelationID})
		if err != nil {
			return nil, err
		}
		failures = append(failures, grf.Failure{
			ID:                  "ara:" + hash[:16],
			Source:              kind,
			Description:         detail,
			RequiredOpposition:  "preserve causal history while increasing state evidence",
			PropertyNecessary:   "provenance-preserving transformation",
			PropertyDeclared:    true,
			State:               grf.REAL,
		})
	}
	return failures, nil
}

func (r *GRCEExecutorRuntime) detectCPU(ctx context.Context, state grf.State, c grf.Context) ([]grf.Failure, error) {
	_ = ctx
	_ = c
	if strings.TrimSpace(state.ID) == "" {
		return []grf.Failure{{
			ID: "grf:state-id-missing", Source: "N07.GRCE",
			Description: "state identifier missing",
			RequiredOpposition: "explicit state identity",
			PropertyNecessary:  "traceable identity",
			PropertyDeclared:   true,
			State:              grf.REAL,
		}}, nil
	}
	return nil, nil
}

func (r *GRCEExecutorRuntime) characterize(ctx context.Context, evidence []grf.Evidence, c grf.Context) ([]grf.Characterization, error) {
	_ = ctx
	out := make([]grf.Characterization, 0, len(evidence))
	for _, item := range evidence {
		out = append(out, grf.Characterization{
			FailureID:   item.FailureID,
			Description: fmt.Sprint(item.Payload["failure"]),
			Properties: map[string]any{
				"evidence_hash":  item.Hash,
				"input_hash":     item.InputHash,
				"sequence_index": c.SequenceIndex,
			},
			State: grf.REAL,
		})
	}
	return out, nil
}

func (r *GRCEExecutorRuntime) dualize(ctx context.Context, chars []grf.Characterization, c grf.Context) ([]grf.Opposition, error) {
	_ = ctx
	out := make([]grf.Opposition, 0, len(chars))
	for _, item := range chars {
		out = append(out, grf.Opposition{
			FailureID:         item.FailureID,
			Description:       item.Description,
			NecessaryProperty: "preserve causal history and provenance while transforming",
			PropertyDeclared:  true,
			State:             grf.REAL,
		})
	}
	return out, nil
}

func (r *GRCEExecutorRuntime) ethicalGate(ctx context.Context, oppositions []grf.Opposition, c grf.Context) error {
	lines := make([]string, 0, len(oppositions))
	for _, item := range oppositions {
		lines = append(lines, item.NecessaryProperty+": "+item.Description)
	}
	sort.Strings(lines)
	audit, err := r.SARA.Audit(ctx, strings.Join(lines, "\n"), c.CorrelationID)
	if err != nil {
		return fmt.Errorf("GRCE_ETR_GATE_FAILED:%w", err)
	}
	ethical, ok := audit["ethical"].(map[string]any)
	if !ok {
		return errors.New("GRCE_ETR_RESULT_MISSING")
	}
	approved, ok := ethical["approved"].(bool)
	if !ok || !approved {
		return errors.New("GRCE_ETR_REJECTED")
	}
	return nil
}

func (r *GRCEExecutorRuntime) analyzeGPU(ctx context.Context, oppositions []grf.Opposition, c grf.Context) ([]grf.Artifact, error) {
	device, err := r.Compute.Select("")
	if err != nil {
		return nil, fmt.Errorf("GRCE_SUPERGPU_SELECT_FAILED:%w", err)
	}
	out := make([]grf.Artifact, 0, len(oppositions))
	for index, item := range oppositions {
		input := make([]float64, len(item.Description))
		for i, value := range []byte(item.Description) {
			input[i] = float64(value)
		}
		values, err := r.Compute.Execute(ctx, device, "normalize", input)
		if err != nil {
			return nil, fmt.Errorf("GRCE_SUPERGPU_ANALYZE_FAILED:%w", err)
		}
		artifact, err := makeArtifact("ANALYZE_GPU", grf.State{ID: item.FailureID, Payload: map[string]any{"description": item.Description}}, c, map[string]any{
			"failure_id":             item.FailureID,
			"device_id":              device.ID,
			"normalized_vector_size": len(values),
		}, item.FailureID, c.SequenceIndex+500+uint64(index))
		if err != nil {
			return nil, err
		}
		out = append(out, artifact)
	}
	return out, nil
}

func (r *GRCEExecutorRuntime) analyzeCPU(ctx context.Context, oppositions []grf.Opposition, c grf.Context) ([]grf.Artifact, error) {
	_ = ctx
	out := make([]grf.Artifact, 0, len(oppositions))
	for index, item := range oppositions {
		encoded, err := grf.HashJSON(item)
		if err != nil {
			return nil, err
		}
		artifact, err := makeArtifact("ANALYZE_CPU", grf.State{ID: item.FailureID, Payload: map[string]any{"description": item.Description}}, c, map[string]any{
			"failure_id":     item.FailureID,
			"description_hash": encoded,
		}, item.FailureID, c.SequenceIndex+1000+uint64(index))
		if err != nil {
			return nil, err
		}
		out = append(out, artifact)
	}
	return out, nil
}

func (r *GRCEExecutorRuntime) mediate(ctx context.Context, artifacts []grf.Artifact, state grf.State, c grf.Context) ([]grf.Artifact, error) {
	_ = ctx
	payloads := make([]any, 0, len(artifacts))
	for _, item := range artifacts {
		payloads = append(payloads, item.Payload)
	}
	return makeArtifactsSingle("MMD", state, c, map[string]any{
		"inputs": payloads,
		"count":  len(payloads),
	}, "")
}

func makeArtifactsSingle(stage string, state grf.State, c grf.Context, payload map[string]any, failureID string) ([]grf.Artifact, error) {
	sequence := c.SequenceIndex + 2000
	switch stage {
	case "ARA":
		sequence = c.SequenceIndex + 2001
	case "ETR":
		sequence = c.SequenceIndex + 2002
	case "ITR":
		sequence = c.SequenceIndex + 2003
	}
	artifact, err := makeArtifact(stage, state, c, payload, failureID, sequence)
	if err != nil {
		return nil, err
	}
	return []grf.Artifact{artifact}, nil
}

func (r *GRCEExecutorRuntime) regenerate(ctx context.Context, artifact grf.Artifact, state grf.State, c grf.Context) (grf.Artifact, error) {
	input := stateText(state)
	regenerated, err := r.SARA.Regenerate(ctx, input, c.CorrelationID)
	if err != nil {
		return grf.Artifact{}, fmt.Errorf("GRCE_ARA_REGENERATE_FAILED:%w", err)
	}
	transformed, ok := regenerated["transformed"].(string)
	if ok && len(transformed) < len(input) {
		return grf.Artifact{}, errors.New("GRCE_ARA_MONOTONICITY_VIOLATION")
	}
	artifacts, err := makeArtifactsSingle("ARA", state, c, map[string]any{
		"input":           input,
		"regeneration":    regenerated,
		"source_artifact": artifact.ID,
	}, artifact.Provenance.CausalFailureID)
	if err != nil {
		return grf.Artifact{}, err
	}
	if len(artifacts) != 1 {
		return grf.Artifact{}, errors.New("GRCE_ARTIFACT_CARDINALITY_VIOLATION")
	}
	return artifacts[0], nil
}

func (r *GRCEExecutorRuntime) ethicalTransform(ctx context.Context, artifact grf.Artifact, state grf.State, c grf.Context) (grf.Artifact, error) {
	raw, err := json.Marshal(artifact.Payload)
	if err != nil {
		return grf.Artifact{}, err
	}
	audit, err := r.SARA.Audit(ctx, string(raw), c.CorrelationID)
	if err != nil {
		return grf.Artifact{}, fmt.Errorf("GRCE_ETR_TRANSFORM_AUDIT_FAILED:%w", err)
	}
	ethical, ok := audit["ethical"].(map[string]any)
	if !ok {
		return grf.Artifact{}, errors.New("GRCE_ETR_TRANSFORM_RESULT_MISSING")
	}
	approved, ok := ethical["approved"].(bool)
	if !ok || !approved {
		return grf.Artifact{}, errors.New("GRCE_ETR_TRANSFORM_REJECTED")
	}
	artifacts, err := makeArtifactsSingle("ETR", state, c, map[string]any{
		"source_artifact": artifact.ID,
		"ethical":        ethical,
	}, artifact.Provenance.CausalFailureID)
	if err != nil {
		return grf.Artifact{}, err
	}
	return artifacts[0], nil
}

func (r *GRCEExecutorRuntime) innovate(ctx context.Context, artifact grf.Artifact, state grf.State, c grf.Context) (grf.Artifact, error) {
	payload := map[string]any{
		"source_artifact":  artifact.ID,
		"preserved_failure": artifact.Provenance.CausalFailureID,
		"proposal":         "extend evidence without deleting historical state",
	}
	artifacts, err := makeArtifactsSingle("ITR", state, c, payload, artifact.Provenance.CausalFailureID)
	if err != nil {
		return grf.Artifact{}, err
	}
	return artifacts[0], nil
}

func (r *GRCEExecutorRuntime) form(ctx context.Context, artifacts []grf.Artifact, state grf.State, c grf.Context) (grf.State, error) {
	_ = ctx
	payload := map[string]any{
		"original_state": state.Payload,
		"artifacts":      artifacts,
		"trace_id":       c.TraceID,
		"correlation_id": c.CorrelationID,
	}
	return grf.State{ID: state.ID + ":grce", EpistemicState: grf.PROJECTED, Payload: payload}, nil
}

func (r *GRCEExecutorRuntime) validateETR(ctx context.Context, state grf.State, c grf.Context) (bool, error) {
	audit, err := r.SARA.Audit(ctx, stateText(state), c.CorrelationID)
	if err != nil {
		return false, err
	}
	ethical, ok := audit["ethical"].(map[string]any)
	if !ok {
		return false, errors.New("GRCE_VALIDATE_ETR_RESULT_MISSING")
	}
	approved, ok := ethical["approved"].(bool)
	return ok && approved, nil
}

func (r *GRCEExecutorRuntime) validateITR(ctx context.Context, state grf.State, c grf.Context) (bool, error) {
	_ = ctx
	_ = c
	originalRaw, err := json.Marshal(state.Payload["original_state"])
	if err != nil {
		return false, err
	}
	transformedRaw, err := json.Marshal(state.Payload["artifacts"])
	if err != nil {
		return false, err
	}
	return len(transformedRaw) >= len(originalRaw) && len(transformedRaw) > 0, nil
}

func (r *GRCEExecutorRuntime) validateRGO(ctx context.Context, state grf.State, c grf.Context) (bool, error) {
	finding := map[string]any{
		"finding_id":  "grce:" + c.CorrelationID,
		"description": stateText(state),
		"provenance": map[string]any{
			"trace_id":       c.TraceID,
			"sequence_index": c.SequenceIndex,
		},
	}
	result, err := r.SARA.RGOTrinity(ctx, finding, c.CorrelationID, c.CorrelationID)
	if err != nil {
		return false, fmt.Errorf("GRCE_RGO_TRINITY_FAILED:%w", err)
	}
	status, _ := result["final_status"].(string)
	return strings.EqualFold(status, "VALIDATED"), nil
}

func (r *GRCEExecutorRuntime) rollback(ctx context.Context, state grf.State, prov []grf.Provenance, evidence []grf.Evidence, c grf.Context) error {
	_, err := r.SARA.State(ctx, c.CorrelationID)
	return err
}

func (r *GRCEExecutorRuntime) feedback(kind string) grce.FeedbackFunc {
	return func(ctx context.Context, state grf.State, prov []grf.Provenance, evidence []grf.Evidence, caps []grf.Capability, c grf.Context) error {
		var fn func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error
		switch kind {
		case "horta":
			fn = r.Feedback.Horta
		case "vagus":
			fn = r.Feedback.Vagus
		case "mesh":
			fn = r.Feedback.Mesh
		}
		if fn == nil {
			return errors.New("GRCE_FEEDBACK_NOT_BOUND:" + kind)
		}
		return fn(ctx, state, prov, evidence, caps, c)
	}
}

func NewGRCEVagoFeedback(gateway *NervoVagoGateway, target, kind string) func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error {
	return func(ctx context.Context, state grf.State, prov []grf.Provenance, evidence []grf.Evidence, caps []grf.Capability, c grf.Context) error {
		if gateway == nil {
			return errors.New("GRCE_NERVO_VAGO_GATEWAY_UNAVAILABLE")
		}
		payload := map[string]any{
			"kind":       kind,
			"state_id":   state.ID,
			"provenance": prov,
			"evidence":   evidence,
			"capabilities": caps,
			"feedback":   "GRCE",
		}
		inputHash, err := grf.HashJSON(payload)
		if err != nil {
			return err
		}
		parentHash := inputHash
		if len(prov) > 0 && prov[len(prov)-1].OutputHash != "" {
			parentHash = prov[len(prov)-1].OutputHash
		}
		messageID := protocol.NewTraceID()
		traceID := protocol.NewTraceID()
		_, err = gateway.Publish(ctx, NervoVagoEnvelope{
			VagusVersion:  "1.0",
			MessageID:     messageID,
			CorrelationID: c.CorrelationID,
			Source:        "N07.GRCE",
			Target:        target,
			Priority:      100,
			TTL:           5000,
			Type:          "nervo.grce.feedback." + kind,
			Payload:       payload,
			Provenance: NervoVagoProvenance{
				TraceID:       traceID,
				CorrelationID: c.CorrelationID,
				MessageID:     messageID,
				SequenceIndex: c.SequenceIndex + 4000 + uint64(len(prov)),
				ParentHash:    parentHash,
				InputHash:     inputHash,
			},
		})
		return err
	}
}

func (r *GRCEExecutorRuntime) extractCapabilities(ctx context.Context, state grf.State, prov []grf.Provenance, evidence []grf.Evidence, c grf.Context) ([]grf.Capability, error) {
	_ = ctx
	if len(prov) == 0 || len(evidence) == 0 {
		return nil, errors.New("GRCE_CAPABILITY_EXTRACTION_INPUT_MISSING")
	}
	failureID := evidence[0].FailureID
	hash, err := grf.HashJSON(map[string]any{
		"failure":    failureID,
		"correlation": c.CorrelationID,
		"state":      state.Payload,
	})
	if err != nil {
		return nil, err
	}
	return []grf.Capability{{
		ID:          "capability:" + hash[:16],
		Description: "preserve-and-transform capability derived from observed failure",
		State:       grf.PROJECTED,
		Provenance: grf.Provenance{
			ParentHash:       mustStateHash(state),
			InputHash:        mustStateHash(state),
			OutputHash:       hash,
			SequenceIndex:    c.SequenceIndex + 3000,
			Stage:            "EXTRACT",
			CausalFailureID:  failureID,
		},
	}}, nil
}
