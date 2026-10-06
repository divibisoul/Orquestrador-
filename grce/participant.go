package grce

import "context"

// Hook identifies one executable boundary of a Golden Rule cycle.
type Hook string

const (
	HookDetect       Hook = "detect"
	HookCharacterize Hook = "characterize"
	HookRegenerate   Hook = "regenerate"
	HookValidate     Hook = "validate"
	HookFreeze       Hook = "freeze"
	HookTrace        Hook = "trace"
)

// EvidenceState is intentionally stricter than implementation status.
type EvidenceState string

const (
	StateReal          EvidenceState = "REAL"
	StateProjected     EvidenceState = "PROJECTED"
	StateBlocked       EvidenceState = "BLOCKED"
	StateUnmeasurable  EvidenceState = "UNMEASURABLE"
)

// Participant is the runtime boundary consumed by the GRCE executor.
// Concrete participants must call a real backend; no mock is implicit.
type Participant interface {
	ExecuteHook(ctx context.Context, hook Hook, input string, correlationID, cycleID string) (HookResult, error)
}

type HookResult struct {
	Hook           Hook
	Operation      string
	InputHash      string
	OutputHash     string
	ParentHash     string
	Payload        map[string]any
	EvidenceState  EvidenceState
}

type StageEvidence struct {
	Index          int
	Hook           Hook
	Operation      string
	CorrelationID  string
	CycleID        string
	ParentHash     string
	OutputHash     string
	State          EvidenceState
}

type CycleResult struct {
	CycleID         string
	CorrelationID   string
	State           EvidenceState
	Stages          []StageEvidence
	FinalOutputHash string
	Error           string
}
