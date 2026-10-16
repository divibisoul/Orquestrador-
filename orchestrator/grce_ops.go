package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/grce"
	"github.com/divibisoul/Orquestrador-/grf"
	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	GRFDescribeOperation = "grf.describe@2.0.0"
	GRCEDescribeOperation = "grce.describe@2.0.0"
	GRCEBindingsOperation = "grce.bindings.describe@2.0.0"
	GRCECycleOperation = "grce.cycle.execute@2.0.0"
	GRCEAuthoritativeSaraOperation = "grce.sara.authoritative.execute@2.0.0"
	GRFParticipantDescribeOperation = "grf.participant.describe@2.0.0"
	GRFParticipantIngestOperation = "grf.participant.ingest@2.0.0"
)

type GRCEControlPlane struct {
	Executor *grce.Executor
}

type GRFSaraBridge interface {
	Configured() bool
	CycleWithContext(context.Context, string, string, string, map[string]any) (map[string]any, error)
}

func RegisterGRFGRCEOperations(e *Engine) error {
	return registerGRFGRCEOperationsWithExecutor(e, nil, nil)
}

func RegisterGRFGRCEOperationsWithSARA(e *Engine, bridge GRFSaraBridge) error {
	return registerGRFGRCEOperationsWithExecutor(e, bridge, nil)
}

func RegisterGRFGRCEOperationsWithExecutor(e *Engine, bridge GRFSaraBridge, executor *grce.Executor) error {
	return registerGRFGRCEOperationsWithExecutor(e, bridge, executor)
}

func registerGRFGRCEOperations(e *Engine, saraBridge GRFSaraBridge) error {
	return registerGRFGRCEOperationsWithExecutor(e, saraBridge, nil)
}

