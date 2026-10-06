package grce

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// Executor runs the bounded Golden Rule cycle without duplicating SARA logic.
type Executor struct {
	participant Participant
}

func New(participant Participant) *Executor {
	return &Executor{participant: participant}
}

func (e *Executor) Execute(ctx context.Context, input, correlationID string) CycleResult {
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		correlationID = protocol.NewTraceID()
	}
	cycleID := correlationID
	result := CycleResult{
		CycleID:       cycleID,
		CorrelationID: correlationID,
		State:         StateBlocked,
		Stages:        make([]StageEvidence, 0, 6),
	}

	if e == nil || e.participant == nil {
		result.Error = "GRCE_SARA_PARTICIPANT_UNCONFIGURED"
		return result
	}
	input = strings.TrimSpace(input)
	if input == "" {
		result.Error = "GRCE_INPUT_REQUIRED"
		result.State = StateUnmeasurable
		return result
	}

	current := input
	parentHash := hashText(current)
	hooks := []Hook{
		HookDetect,
		HookCharacterize,
		HookRegenerate,
		HookValidate,
		HookFreeze,
		HookTrace,
	}

	for index, hook := range hooks {
		hookInput := current
		stage, err := e.participant.ExecuteHook(ctx, hook, hookInput, correlationID, cycleID)
		if err != nil {
			result.Error = err.Error()
			result.Stages = append(result.Stages, StageEvidence{
				Index: index + 1, Hook: hook, CorrelationID: correlationID,
				CycleID: cycleID, ParentHash: parentHash, OutputHash: "",
				State: StateBlocked,
			})
			return result
		}
		if stage.OutputHash == "" {
			stage.OutputHash = hashAny(stage.Payload)
		}
		stage.ParentHash = parentHash
		result.Stages = append(result.Stages, StageEvidence{
			Index: index + 1,
			Hook: hook,
			Operation: stage.Operation,
			CorrelationID: correlationID,
			CycleID: cycleID,
			ParentHash: parentHash,
			OutputHash: stage.OutputHash,
			State: stage.EvidenceState,
		})
		parentHash = stage.OutputHash

		if hook == HookRegenerate {
			if transformed, ok := stage.Payload["transformed"].(string); ok && strings.TrimSpace(transformed) != "" {
				current = transformed
			}
		}
	}

	result.FinalOutputHash = parentHash
	if reporter, ok := e.participant.(ProviderEvidenceReporter); ok {
		result.Providers = reporter.ProviderEvidence()
	}
	result.State = StateReal
	for _, stage := range result.Stages {
		if stage.State != StateReal {
			result.State = stage.State
			break
		}
	}
	return result
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hashAny(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

