package octacore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/executor"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type ExecutiveCore struct {
	engine   *orchestrator.Engine
	neocortex *prefrontal.Neocortex
	tce      *executor.SimulatedExecutor
	compute  *supergpu.Runtime
	peers    *mesh.PeerClient
	maxCollaborators int
}

type CoreStageReport struct {
	Name       string         `json:"name"`
	State      string         `json:"state"`
	DurationMS int64          `json:"duration_ms"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type AgentContribution struct {
	Capability string         `json:"capability"`
	Target     string         `json:"target,omitempty"`
	Source     string         `json:"source,omitempty"`
	State      string         `json:"state"`
	Output     map[string]any `json:"output,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type AgentCollaboration struct {
	Capability string         `json:"capability"`
	Payload    map[string]any `json:"payload"`
	Required   bool           `json:"required"`
}

type ExecutiveCoreRequest struct {
	JobID         string                 `json:"job_id"`
	CorrelationID string                 `json:"correlation_id"`
	TaskID        string                 `json:"task_id"`
	Operation     string                 `json:"operation"`
	Device        string                 `json:"device,omitempty"`
	Input         []float64              `json:"input"`
	Workloads     []core.Workload        `json:"workloads"`
	Risk          float64                `json:"risk"`
	Cost          float64                `json:"cost"`
	Urgency       float64                `json:"urgency"`
	Impact        float64                `json:"impact"`
	Collaborators []AgentCollaboration   `json:"collaborators,omitempty"`
}

type ExecutiveCoreResult struct {
	Processor          string             `json:"processor"`
	CoreName           string             `json:"core_name"`
	JobID              string             `json:"job_id"`
	CorrelationID      string             `json:"correlation_id"`
	State              string             `json:"state"`
	DecisionID         string             `json:"decision_id,omitempty"`
	Device             string             `json:"device,omitempty"`
	DeviceBackend      string             `json:"device_backend,omitempty"`
	HardwareAccelerated bool              `json:"hardware_accelerated"`
	TCEEnabled         bool               `json:"tce_enabled"`
	SimulationOnly     bool               `json:"simulation_only"`
	Output             []float64           `json:"output,omitempty"`
	Stages             []CoreStageReport   `json:"stages"`
	Agents             []AgentContribution `json:"agents,omitempty"`
}

func NewExecutiveCore(engine *orchestrator.Engine, neocortex *prefrontal.Neocortex, tce *executor.SimulatedExecutor, compute *supergpu.Runtime, peers *mesh.PeerClient) (*ExecutiveCore, error) {
	if engine == nil {
		return nil, errors.New("executive core requires orchestrator engine")
	}
	if neocortex == nil {
		return nil, errors.New("executive core requires prefrontal neocortex")
	}
	if compute == nil {
		return nil, errors.New("executive core requires SuperGPU runtime")
	}
	if peers == nil {
		return nil, errors.New("executive core requires Mesh peer client")
	}
	return &ExecutiveCore{
		engine: engine,
		neocortex: neocortex,
		tce: tce,
		compute: compute,
		peers: peers,
		maxCollaborators: 8,
	}, nil
}

func (c *ExecutiveCore) Name() string { return "N07.executive-octacore" }

func (c *ExecutiveCore) Describe() map[string]any {
	tce := map[string]any{"connected": c != nil && c.tce != nil, "simulation_only": true}
	if c != nil && c.tce != nil && c.tce.Engine != nil {
		tce["enabled"] = c.tce.Engine.Config.Enabled
		tce["mode"] = c.tce.Mode
		tce["parallel_units"] = c.tce.Engine.Config.EffectiveParallelUnits()
	} else {
		tce["enabled"] = false
	}
	return map[string]any{
		"name": c.Name(),
		"type": "unified_logical_processing_core",
		"topology": "4-components-x-2-lanes=8-logical-lanes",
		"components": []string{"orchestrator", "prefrontal_neocortex", "orbital_reasoning_tce", "supergpu"},
		"communication": "SOUL Mesh",
		"orchestrator_role": "route-validation-correlation-control",
		"prefrontal_role": "risk-admission-decision",
		"orbital_role": "deterministic-resource-estimation",
		"supergpu_role": "real-bounded-execution",
		"tce": tce,
		"external_agent_profiles": AgentProfiles(),
	}
}

func (c *ExecutiveCore) Health() map[string]any {
	if c == nil {
		return map[string]any{"status": "DEGRADED", "reason": "executive core unavailable"}
	}
	gpu := c.compute.Health()
	tceConnected := c.tce != nil && c.tce.Engine != nil
	status := "READY"
	if !tceConnected || !c.tce.Engine.Config.Enabled {
		status = "DEGRADED"
	}
	return map[string]any{
		"name": c.Name(),
		"status": status,
		"communication": "SOUL Mesh",
		"orchestrator": c.engine.Health(),
		"prefrontal_neocortex": c.neocortex.Health(),
		"orbital_reasoning_tce": map[string]any{
			"connected": tceConnected,
			"enabled": tceConnected && c.tce.Engine.Config.Enabled,
			"simulation_only": true,
		},
		"supergpu": gpu,
		"mesh_peer_count": len(c.peers.ConfiguredPeers()),
		"agent_profiles": AgentProfiles(),
	}
}

