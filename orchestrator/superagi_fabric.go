package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	SuperAGIFabricDescribeOperation = "superagi.fabric.describe@1.0.0"
	SuperAGIFabricExecuteOperation  = "superagi.fabric.execute@1.0.0"
)

type SuperAGIFabric struct {
	engine *Engine
}

func NewSuperAGIFabric(engine *Engine) (*SuperAGIFabric, error) {
	if engine == nil {
		return nil, errors.New("orchestrator engine is required")
	}
	return &SuperAGIFabric{engine: engine}, nil
}

func (f *SuperAGIFabric) Describe() map[string]any {
	return map[string]any{
		"name":            "SOUL SuperAGI Control Fabric",
		"nature":          "composed-control-plane",
		"claim":           "architecture-composition; not proof of general intelligence",
		"canonicalOwner":  "N07",
		"mesh":            "soul-mesh/1",
		"contractVersion": protocol.SoulMeshContractVersion,
		"components": []string{
			"native-nuclei-N01-N06",
			"N07-Mesh",
			"SuperGPU",
			"MultiAgentFacade",
			"AgentArsenal",
			"Octacore",
			"Prefrontal",
			"SARA",
			"JEV",
		},
		"executionRule": "agent-stage -> optional compute-stage -> result-validation; accelerator requirement is explicit",
		"failClosed":    true,
		"computePolicy": "optional; CPU is valid unless require_accelerator=true",
	}
}

func RegisterSuperAGIFabricOperations(e *Engine) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if _, err := NewSuperAGIFabric(e); err != nil {
		return err
	}

	if err := e.Register(SuperAGIFabricDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		fabric := NewSuperAGIFabricMust(e)
		raw, err := json.Marshal(fabric.Describe())
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.superagi", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"fabric_json": string(raw)},
		}, nil
	}); err != nil {
		return err
	}

	return e.Register(SuperAGIFabricExecuteOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		provider := strings.TrimSpace(m.Metadata["agent_provider"])
		goal := strings.TrimSpace(m.Metadata["goal"])
		computeOperation := strings.TrimSpace(m.Metadata["compute_operation"])
		valuesRaw := strings.TrimSpace(m.Metadata["compute_values_json"])
		requireAccelerator := strings.EqualFold(strings.TrimSpace(m.Metadata["require_accelerator"]), "true")
		computeRequested := computeOperation != "" || valuesRaw != "" || requireAccelerator
		if provider == "" || goal == "" {
			return superAGIFabricFail(m, "SUPERAGI_AGENT_STAGE_REQUIRED")
		}
		if computeRequested && (computeOperation == "" || valuesRaw == "") {
			return superAGIFabricFail(m, "SUPERAGI_COMPUTE_STAGE_REQUIRED")
		}

		agentMode := strings.ToLower(strings.TrimSpace(m.Metadata["agent_mode"]))
		var agentResult protocol.Result
		var agentErr error
		switch agentMode {
		case "", "multiagent-facade":
			agentMetadata := map[string]string{
				"provider":  provider,
				"goal":      goal,
				"roles":     m.Metadata["roles"],
				"maxRounds": m.Metadata["maxRounds"],
			}
			agentMessage := protocol.Propagate(m, "N07.superagi", "N07", MultiAgentExecuteOperation, nil)
			agentMessage.Metadata = agentMetadata
			agentResult, agentErr = fexecute(ctx, e, agentMessage)
		case "agent-arsenal":
			source := strings.TrimSpace(m.Metadata["agent_source"])
			artifactPath := strings.TrimSpace(m.Metadata["agent_path"])
			if source == "" || artifactPath == "" {
				return superAGIFabricFail(m, "SUPERAGI_AGENT_ARSENAL_BOUNDARY_REQUIRED")
			}
			agentMessage := protocol.Propagate(m, "N07.superagi", "N07", "agent.arsenal.activate@1.0.0", nil)
			agentMessage.Metadata = map[string]string{
				"source": source,
				"path":   artifactPath,
				"task":   goal,
			}
			agentResult, agentErr = fexecute(ctx, e, agentMessage)
		default:
			return superAGIFabricFail(m, "SUPERAGI_AGENT_MODE_UNSUPPORTED")
		}
		if agentErr != nil {
			return superAGIFabricFail(m, "SUPERAGI_AGENT_STAGE_FAILED:"+agentResult.Error)
		}

		if !computeRequested {
			return protocol.Result{
				TraceID: m.TraceID, CorrelationID: m.CorrelationID,
				Source: "N07.superagi", Target: m.Source, Status: "ok",
				Metadata: map[string]string{
					"agent_stage":         "PASS",
					"compute_stage":       "SKIPPED",
					"compute_requested":   "false",
					"accelerator":         "false",
					"provider":            provider,
					"agent_result_json":   agentResult.Metadata["result_json"],
				},
			}, nil
		}

		var values []float64
		if err := json.Unmarshal([]byte(valuesRaw), &values); err != nil || len(values) == 0 {
			return superAGIFabricFail(m, "SUPERAGI_COMPUTE_VALUES_INVALID")
		}
		computeMessage := protocol.Propagate(m, m.Source, "N07", "mesh.supergpu.execute@1.0.0", values)
		computeMessage.Metadata = map[string]string{
			"operation":           computeOperation,
			"device":              m.Metadata["device"],
			"require_accelerator": strconv.FormatBool(requireAccelerator),
		}
		computeResult, computeErr := fexecute(ctx, e, computeMessage)
		if computeErr != nil {
			return superAGIFabricFail(m, "SUPERAGI_COMPUTE_STAGE_FAILED:"+computeResult.Error)
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.superagi", Target: m.Source, Status: "ok",
			Payload: computeResult.Payload,
			Metadata: map[string]string{
				"agent_stage":       "PASS",
				"compute_stage":     "PASS",
				"compute_requested": "true",
				"device":            computeResult.Metadata["device"],
				"backend":           computeResult.Metadata["backend"],
				"accelerator":       computeResult.Metadata["accelerator"],
				"provider":          provider,
				"agent_result_json": agentResult.Metadata["result_json"],
			},
		}, nil
	})
}

func fexecute(ctx context.Context, e *Engine, m protocol.Message) (protocol.Result, error) {
	handler, err := e.Route(m)
	if err != nil {
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.superagi", Target: m.Source, Status: "error", Error: err.Error()}, err
	}
	return handler(ctx, m)
}

func superAGIFabricFail(m protocol.Message, code string) (protocol.Result, error) {
	err := errors.New(code)
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.superagi", Target: m.Source, Status: "error", Error: code}, err
}

func NewSuperAGIFabricMust(e *Engine) *SuperAGIFabric {
	return &SuperAGIFabric{engine: e}
}