func registerGRFGRCEOperationsWithExecutor(e *Engine, saraBridge GRFSaraBridge, executor *grce.Executor) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}

	if err := e.Register(GRFDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if _, err := requireGRFContext(m); err != nil {
			return grfOperationFailure(m, err)
		}
		binding, err := readJSONFile("integrations/grf/system-binding.json")
		if err != nil {
			return grfOperationFailure(m, err)
		}
		fabric, err := readJSONFile("integrations/soul-29-capability-fabric.json")
		if err != nil {
			return grfOperationFailure(m, err)
		}
		_ = ctx
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.GRF", Target: m.Source, Status: "ok",
			Metadata: map[string]string{
				"framework": "GRF",
				"version": "2.0",
				"epistemic_state": string(grf.ACTIVE),
				"system_binding_json": binding,
				"soul29_fabric_json": fabric,
			},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(GRCEDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if _, err := requireGRFContext(m); err != nil {
			return grfOperationFailure(m, err)
		}
		invariants := grf.CanonicalInvariantSet()
		ids := grf.InvariantIDs(invariants)
		bindings := grceBindingInventory(e)
		hookReady := executor != nil
		reason := "GRCE runtime executor is connected to the declared SARA/SuperGPU/Nervo Vago boundaries; execution evidence is still required"
		if !hookReady {
			reason = "GRCE executor is not bound; operation remains fail-closed"
		}
		payload := map[string]any{
			"framework": "GRF",
			"executor": "GRCE",
			"version": "2.0",
			"invariants": ids,
			"hook_ready": hookReady,
			"epistemic_state": string(grf.PROJECTED),
			"reason": reason,
			"bindings": bindings,
		}
		raw, _ := json.Marshal(payload)
		_ = ctx
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.GRCE", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"grce_json": string(raw)},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(GRCEBindingsOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if _, err := requireGRFContext(m); err != nil {
			return grfOperationFailure(m, err)
		}
		raw, _ := json.Marshal(grceBindingInventory(e))
		_ = ctx
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.GRCE", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"bindings_json": string(raw), "epistemic_state": string(grf.PROJECTED)},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(GRFParticipantDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if _, err := requireGRFContext(m); err != nil {
			return grfOperationFailure(m, err)
		}
		raw, _ := json.Marshal(map[string]any{
			"participants": []string{"N01","N02","N03","N04","N05","N06","N07","SARA","JEV"},
			"new_external_contracts": 10,
			"original_soul25_preserved": true,
			"canonical_mesh": "soul-mesh/1",
			"epistemic_state": string(grf.PROJECTED),
		})
		_ = ctx
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.GRF", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"participant_json": string(raw)},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(GRFParticipantIngestOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		grfContext, err := requireGRFContext(m)
		if err != nil {
			return grfOperationFailure(m, err)
		}
		id := strings.TrimSpace(m.Metadata["participant_id"])
		if id == "" {
			return grfOperationFailure(m, errors.New("GRF_PARTICIPANT_ID_REQUIRED"))
		}
		registry, err := grf.ParticipantRegistry()
		if err != nil {
			return grfOperationFailure(m, err)
		}
		participant, ok := registry[id]
		if !ok {
			return grfOperationFailure(m, errors.New("GRF_PARTICIPANT_NOT_REGISTERED:"+id))
		}
		payload := map[string]any{}
		if raw := strings.TrimSpace(m.Metadata["grf_input_json"]); raw != "" {
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				return grfOperationFailure(m, errors.New("GRF_INPUT_JSON_INVALID"))
			}
		}
		input := grf.State{
			ID: strings.TrimSpace(m.Metadata["state_id"]),
			EpistemicState: grf.PRESERVED,
			Payload: payload,
		}
		if input.ID == "" {
			input.ID = "grf-input:" + m.TraceID
		}
		out, prov, evidence := participant.Ingest(input, grfContext)
		raw, _ := json.Marshal(map[string]any{"state":out,"provenance":prov,"evidence":evidence})
		_ = ctx
		return protocol.Result{
			TraceID:m.TraceID, CorrelationID:m.CorrelationID,
			Source:"N07.GRF", Target:m.Source, Status:"ok",
			Metadata:map[string]string{"participant_id":id,"participant_ingest_json":string(raw),"epistemic_state":string(evidence.State)},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(GRCEAuthoritativeSaraOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if _, err := requireGRFContext(m); err != nil {
			return grfOperationFailure(m, err)
		}
		if saraBridge == nil || !saraBridge.Configured() {
			return grfOperationFailure(m, errors.New("GRCE_SARA_BRIDGE_UNCONFIGURED"))
		}
		input := strings.TrimSpace(m.Metadata["grf_input_text"])
		if input == "" {
			input = strings.TrimSpace(m.Metadata["sara_input"])
		}
		if input == "" {
			return grfOperationFailure(m, errors.New("GRCE_SARA_INPUT_REQUIRED"))
		}
		grfContext, _ := requireGRFContext(m)
		contextPayload := map[string]any{
			"grf_version":"2.0",
			"trace_id":grfContext.TraceID,
			"correlation_id":grfContext.CorrelationID,
			"sequence_index":grfContext.SequenceIndex,
			"executor":"GRCE",
			"delegation":"SARA_AUTHORITATIVE",
		}
		cycleID := strings.TrimSpace(m.Metadata["sara_cycle_id"])
		out, err := saraBridge.CycleWithContext(ctx, input, cycleID, m.CorrelationID, contextPayload)
		if err != nil {
			return grfOperationFailure(m, errors.New("GRCE_SARA_CYCLE_FAILED:"+err.Error()))
		}
		converged, _ := out["converged"].(bool)
		traceHash := strings.TrimSpace(fmt.Sprint(out["trace_hash"]))
		federatedHash := strings.TrimSpace(fmt.Sprint(out["federated_context_hash"]))
		epistemic := grf.PROJECTED
		status := "degraded"
		if converged && traceHash != "" && federatedHash != "" {
			epistemic = grf.REAL
			status = "ok"
		}
		raw, _ := json.Marshal(out)
		return protocol.Result{
			TraceID:m.TraceID,CorrelationID:m.CorrelationID,Source:"N07.GRCE",Target:m.Source,Status:status,
			Metadata:map[string]string{
				"epistemic_state":string(epistemic),
				"execution_mode":"SARA_AUTHORITATIVE",
				"trace_hash":traceHash,
				"federated_context_hash":federatedHash,
				"sara_result_json":string(raw),
			},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(GRCECycleOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		grfContext, err := requireGRFContext(m)
		if err != nil {
			return grfOperationFailure(m, err)
		}
		if executor == nil {
			raw, _ := json.Marshal(grceBindingInventory(e))
			return protocol.Result{
				TraceID: m.TraceID, CorrelationID: m.CorrelationID,
				Source: "N07.GRCE", Target: m.Source, Status: "blocked",
				Error: "GRCE_RUNTIME_HOOKS_NOT_BOUND",
				Metadata: map[string]string{
					"epistemic_state": string(grf.BLOCKED),
					"bindings_json": string(raw),
				},
			}, errors.New("GRCE_RUNTIME_HOOKS_NOT_BOUND")
		}
		inputRaw := strings.TrimSpace(m.Metadata["grf_input_json"])
		payload := map[string]any{}
		if inputRaw != "" {
			if err := json.Unmarshal([]byte(inputRaw), &payload); err != nil {
				return grfOperationFailure(m, errors.New("GRCE_INPUT_JSON_INVALID"))
			}
		} else {
			textInput := strings.TrimSpace(m.Metadata["grf_input_text"])
			if textInput == "" {
				return grfOperationFailure(m, errors.New("GRCE_INPUT_REQUIRED"))
			}
			payload["text"] = textInput
		}
		state := grf.State{
			ID: strings.TrimSpace(m.Metadata["state_id"]),
			EpistemicState: grf.PRESERVED,
			Payload: payload,
		}
		if state.ID == "" {
			state.ID = "grce-input:" + m.TraceID
		}
		result, err := executor.Run(ctx, state, grfContext)
		raw, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return grfOperationFailure(m, marshalErr)
		}
		status := "degraded"
		if err == nil {
			status = "ok"
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.GRCE", Target: m.Source, Status: status,
			Metadata: map[string]string{
				"epistemic_state": string(result.EpistemicState),
				"execution_mode": "GRCE_EXECUTOR",
				"grce_result_json": string(raw),
			},
			Error: func() string {
				if err != nil { return err.Error() }
				return ""
			}(),
		}, err
	}); err != nil {
		return err
	}
	return nil
}

