package rgo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SchemaVersion = "1.0.0"

type VerificationState string
const (
	Verified VerificationState = "VERIFIED"
	Unverified VerificationState = "UNVERIFIED"
	Inconclusive VerificationState = "INCONCLUSIVE"
)

type EpistemicMode string
const (
	Execution EpistemicMode = "EXECUTION"
	Inspection EpistemicMode = "INSPECTION"
	Inference EpistemicMode = "INFERENCE"
	External EpistemicMode = "EXTERNAL"
)

type ActionabilityStatus string
const (
	Actionable ActionabilityStatus = "ACTIONABLE"
	NonActionable ActionabilityStatus = "NON_ACTIONABLE"
	Informational ActionabilityStatus = "INFORMATIONAL"
	Duplicate ActionabilityStatus = "DUPLICATE"
	InconclusiveActionability ActionabilityStatus = "INCONCLUSIVE"
	UndefinedActionability ActionabilityStatus = "UNDEFINED"
)

type DualStatus string
const (
	DualDerived DualStatus = "DERIVED_FROM_CONTRACT"
	DualUnresolved DualStatus = "UNRESOLVED"
	DualNotApplicable DualStatus = "NOT_APPLICABLE"
)

type CapabilityState string
const (
	CapabilityDerived CapabilityState = "DERIVED"
	CapabilityImplemented CapabilityState = "IMPLEMENTED"
	CapabilityTested CapabilityState = "TESTED"
	CapabilityValidated CapabilityState = "VALIDATED"
	CapabilityPromotable CapabilityState = "PROMOTABLE"
	CapabilityPromoted CapabilityState = "PROMOTED"
)

type EvidenceRef struct {
	ID string `json:"id"`
	Kind string `json:"kind"`
	Ref string `json:"ref"`
}

type CorrectionBoundary struct {
	ProblemToResolve string `json:"problem_to_resolve"`
	RequiredProperty string `json:"required_property"`
}

type Failure struct {
	Type string `json:"type"`
	Title string `json:"title"`
	Description string `json:"description"`
	Nature string `json:"nature"`
	Cause string `json:"cause,omitempty"`
	Impact string `json:"impact,omitempty"`
}

type Envelope struct {
	SchemaVersion string `json:"schema_version"`
	FindingID string `json:"finding_id"`
	ObjectID string `json:"object_id"`
	Timestamp string `json:"timestamp"`
	CorrelationID string `json:"correlation_id"`
	TraceID string `json:"trace_id"`
	Source struct {
		System string `json:"system"`
		Module string `json:"module"`
		Version string `json:"version"`
	} `json:"source"`
	Epistemic struct {
		Mode EpistemicMode `json:"mode"`
		Verification VerificationState `json:"verification_state"`
	} `json:"epistemic"`
	Actionability struct {
		Status ActionabilityStatus `json:"status"`
		Reason string `json:"reason,omitempty"`
	} `json:"actionability"`
	Failure Failure `json:"failure"`
	CorrectionBoundary CorrectionBoundary `json:"correction_boundary"`
	Dual struct {
		Status DualStatus `json:"status"`
		Property string `json:"property,omitempty"`
		EvidenceRefs []string `json:"evidence_refs,omitempty"`
	} `json:"dual"`
	Capability struct {
		ID string `json:"id,omitempty"`
		State CapabilityState `json:"state,omitempty"`
	} `json:"capability"`
	Evidence []EvidenceRef `json:"evidence"`
	Provenance struct {
		Origin string `json:"origin"`
		ParentIDs []string `json:"parent_ids,omitempty"`
		InputHash string `json:"input_hash"`
	} `json:"provenance"`
}

func (e Envelope) Validate() error {
	if e.SchemaVersion != SchemaVersion { return fmt.Errorf("RGO_SCHEMA_VERSION_UNSUPPORTED:%s", e.SchemaVersion) }
	if strings.TrimSpace(e.FindingID) == "" || strings.TrimSpace(e.ObjectID) == "" { return errors.New("RGO_ID_REQUIRED") }
	if _, err := time.Parse(time.RFC3339Nano, e.Timestamp); err != nil { return fmt.Errorf("RGO_TIMESTAMP_INVALID:%w", err) }
	if strings.TrimSpace(e.CorrelationID) == "" || strings.TrimSpace(e.TraceID) == "" { return errors.New("RGO_CORRELATION_TRACE_REQUIRED") }
	if strings.TrimSpace(e.Source.System) == "" || strings.TrimSpace(e.Source.Module) == "" || strings.TrimSpace(e.Source.Version) == "" { return errors.New("RGO_SOURCE_REQUIRED") }
	if e.Epistemic.Mode == "" || e.Epistemic.Verification == "" { return errors.New("RGO_EPISTEMIC_REQUIRED") }
	if e.Actionability.Status == "" { return errors.New("RGO_ACTIONABILITY_REQUIRED") }
	if strings.TrimSpace(e.Failure.Type) == "" || strings.TrimSpace(e.Failure.Description) == "" || strings.TrimSpace(e.Failure.Nature) == "" { return errors.New("RGO_FAILURE_REQUIRED") }
	
	if e.Dual.Status == DualDerived && (strings.TrimSpace(e.CorrectionBoundary.ProblemToResolve) == "" || strings.TrimSpace(e.CorrectionBoundary.RequiredProperty) == "" || strings.TrimSpace(e.Dual.Property) == "") { return errors.New("RGO_DUAL_DERIVATION_REQUIRED") }
	if len(e.Evidence) == 0 { return errors.New("RGO_EVIDENCE_REQUIRED") }
	if strings.TrimSpace(e.Provenance.Origin) == "" || strings.TrimSpace(e.Provenance.InputHash) == "" { return errors.New("RGO_PROVENANCE_REQUIRED") }
	for _, ev := range e.Evidence {
		if strings.TrimSpace(ev.ID) == "" || strings.TrimSpace(ev.Kind) == "" || strings.TrimSpace(ev.Ref) == "" { return errors.New("RGO_EVIDENCE_REF_INVALID") }
	}
	return nil
}

func DeriveDual(e Envelope) Envelope {
	if strings.TrimSpace(e.CorrectionBoundary.ProblemToResolve) == "" || strings.TrimSpace(e.CorrectionBoundary.RequiredProperty) == "" {
		e.Dual.Status = DualUnresolved
		e.Dual.Property = ""
		e.Dual.EvidenceRefs = nil
		return e
	}
	e.Dual.Status = DualDerived
	e.Dual.Property = e.CorrectionBoundary.RequiredProperty
	ids := make([]string, 0, len(e.Evidence))
	for _, ev := range e.Evidence { ids = append(ids, ev.ID) }
	e.Dual.EvidenceRefs = ids
	return e
}

func CanonicalHash(e Envelope) (string, error) {
	raw, err := json.Marshal(e)
	if err != nil { return "", err }
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
