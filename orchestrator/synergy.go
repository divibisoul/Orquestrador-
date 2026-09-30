package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const FavoriteSynergySequenceID = "diego.favorite.n01-n05-sara-n02-n03-n06-n04-n07"

type SynergyNode struct {
	Target       string   `json:"target"`
	Plane        string   `json:"plane"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type SynergyEdge struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Composition string `json:"composition"`
	Transport   string `json:"transport"`
}

type SynergySequence struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Purpose     string        `json:"purpose"`
	Nodes       []SynergyNode `json:"nodes"`
	Edges       []SynergyEdge `json:"edges"`
	Preservation []string     `json:"preservation"`
	Status      string        `json:"status"`
}

// FavoriteSynergySequence models a composition route, not a replacement
// topology. The canonical N01..N07 adjacency graph remains authoritative.
// SARA is G0/system control and is therefore represented as a cross-plane node.
func FavoriteSynergySequence() SynergySequence {
	nodes := []SynergyNode{
		{Target: "N01", Plane: "soul-mesh", Role: "reference-gateway-and-context"},
		{Target: "N05", Plane: "soul-mesh", Role: "dispatch-inference-execution"},
		{Target: "SARA", Plane: "g0-regenerative", Role: "audit-regeneration-governance"},
		{Target: "N02", Plane: "soul-mesh", Role: "conversation-generation-cognition"},
		{Target: "N03", Plane: "soul-mesh", Role: "audio-speech-multimodal-perception"},
		{Target: "N06", Plane: "soul-mesh", Role: "cognition-synthesis-audit"},
		{Target: "N04", Plane: "soul-mesh", Role: "tools-documents-artifacts"},
		{Target: "N07", Plane: "soul-mesh", Role: "orchestration-neural-compute"},
	}
	edges := make([]SynergyEdge, 0, len(nodes)-1)
	compositions := []string{
		"context→dispatch",
		"dispatch→governance",
		"governance→generation",
		"generation→perception",
		"perception→cognition",
		"cognition↔tooling",
		"tooling→orchestration",
	}
	for i := 0; i < len(nodes)-1; i++ {
		from, to := nodes[i].Target, nodes[i+1].Target
		transport := "Soul Mesh"
		if to == "SARA" || from == "SARA" {
			transport = "N07→SARA proxy"
		}
		edges = append(edges, SynergyEdge{From: from, To: to, Composition: compositions[i], Transport: transport})
	}
	return SynergySequence{
		ID: FavoriteSynergySequenceID,
		Name: "N01 → N05 → SARA → N02 → N03 → (N06 ↔ N04) → N07",
		Purpose: "compose specialized capabilities across independent runtimes instead of treating the federation as a single linear processor",
		Nodes: nodes,
		Edges: edges,
		Preservation: []string{
			"canonical seven-nucleus Mesh topology",
			"native capability ownership of N01..N07",
			"SARA as G0/non-nucleus authority",
			"existing pairwise and adjacent fusion paths",
		},
		Status: "STRUCTURALLY_DEFINED",
	}
}

func ValidateSynergySequence(sequence SynergySequence) error {
	if strings.TrimSpace(sequence.ID) == "" {
		return errors.New("synergy sequence id is required")
	}
	if len(sequence.Nodes) < 2 || len(sequence.Edges) != len(sequence.Nodes)-1 {
		return errors.New("synergy sequence must have n nodes and n-1 edges")
	}
	allowed := map[string]bool{
		"N01": true, "N02": true, "N03": true, "N04": true,
		"N05": true, "N06": true, "N07": true, "SARA": true,
	}
	for _, node := range sequence.Nodes {
		if !allowed[strings.TrimSpace(node.Target)] {
			return fmt.Errorf("unsupported synergy node: %s", node.Target)
		}
	}
	for i, edge := range sequence.Edges {
		if edge.From != sequence.Nodes[i].Target || edge.To != sequence.Nodes[i+1].Target {
			return fmt.Errorf("synergy edge %d does not preserve declared node order", i)
		}
		if strings.TrimSpace(edge.Transport) == "" {
			return fmt.Errorf("synergy edge %d transport is required", i)
		}
	}
	return nil
}

// ExecuteSynergyRoute is deliberately generic. Each stage receives the previous
// stage output in the caller's envelope; capability-specific payload adaptation
// remains owned by the capability provider and is never guessed here.
type SynergyInvoker interface {
	Invoke(context.Context, string, string, map[string]any, string) (map[string]any, error)
}

type SynergyInvokerFunc func(context.Context, string, string, map[string]any, string) (map[string]any, error)

func (f SynergyInvokerFunc) Invoke(ctx context.Context, target, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	return f(ctx, target, capability, payload, correlation)
}

type SynergyTraceStep struct {
	Target        string         `json:"target"`
	Capability    string         `json:"capability"`
	CorrelationID string         `json:"correlation_id"`
	Status        string         `json:"status"`
	StartedAt     time.Time      `json:"started_at"`
	DurationMs    int64          `json:"duration_ms"`
	Output        map[string]any `json:"output,omitempty"`
	Error         string         `json:"error,omitempty"`
}

type SynergyExecution struct {
	Sequence      SynergySequence  `json:"sequence"`
	CorrelationID string            `json:"correlation_id"`
	Trace         []SynergyTraceStep `json:"trace"`
	FinalTarget   string            `json:"final_target"`
	Status        string            `json:"status"`
	FinalOutput   map[string]any    `json:"final_output,omitempty"`
}

func ExecuteSynergyRoute(ctx context.Context, sequence SynergySequence, invoker SynergyInvoker, correlation string, capabilities map[string]string, payload map[string]any) (SynergyExecution, error) {
	if ctx == nil {
		return SynergyExecution{}, errors.New("context is nil")
	}
	if invoker == nil {
		return SynergyExecution{}, errors.New("synergy invoker is not configured")
	}
	if correlation = strings.TrimSpace(correlation); correlation == "" {
		return SynergyExecution{}, errors.New("correlation is required")
	}
	if err := ValidateSynergySequence(sequence); err != nil {
		return SynergyExecution{}, err
	}

	current := cloneSynergyPayload(payload)
	trace := make([]SynergyTraceStep, 0, len(sequence.Nodes))
	for index, node := range sequence.Nodes {
		start := time.Now().UTC()
		target := node.Target
		capability := strings.TrimSpace(capabilities[target])
		if target == "N07" {
			trace = append(trace, SynergyTraceStep{
				Target: target, Capability: capability, CorrelationID: correlation,
				Status: "LOCAL_FINALIZE", StartedAt: start,
				DurationMs: time.Since(start).Milliseconds(), Output: current,
			})
			return SynergyExecution{
				Sequence: sequence, CorrelationID: correlation, Trace: trace,
				FinalTarget: target, Status: "ok", FinalOutput: current,
			}, nil
		}
		if capability == "" {
			return SynergyExecution{
				Sequence: sequence, CorrelationID: correlation, Trace: trace,
				FinalTarget: target, Status: "blocked",
			}, fmt.Errorf("synergy capability missing for %s", target)
		}
		current["synergy_sequence_id"] = sequence.ID
		current["synergy_stage"] = index
		current["synergy_target"] = target
		current["previous_output"] = current["last_output"]

		output, err := invoker.Invoke(ctx, target, capability, current, correlation)
		step := SynergyTraceStep{
			Target: target, Capability: capability, CorrelationID: correlation,
			StartedAt: start, DurationMs: time.Since(start).Milliseconds(),
		}
		if err != nil {
			step.Status = "error"
			step.Error = err.Error()
			trace = append(trace, step)
			return SynergyExecution{
				Sequence: sequence, CorrelationID: correlation, Trace: trace,
				FinalTarget: target, Status: "error",
			}, err
		}
		step.Status = "ok"
		step.Output = output
		trace = append(trace, step)
		current["last_output"] = output
		current["previous_target"] = target
	}
	return SynergyExecution{Sequence: sequence, CorrelationID: correlation, Trace: trace, FinalTarget: "N07", Status: "ok", FinalOutput: current}, nil
}

func cloneSynergyPayload(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+8)
	for key, value := range in {
		out[key] = value
	}
	return out
}
