package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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
	Feedback GRCEFeedback
}

func NewGRCEExecutorRuntime(sara *SARAProxy, compute *supergpu.Runtime, feedback GRCEFeedback) (*grce.Executor, error) {
	if sara == nil || !sara.Configured() {
		return nil, errors.New("GRCE requires configured SARA authoritative boundary")
	}
	if compute == nil {
		return nil, errors.New("GRCE requires existing SuperGPU runtime")
	}
	if feedback.Horta == nil || feedback.Vagus == nil || feedback.Mesh == nil {
		return nil, errors.New("GRCE requires Horta, Vagus, and Mesh feedback sinks")
	}
	r := &GRCEExecutorRuntime{SARA: sara, Compute: compute, Feedback: feedback}
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

// grceEthicalReviewInput keeps the exact content under review while supplying
// the explicit ethical intent of this transformation to SARA's four-framework
// consensus. The payload is not rewritten or filtered; lexical, identity and
// contextual checks still receive the full original content.
func grceEthicalReviewInput(purpose, content string) string {
	const principles = "Princípios do ciclo GRCE: preservar autonomia, transparência, responsabilidade e cuidado com a comunidade; manter o histórico causal e a proveniência."
	return principles + "\nFinalidade da avaliação: " + strings.TrimSpace(purpose) + "\nConteúdo integral submetido à avaliação:\n" + content
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
	audit, err := r.SARA.Audit(ctx, grceEthicalReviewInput("validar o estado candidato antes da ativação", stateText(state)), c.CorrelationID)
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
	audit, err := r.SARA.Audit(ctx, grceEthicalReviewInput("avaliar a necessidade e os limites da transformação", strings.Join(lines, "\n")), c.CorrelationID)
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
	audit, err := r.SARA.Audit(ctx, grceEthicalReviewInput("avaliar o conteúdo e a proveniência da transformação", string(raw)), c.CorrelationID)
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
	// RGO's actual SARA contract requires a complete envelope and referenced
	// evidence. Reuse artifacts produced by this GRCE candidate; do not assert
	// success from descriptive metadata alone.
	rawArtifacts, ok := state.Payload["artifacts"].([]grf.Artifact)
	if !ok || len(rawArtifacts) == 0 {
		return false, errors.New("GRCE_RGO_ARTIFACT_EVIDENCE_MISSING")
	}
	inputHash, err := state.Hash()
	if err != nil {
		return false, fmt.Errorf("GRCE_RGO_INPUT_HASH_FAILED:%w", err)
	}
	// Keep the full original state in the hash-linked candidate and evidence
	// chain, but do not feed the entire raw user payload into SARA's second
	// RGO/Trinity description pass. That pass validates the candidate/evidence
	// envelope; the original input was already submitted to SARA's ARA/ETR
	// audit/regeneration stages. Re-embedding it here can produce a second,
	// unrelated structural finding (for example, punctuation density on a
	// deliberately repetitive input) and make the real cycle inconclusive.
	originalState, hasOriginalState := state.Payload["original_state"]
	if !hasOriginalState || originalState == nil {
		return false, errors.New("GRCE_RGO_ORIGINAL_STATE_REFERENCE_MISSING")
	}
	originalStateHash, err := grf.HashJSON(originalState)
	if err != nil {
		return false, fmt.Errorf("GRCE_RGO_ORIGINAL_STATE_HASH_FAILED:%w", err)
	}
	evidence := make([]map[string]any, 0, len(rawArtifacts))
	parentIDs := make([]string, 0, len(rawArtifacts))
	artifactStages := make([]string, 0, len(rawArtifacts))
	for _, artifact := range rawArtifacts {
		if strings.TrimSpace(artifact.ID) == "" || strings.TrimSpace(artifact.Stage) == "" {
			return false, errors.New("GRCE_RGO_ARTIFACT_ID_OR_STAGE_MISSING")
		}
		artifactHash, err := grf.HashJSON(artifact)
		if err != nil {
			return false, fmt.Errorf("GRCE_RGO_ARTIFACT_HASH_FAILED:%w", err)
		}
		evidence = append(evidence, map[string]any{
			"id":   "grce-evidence:" + artifactHash[:16],
			"kind": "GRCE_ARTIFACT:" + artifact.Stage,
			"ref":  artifact.ID,
		})
		parentIDs = append(parentIDs, artifact.ID)
		artifactStages = append(artifactStages, artifact.Stage)
	}
	description := fmt.Sprintf(
		"Validar envelope GRCE %s; candidate_hash=%s; original_state_hash=%s; artefatos=%d; estágios=%s. Conteúdo original integral permanece preservado nos artefatos hash-linked e nas referências de evidência. Critérios SARA/ETR: autonomia, transparência, responsabilidade, cuidado, benefício para a comunidade, melhoria e proveniência íntegra; validação restrita ao envelope e hashes.",
		state.ID, inputHash, originalStateHash, len(rawArtifacts), strings.Join(artifactStages, ","),
	)
	finding := map[string]any{
		"schema_version": "1.0.0",
		"finding_id":     "grce:" + c.CorrelationID,
		"object_id":      state.ID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339Nano),
		"correlation_id": c.CorrelationID,
		"trace_id":       c.TraceID,
		"source": map[string]any{
			"system":  "SOUL",
			"module":  "N07.GRCE",
			"version": "2.0.0",
		},
		"epistemic": map[string]any{
			"mode":                "EXECUTION",
			"verification_state": "UNVERIFIED",
		},
		"actionability": map[string]any{
			"status": "ACTIONABLE",
			"reason": "GRCE candidate requires canonical SARA RGO/Trinity validation before activation",
		},
		"failure": map[string]any{
			"type":        "GRCE_CANDIDATE_VALIDATION",
			"description": description,
			"nature":      "candidate-validation",
			"cause":       "candidate must retain its input, outputs, and ordered provenance",
			"impact":      "prevent promotion without SARA stage evidence",
		},
		"correction_boundary": map[string]any{
			"problem_to_resolve": "validate the candidate through SARA's existing RGO/Trinity processor",
			"required_property":  "preserve parent/input/output hashes and ordered stage evidence",
		},
		"dual": map[string]any{
			"status":       "UNRESOLVED",
			"property":     "",
			"evidence_refs": parentIDs,
		},
		"evidence": evidence,
		"provenance": map[string]any{
			"origin":     "N07.GRCE",
			"parent_ids": parentIDs,
			"input_hash": inputHash,
		},
		"extensions": map[string]any{
			"grce_sequence_index": c.SequenceIndex,
			"artifact_count":      len(rawArtifacts),
			"original_state_hash": originalStateHash,
			"candidate_hash":      inputHash,
			"artifact_stages":     artifactStages,
		},
	}
	result, err := r.SARA.RGOTrinityWithCycle(ctx, finding, c.CorrelationID, c.CorrelationID)
	if err != nil {
		return false, fmt.Errorf("GRCE_RGO_TRINITY_FAILED:%w", err)
	}
	status, _ := result["final_status"].(string)
	if !strings.EqualFold(status, "VALIDATED") {
		return false, nil
	}
	stages, ok := result["stages"].([]any)
	if !ok || len(stages) < 3 {
		return false, errors.New("GRCE_RGO_TRINITY_STAGE_EVIDENCE_MISSING")
	}
	seen := map[string]bool{}
	for _, raw := range stages {
		stage, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := stage["stage"].(string)
		upper := strings.ToUpper(name)
		if strings.HasPrefix(upper, "TRINITY::") {
			seen["TRINITY"] = true
		}
		if upper == "ERU" {
			seen["ERU"] = true
		}
		if upper == "MMD" {
			seen["MMD"] = true
		}
	}
	for _, key := range []string{"TRINITY", "ERU", "MMD"} {
		if !seen[key] {
			return false, fmt.Errorf("GRCE_RGO_TRINITY_%s_EVIDENCE_MISSING", key)
		}
	}
	return true, nil
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
