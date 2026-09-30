package rgo

import (
	"fmt"
	"strings"
	"time"
)

func FromBugShield(payload map[string]any) (Envelope, error) {
	schema, _ := payload["schema_version"].(string)
	if !strings.HasPrefix(schema, "1.1.") && !strings.HasPrefix(schema, "1.2.") {
		return Envelope{}, fmt.Errorf("BUGSHIELD_SCHEMA_UNSUPPORTED:%s", schema)
	}
	scanID, _ := payload["scan_id"].(string)
	timestamp, _ := payload["timestamp"].(string)
	scanner, _ := payload["scanner"].(map[string]any)
	scope, _ := payload["scope"].(map[string]any)
	finding, _ := payload["finding"].(map[string]any)
	if scanID == "" || timestamp == "" || scanner == nil || scope == nil || finding == nil {
		return Envelope{}, fmt.Errorf("BUGSHIELD_REQUIRED_SECTION_MISSING")
	}
	inputHash, _ := scope["input_hash"].(string)
	if inputHash == "" {
		return Envelope{}, fmt.Errorf("BUGSHIELD_INPUT_HASH_REQUIRED")
	}
	findingID, _ := finding["id"].(string)
	description, _ := finding["description"].(string)
	findingType, _ := finding["type"].(string)
	category, _ := finding["category"].(string)
	epMode, _ := finding["epistemic_mode"].(string)
	verify, _ := finding["verification_state"].(string)
	if findingID == "" || description == "" || findingType == "" || category == "" || epMode == "" || verify == "" {
		return Envelope{}, fmt.Errorf("BUGSHIELD_FINDING_REQUIRED")
	}
	evidenceRaw, _ := finding["evidence"].([]any)
	evidence := make([]EvidenceRef, 0, len(evidenceRaw))
	for _, item := range evidenceRaw {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, _ := row["id"].(string)
		kind, _ := row["kind"].(string)
		ref, _ := row["ref"].(string)
		if id != "" && kind != "" && ref != "" {
			evidence = append(evidence, EvidenceRef{ID: id, Kind: kind, Ref: ref})
		}
	}
	if len(evidence) == 0 {
		return Envelope{}, fmt.Errorf("BUGSHIELD_EVIDENCE_REQUIRED")
	}
	var env Envelope
	env.SchemaVersion = SchemaVersion
	env.FindingID = findingID
	env.ObjectID = inputHash
	env.Timestamp = timestamp
	if _, err := time.Parse(time.RFC3339Nano, timestamp); err != nil {
		return Envelope{}, err
	}
	env.CorrelationID = scanID
	env.TraceID = scanID + ":" + findingID
	env.Source.System, _ = scanner["name"].(string)
	env.Source.Version, _ = scanner["version"].(string)
	env.Source.Module = "BugShield"
	env.Epistemic.Mode = EpistemicMode(epMode)
	env.Epistemic.Verification = VerificationState(verify)
	env.Actionability.Status = UndefinedActionability
	env.Failure.Type = findingType
	env.Failure.Title, _ = finding["type"].(string)
	env.Failure.Description = description
	env.Failure.Nature = category
	env.Failure.Cause, _ = finding["why_it_matters"].(string)
	boundary, _ := finding["correction_boundary"].(map[string]any)
	if boundary != nil {
		env.CorrectionBoundary.ProblemToResolve, _ = boundary["problem_to_resolve"].(string)
		env.CorrectionBoundary.RequiredProperty, _ = boundary["required_property"].(string)
	}
	env.Dual.Status = DualUnresolved
	env.Evidence = evidence
	env.Provenance.Origin = "BugShield:" + scanID
	env.Provenance.InputHash = inputHash
	env.Extensions = map[string]any{"bugshield": payload}
	return DeriveDual(env), nil
}