func (c *ExecutiveCore) Execute(ctx context.Context, req ExecutiveCoreRequest) (ExecutiveCoreResult, error) {
	started := time.Now()
	if c == nil {
		return ExecutiveCoreResult{}, errors.New("executive core unavailable")
	}
	if ctx == nil {
		return ExecutiveCoreResult{}, errors.New("context is nil")
	}
	if strings.TrimSpace(req.TaskID) == "" {
		req.TaskID = req.JobID
	}
	if err := validateCoreRequest(req); err != nil {
		return ExecutiveCoreResult{Processor: "Octacore", CoreName: c.Name(), JobID: req.JobID, CorrelationID: req.CorrelationID, State: "REJECTED"}, err
	}

	result := ExecutiveCoreResult{
		Processor: "Octacore",
		CoreName: c.Name(),
		JobID: req.JobID,
		CorrelationID: req.CorrelationID,
		State: "RUNNING",
		TCEEnabled: c.tce != nil && c.tce.Engine != nil && c.tce.Engine.Config.Enabled,
		SimulationOnly: true,
		Stages: make([]CoreStageReport, 0, 8),
	}

	stage := func(name string, fn func() (map[string]any, error)) error {
		after := time.Now()
		meta, err := fn()
		report := CoreStageReport{Name: name, State: "REAL", DurationMS: time.Since(after).Milliseconds(), Metadata: meta}
		if err != nil {
			report.State = classifyCoreStageError(err)
			report.Error = err.Error()
			result.Stages = append(result.Stages, report)
			return err
		}
		result.Stages = append(result.Stages, report)
		return nil
	}

	innerOperations := []string{
		"prefrontal.admission@1.0.0",
		"transcendental.estimate@1.0.0",
		"supergpu.execute@1.0.0",
	}
	if err := stage("orchestrator.resolve", func() (map[string]any, error) {
		resolved := make([]string, 0, len(innerOperations))
		for _, operation := range innerOperations {
			message := protocol.NewMessage(protocol.N07, protocol.N07, "command", operation, nil)
			if _, err := c.engine.Route(message); err != nil {
				return nil, fmt.Errorf("operation %s unavailable: %w", operation, err)
			}
			resolved = append(resolved, operation)
		}
		return map[string]any{"resolved_operations": resolved, "correlation_id": req.CorrelationID}, nil
	}); err != nil {
		result.State = "REJECTED"
		return result, err
	}

	agentResults, agentErr := c.runCollaborators(ctx, req)
	result.Agents = agentResults
	if agentErr != nil {
		result.State = "REJECTED"
		return result, agentErr
	}

	if err := stage("orbital.reasoning.estimate", func() (map[string]any, error) {
		if c.tce == nil || c.tce.Engine == nil || !c.tce.Engine.Config.Enabled {
			return nil, errors.New("TCE_DISABLED")
		}
		estimate, err := c.tce.EstimateCost(ctx, core.Plan{Workloads: req.Workloads, Strategy: "auto"})
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"architecture": estimate.Architecture,
			"estimated_time_ms": estimate.EstimatedTime.Milliseconds(),
			"estimated_memory_gb": estimate.EstimatedMemoryGB,
			"confidence": estimate.Confidence,
			"simulation_only": true,
		}, nil
	}); err != nil {
		result.State = "BLOCKED"
		return result, err
	}

	var decision prefrontal.Decision
	if err := stage("prefrontal.neocortex.admission", func() (map[string]any, error) {
		candidate, err := c.neocortex.Evaluate(ctx, req.TaskID, req.Input, req.Risk, req.Cost, req.Urgency, req.Impact)
		if err != nil {
			return nil, err
		}
		if err := c.neocortex.UpdateWorkingMemory([]prefrontal.Candidate{candidate}); err != nil {
			return nil, err
		}
		decision, err = c.neocortex.Commit(candidate, "octacore-integrated-admission")
		if err != nil {
			return nil, err
		}
		return map[string]any{"decision_id": decision.ID, "score": decision.Score, "neural_dimensions": candidate.Context["neural_dimensions"]}, nil
	}); err != nil {
		result.State = "BLOCKED"
		return result, err
	}
	result.DecisionID = decision.ID

	var output []float64
	var device supergpu.Device
	if err := stage("supergpu.execution", func() (map[string]any, error) {
		var err error
		device, err = c.compute.Select(req.Device)
		if err != nil {
			return nil, err
		}
		if err := c.compute.Reserve(device.ID, req.JobID); err != nil {
			return nil, err
		}
		defer c.compute.Release(device.ID, req.JobID)
		output, err = c.compute.Execute(supergpu.WithCorrelationID(ctx, req.CorrelationID), device, req.Operation, req.Input)
		if err != nil {
			return nil, err
		}
		return map[string]any{"device": device.ID, "backend": device.Backend, "hardware_accelerated": device.Backend != "cpu", "output_size": len(output)}, nil
	}); err != nil {
		result.State = "BLOCKED"
		return result, err
	}

	result.Output = output
	result.Device = device.ID
	result.DeviceBackend = device.Backend
	result.HardwareAccelerated = device.Backend != "cpu"
	result.State = "READY"
	_ = started
	return result, nil
}