func requireGRFContext(m protocol.Message) (grf.Context, error) {
	if strings.TrimSpace(m.TraceID) == "" {
		return grf.Context{}, errors.New("GRF_TRACE_ID_REQUIRED")
	}
	if strings.TrimSpace(m.CorrelationID) == "" {
		return grf.Context{}, errors.New("GRF_CORRELATION_ID_REQUIRED")
	}
	raw := strings.TrimSpace(m.Metadata["grf_sequence_index"])
	if raw == "" {
		return grf.Context{}, errors.New("GRF_SEQUENCE_INDEX_REQUIRED")
	}
	sequence, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || sequence == 0 {
		return grf.Context{}, errors.New("GRF_SEQUENCE_INDEX_INVALID")
	}
	return grf.Context{TraceID:m.TraceID,CorrelationID:m.CorrelationID,SequenceIndex:sequence}, nil
}

func readJSONFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	return string(raw), nil
}

func grfOperationFailure(m protocol.Message, err error) (protocol.Result, error) {
	return protocol.Result{
		TraceID: m.TraceID, CorrelationID: m.CorrelationID,
		Source: "N07.GRF", Target: m.Source, Status: "blocked",
		Error: err.Error(),
	}, err
}

func grceBindingInventory(e *Engine) []map[string]any {
	type stage struct {
		name string
		owner string
		required []string
	}
	stages := []stage{
		{"L0", "canonical-soul-mesh", []string{"mesh.discovery@1.0.0","mesh.capability.resolve@1.0.0"}},
		{"L1", "SARA/ERU", []string{"sara.state@1.0.0","sara.trace@1.0.0"}},
		{"L2", "MMD", []string{"rgo.trinity.process@1.0.0"}},
		{"L3", "SARA/ARA", []string{"sara.regenerate@1.0.0"}},
		{"L4", "SARA/ETR/JEV", []string{"sara.audit@1.0.0","jev.systemone@1.0.0"}},
		{"L5", "SARA/ITR", []string{}},
		{"L6", "SARA/RGO", []string{"rgo.trinity.process@1.0.0"}},
		{"L7", "N07/GRCE", []string{GRCECycleOperation}},
	}
	out := make([]map[string]any, 0, len(stages))
	for _, item := range stages {
		missing := make([]string, 0)
		for _, op := range item.required {
			if !containsOperation(e.Operations(), op) {
				missing = append(missing, op)
			}
		}
		sort.Strings(missing)
		state := string(grf.REAL)
		if len(missing) > 0 || item.name == "L2" || item.name == "L5" {
			state = string(grf.PROJECTED)
		}
		out = append(out, map[string]any{
			"layer": item.name,
			"owner": item.owner,
			"required_operations": item.required,
			"missing_operations": missing,
			"state": state,
		})
	}
	return out
}

func containsOperation(ops []string, want string) bool {
	for _, op := range ops {
		if op == want {
			return true
		}
	}
	return false
}

func GRCEReadiness(e *Engine) map[string]any {
	bindings := grceBindingInventory(e)
	ready := true
	for _, item := range bindings {
		if item["layer"] == "L7" {
			continue
		}
		if item["state"] != string(grf.REAL) {
			ready = false
			break
		}
	}
	state := string(grf.PROJECTED)
	if ready {
		state = string(grf.ACTIVE)
	}
	return map[string]any{
		"ready": ready,
		"epistemic_state": state,
		"bindings": bindings,
	}
}
