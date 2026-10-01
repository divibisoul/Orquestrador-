package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/executor"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
)

// RegisterOrbitalReasoningOperations connects the deterministic TCE simulation
// boundary to the canonical Prefrontal admission boundary. TCE supplies
// resource/trajectory evidence; Prefrontal remains the decision authority.
func RegisterOrbitalReasoningOperations(e *Engine) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}

	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("N07_TCE_ENABLED")), "true")
	var sim *executor.SimulatedExecutor
	if enabled {
		cfg := core.DefaultConfig()
		cfg.Enabled = true
		if mode := strings.TrimSpace(os.Getenv("N07_TCE_MODE")); mode != "" {
			cfg.Mode = mode
		}
		if raw := strings.TrimSpace(os.Getenv("N07_TCE_EFFICIENCY")); raw != "" {
			if value, err := strconv.ParseFloat(raw, 64); err == nil {
				cfg.Simulation.EfficiencyFactor = value
			}
		}
		engine, err := core.NewEngine(cfg)
		if err != nil {
			return err
		}
		sim, err = executor.New(engine, models.DefaultCatalog(cfg), cfg.Mode)
		if err != nil {
			return err
		}
	}

	if err := e.Register("prefrontal.orbital.evaluate@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if sim == nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal.orbital", Target: m.Source, Status: "blocked", Error: "TCE_DISABLED"}, errors.New("transcendental compute simulation is disabled")
		}
		var workloads []core.Workload
		if err := json.Unmarshal([]byte(strings.TrimSpace(m.Metadata["workloads_json"])), &workloads); err != nil || len(workloads) == 0 {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal.orbital", Target: m.Source, Status: "rejected", Error: "WORKLOADS_REQUIRED"}, errors.New("metadata.workloads_json must contain at least one workload")
		}
		plan := core.Plan{Workloads: workloads, Strategy: strings.TrimSpace(m.Metadata["strategy"])}
		estimate, err := sim.EstimateCost(ctx, plan)
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal.orbital", Target: m.Source, Status: "error", Error: err.Error()}, err
		}

		var candidate struct {
			ID string
			Cost, Risk, Utility, Uncertainty, Urgency, Impact float64
		}
		if err := json.Unmarshal([]byte(strings.TrimSpace(m.Metadata["candidate_json"])), &candidate); err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal.orbital", Target: m.Source, Status: "rejected", Error: "CANDIDATE_REQUIRED"}, errors.New("metadata.candidate_json is required")
		}
		neocortex, err := prefrontal.NewNeocortex(e.cortex, e.neural)
		if err != nil {
			return protocol.Result{}, err
		}
		pfcCandidate, err := neocortex.Evaluate(ctx, candidate.ID, m.Payload, candidate.Risk, candidate.Cost, candidate.Urgency, candidate.Impact)
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal.orbital", Target: m.Source, Status: "rejected", Error: err.Error()}, err
		}
		decision, err := neocortex.Commit(pfcCandidate, "prefrontal-orbital-evaluation")
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.prefrontal.orbital", Target: m.Source, Status: "rejected", Error: err.Error()}, err
		}
		estimateJSON, _ := json.Marshal(estimate)
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.prefrontal.orbital", Target: m.Source, Status: "ok",
			Metadata: map[string]string{
				"decision_id": decision.ID,
				"decision_score": floatString(decision.Score),
				"simulation_architecture": estimate.Architecture,
				"simulation_estimated_ms": strconv.FormatInt(estimate.EstimatedTime.Milliseconds(), 10),
				"simulation_confidence": floatString(estimate.Confidence),
				"simulation_cost_json": string(estimateJSON),
				"simulation_is_not_hardware": "true",
			},
		}, nil
	}); err != nil {
		return err
	}

	return e.Register("transcendental.estimate@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if sim == nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.tce", Target: m.Source, Status: "blocked", Error: "TCE_DISABLED"}, errors.New("transcendental compute simulation is disabled")
		}
		var workloads []core.Workload
		if err := json.Unmarshal([]byte(strings.TrimSpace(m.Metadata["workloads_json"])), &workloads); err != nil || len(workloads) == 0 {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.tce", Target: m.Source, Status: "rejected", Error: "WORKLOADS_REQUIRED"}, errors.New("metadata.workloads_json must contain at least one workload")
		}
		estimate, err := sim.EstimateCost(ctx, core.Plan{Workloads: workloads, Strategy: strings.TrimSpace(m.Metadata["strategy"])})
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.tce", Target: m.Source, Status: "error", Error: err.Error()}, err
		}
		b, _ := json.Marshal(estimate)
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.tce", Target: m.Source, Status: "ok", Metadata: map[string]string{"cost_json": string(b), "architecture": estimate.Architecture, "simulation_only": "true"}}, nil
	}); err != nil {
		return err
	}
	return nil
}
