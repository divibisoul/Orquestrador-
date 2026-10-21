package grce

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/divibisoul/Orquestrador-/backend"
)

type saraClient interface {
	Audit(context.Context, string, string) (map[string]any, error)
	Capabilities(context.Context, string) (map[string]any, error)
	Regenerate(context.Context, string, string) (map[string]any, error)
	State(context.Context, string) (map[string]any, error)
	Trace(context.Context, string, string) (map[string]any, error)
	Cycle(context.Context, string, string, map[string]any) (map[string]any, error)
}

// SARAParticipant is the real Go->HTTP boundary to the Python SARA runtime.
type SARAParticipant struct {
	client saraClient
}

func NewSARAParticipant(client *backend.SARAProxy) *SARAParticipant {
	if client == nil {
		return &SARAParticipant{}
	}
	return &SARAParticipant{client: client}
}

func (p *SARAParticipant) ExecuteHook(ctx context.Context, hook Hook, input, correlationID, cycleID string) (HookResult, error) {
	if p == nil || p.client == nil {
		return HookResult{Hook: hook, EvidenceState: StateBlocked}, errors.New("GRCE_SARA_PARTICIPANT_UNCONFIGURED")
	}

	var operation string
	var payload map[string]any
	var err error

	switch hook {
	case HookDetect:
		operation = "sara.audit@1.0.0"
		payload, err = p.client.Audit(ctx, input, correlationID)
	case HookCharacterize:
		operation = "sara.capabilities@1.0.0"
		payload, err = p.client.Capabilities(ctx, correlationID)
	case HookRegenerate:
		operation = "sara.regenerate@1.0.0"
		payload, err = p.client.Regenerate(ctx, input, correlationID)
	case HookValidate:
		operation = "sara.audit@1.0.0"
		payload, err = p.client.Audit(ctx, input, correlationID)
	case HookFreeze:
		operation = "sara.state@1.0.0"
		payload, err = p.client.State(ctx, correlationID)
	case HookTrace:
		operation = "sara.trace@1.0.0"
		payload, err = p.client.Trace(ctx, cycleID, correlationID)
	default:
		return HookResult{Hook: hook, EvidenceState: StateUnmeasurable}, errors.New("GRCE_UNKNOWN_HOOK")
	}

	if err != nil {
		return HookResult{Hook: hook, Operation: operation, EvidenceState: StateBlocked}, err
	}

	// Only a successful response from the real SARA service is REAL evidence.
	outputHash := hashAny(payload)
	result := HookResult{
		Hook:          hook,
		Operation:     operation,
		InputHash:     hashText(input),
		OutputHash:    outputHash,
		Payload:       payload,
		EvidenceState: StateReal,
	}

	// Preserve the actual correlation evidence returned by SARA when present.
	if raw, ok := payload["correlation_id"]; ok {
		if value, ok := raw.(string); ok && value != correlationID {
			return result, errors.New("GRCE_SARA_CORRELATION_MISMATCH")
		}
	}

	return result, nil
}

func (p *SARAParticipant) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"type": "SARAParticipant", "runtime": "HTTP"})
}