func (c *ExecutiveCore) runCollaborators(ctx context.Context, req ExecutiveCoreRequest) ([]AgentContribution, error) {
	if len(req.Collaborators) == 0 {
		return nil, nil
	}
	if len(req.Collaborators) > c.maxCollaborators {
		return nil, errors.New("collaborator count exceeds executive core limit")
	}
	results := make([]AgentContribution, len(req.Collaborators))
	var wg sync.WaitGroup
	var firstRequiredErr error
	var errMu sync.Mutex

	for i, collaboration := range req.Collaborators {
		i := i
		collaboration := collaboration
		wg.Add(1)
		go func() {
			defer wg.Done()
			contribution := AgentContribution{Capability: strings.TrimSpace(collaboration.Capability), State: "BLOCKED"}
			profile := FindAgentProfile(contribution.Capability)
			if profile != nil {
				contribution.Source = profile.Source
			}
			if contribution.Capability == "" || collaboration.Payload == nil {
				contribution.State = "REJECTED"
				contribution.Error = "capability and payload are required"
				if collaboration.Required {
					errMu.Lock()
					if firstRequiredErr == nil {
						firstRequiredErr = errors.New(contribution.Error)
					}
					errMu.Unlock()
				}
				results[i] = contribution
				return
			}
			if profile != nil && strings.TrimSpace(profile.AdapterOwner) != "" {
				out, err := c.peers.CallWithCorrelation(ctx, profile.AdapterOwner, contribution.Capability, collaboration.Payload, req.CorrelationID)
				if err != nil {
					contribution.State = "UNMEASURABLE"
					contribution.Target = profile.AdapterOwner
					contribution.Error = err.Error()
					if collaboration.Required {
						errMu.Lock()
						if firstRequiredErr == nil {
							firstRequiredErr = err
						}
						errMu.Unlock()
					}
					results[i] = contribution
					return
				}
				contribution.State = "REAL"
				contribution.Target = profile.AdapterOwner
				contribution.Output = out
				results[i] = contribution
				return
			}

			out, target, err := c.peers.CallBestDynamic(ctx, contribution.Capability, collaboration.Payload, req.CorrelationID)
			if err != nil {
				contribution.State = "UNMEASURABLE"
				contribution.Error = err.Error()
				if collaboration.Required {
					errMu.Lock()
					if firstRequiredErr == nil {
						firstRequiredErr = err
					}
					errMu.Unlock()
				}
				results[i] = contribution
				return
			}
			contribution.State = "REAL"
			contribution.Target = target
			contribution.Output = out
			results[i] = contribution
		}()
	}
	wg.Wait()
	return results, firstRequiredErr
}

func validateCoreRequest(req ExecutiveCoreRequest) error {
	if strings.TrimSpace(req.JobID) == "" {
		return errors.New("job_id is required")
	}
	if strings.TrimSpace(req.CorrelationID) == "" {
		return errors.New("correlation_id is required")
	}
	if strings.TrimSpace(req.TaskID) == "" {
		return errors.New("task_id is required")
	}
	if strings.TrimSpace(req.Operation) == "" {
		return errors.New("operation is required")
	}
	if len(req.Input) == 0 {
		return errors.New("input is required")
	}
	if len(req.Workloads) == 0 {
		return errors.New("at least one TCE workload is required")
	}
	for _, v := range req.Input {
		if v != v || v > 1e308 || v < -1e308 {
			return errors.New("input contains non-finite value")
		}
	}
	return nil
}

func classifyCoreStageError(err error) string {
	switch {
	case strings.Contains(err.Error(), "TCE_DISABLED"):
		return "BLOCKED"
	case strings.Contains(err.Error(), "unavailable"), strings.Contains(err.Error(), "not configured"):
		return "UNMEASURABLE"
	default:
		return "ERROR"
	}
}

func DecodeExecutiveCoreRequest(metadata map[string]string) (ExecutiveCoreRequest, error) {
	raw := strings.TrimSpace(metadata["octacore_core_request_json"])
	if raw == "" {
		return ExecutiveCoreRequest{}, errors.New("metadata.octacore_core_request_json is required")
	}
	var req ExecutiveCoreRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return ExecutiveCoreRequest{}, err
	}
	if req.CorrelationID == "" {
		req.CorrelationID = strings.TrimSpace(metadata["correlation_id"])
	}
	return req, validateCoreRequest(req)
}

func EnsureCoreCorrelation(ctx context.Context, correlation string) context.Context {
	return supergpu.WithCorrelationID(ctx, correlation)
}
