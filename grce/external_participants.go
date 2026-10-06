package grce

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type ProviderEvidenceReporter interface {
	ProviderEvidence() []ProviderEvidence
}

// CompositeParticipant preserves the canonical SARA participant and adds
// explicitly configured external participants to their designated GRCE hooks.
type CompositeParticipant struct {
	primary     Participant
	auxiliaries []GoldenRuleParticipant
	mu          sync.Mutex
	evidence    []ProviderEvidence
}

func NewCompositeParticipant(primary Participant, auxiliaries ...GoldenRuleParticipant) *CompositeParticipant {
	filtered := make([]GoldenRuleParticipant, 0, len(auxiliaries))
	for _, p := range auxiliaries {
		if p != nil {
			filtered = append(filtered, p)
		}
	}
	return &CompositeParticipant{
		primary:     primary,
		auxiliaries: filtered,
		evidence:    make([]ProviderEvidence, 0, len(filtered)),
	}
}

func (c *CompositeParticipant) ExecuteHook(ctx context.Context, hook Hook, input, correlationID, cycleID string) (HookResult, error) {
	if c == nil || c.primary == nil {
		return HookResult{Hook: hook, EvidenceState: StateBlocked}, errors.New("GRCE_PRIMARY_PARTICIPANT_UNCONFIGURED")
	}

	primary, err := c.primary.ExecuteHook(ctx, hook, input, correlationID, cycleID)
	if err != nil {
		return primary, err
	}

	for _, provider := range c.auxiliaries {
		if staged, ok := provider.(interface{ Hook() Hook }); !ok || staged.Hook() != hook {
			continue
		}

		started := provider.ExecuteHook
		_ = started
		aux, auxErr := provider.ExecuteHook(ctx, hook, input, correlationID, cycleID)
		name := auxiliaryProviderName(provider)
		evidence := ProviderEvidence{
			Provider: name,
			Hook: hook,
			Operation: aux.Operation,
			CorrelationID: correlationID,
			CycleID: cycleID,
			ParentHash: primary.OutputHash,
			OutputHash: aux.OutputHash,
			State: StateReal,
		}
		if auxErr != nil {
			evidence.State = StateBlocked
			evidence.Error = auxErr.Error()
		}
		if aux.OutputHash == "" {
			evidence.OutputHash = hashAny(aux.Payload)
		}
		c.mu.Lock()
		c.evidence = append(c.evidence, evidence)
		c.mu.Unlock()
	}

	return primary, nil
}

func (c *CompositeParticipant) ProviderEvidence() []ProviderEvidence {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ProviderEvidence, len(c.evidence))
	copy(out, c.evidence)
	return out
}

type CommandGoldenRuleParticipant struct {
	id      string
	hook    Hook
	command string
}

func NewCommandGoldenRuleParticipant(id string, hook Hook, command string) *CommandGoldenRuleParticipant {
	return &CommandGoldenRuleParticipant{id: id, hook: hook, command: strings.TrimSpace(command)}
}

func (p *CommandGoldenRuleParticipant) ID() string   { return p.id }
func (p *CommandGoldenRuleParticipant) Hook() Hook   { return p.hook }

func (p *CommandGoldenRuleParticipant) ExecuteHook(ctx context.Context, hook Hook, input, correlationID, cycleID string) (HookResult, error) {
	result := HookResult{
		Hook: hook,
		Operation: p.id + ".golden-rule." + string(hook),
		InputHash: hashText(input),
		EvidenceState: StateBlocked,
	}
	if p == nil || strings.TrimSpace(p.command) == "" {
		return result, errors.New(p.id + "_COMMAND_NOT_CONFIGURED")
	}

	request := map[string]any{
		"provider":       p.id,
		"hook":           hook,
		"input":          input,
		"correlation_id": correlationID,
		"cycle_id":       cycleID,
	}
	raw, err := json.Marshal(request)
	if err != nil {
		return result, err
	}

	cmd := exec.CommandContext(ctx, "sh", "-lc", p.command)
	cmd.Stdin = bytes.NewReader(raw)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		result.Payload = map[string]any{
			"status": "BLOCKED",
			"stdout": strings.TrimSpace(stdout.String()),
			"stderr": strings.TrimSpace(stderr.String()),
		}
		result.OutputHash = hashAny(result.Payload)
		return result, err
	}

	out := strings.TrimSpace(stdout.String())
	payload := map[string]any{
		"status": "REAL",
		"stdout": out,
	}
	if out != "" && strings.HasPrefix(out, "{") {
		var decoded any
		if json.Unmarshal([]byte(out), &decoded) == nil {
			payload["result"] = decoded
		}
	}
	if errText := strings.TrimSpace(stderr.String()); errText != "" {
		payload["stderr"] = errText
	}
	result.Payload = payload
	result.OutputHash = hashAny(payload)
	result.EvidenceState = StateReal
	return result, nil
}

func ConfiguredExternalParticipants() []GoldenRuleParticipant {
	return []GoldenRuleParticipant{
		NewCommandGoldenRuleParticipant(
			"bijux-core",
			HookCharacterize,
			os.Getenv("SOUL_GRCE_BIJUX_COMMAND"),
		),
		NewCommandGoldenRuleParticipant(
			"ouro-loop",
			HookValidate,
			os.Getenv("SOUL_GRCE_OURO_LOOP_COMMAND"),
		),
		NewCommandGoldenRuleParticipant(
			"recuris",
			HookTrace,
			os.Getenv("SOUL_GRCE_RECURIS_COMMAND"),
		),
	}
}

func auxiliaryProviderName(participant GoldenRuleParticipant) string {
	if named, ok := participant.(interface{ ID() string }); ok {
		return named.ID()
	}
	return "external-provider"
}
