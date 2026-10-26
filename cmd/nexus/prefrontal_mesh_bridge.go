package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

const n01PrefrontalOptimizeCapability = "neural.prefrontal.optimize@1.0.0"
const n01PrefrontalOptimizerID = "N01.neocortex-prefrontal-optimizer"

// registerN01PrefrontalBridge makes the N07 specialist use the canonical authenticated
// Mesh peer client; it never assumes N01 is reachable or fabricates a successful result.
func registerN01PrefrontalBridge(engine *orchestrator.Engine, peers *mesh.PeerClient) error {
	if engine == nil || peers == nil {
		return errors.New("N01_PREFRONTAL_BRIDGE_DEPENDENCIES_REQUIRED")
	}
	return engine.Register(n01PrefrontalOptimizeCapability, func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.prefrontal-bridge", Target: message.Source, Status: "cancelled", Error: err.Error()}, err
		}
		description := strings.TrimSpace(message.Metadata["goal_description"])
		if description == "" {
			description = strings.TrimSpace(message.Metadata["description"])
		}
		if description == "" {
			return n01PrefrontalBridgeError(message, errors.New("N01_PREFRONTAL_GOAL_DESCRIPTION_REQUIRED"))
		}
		weakPoint := strings.TrimSpace(message.Metadata["weak_point"])
		if weakPoint == "" {
			weakPoint = description
		}
		priority := 0.75
		if raw := strings.TrimSpace(message.Metadata["priority"]); raw != "" {
			parsed, err := strconv.ParseFloat(raw, 64)
			if err != nil || parsed < 0 || parsed > 1 {
				return n01PrefrontalBridgeError(message, errors.New("N01_PREFRONTAL_PRIORITY_MUST_BE_BETWEEN_0_AND_1"))
			}
			priority = parsed
		}
		goalID := strings.TrimSpace(message.Metadata["goal_id"])
		if goalID == "" {
			goalID = "n07-prefrontal:" + message.CorrelationID
		}
		contextValue := map[string]any{
			"requestedByAgent":  strings.TrimSpace(message.Metadata["agent_id"]),
			"parentTraceID":     strings.TrimSpace(message.Metadata["parent_trace_id"]),
			"n07ChildTraceID":   message.TraceID,
			"correlationID":     message.CorrelationID,
			"payloadVector":     append([]float64(nil), message.Payload...),
			"operationRequested": n01PrefrontalOptimizeCapability,
		}
		if raw := strings.TrimSpace(message.Metadata["context_json"]); raw != "" {
			var extra any
			if err := json.Unmarshal([]byte(raw), &extra); err != nil {
				return n01PrefrontalBridgeError(message, fmt.Errorf("N01_PREFRONTAL_CONTEXT_JSON_INVALID: %w", err))
			}
			contextValue["requestContext"] = extra
		}
		payload := map[string]any{
			"goalId":     goalID,
			"description": description,
			"weakPoint": weakPoint,
			"priority":   priority,
			"context":    contextValue,
		}
		response, err := peers.CallWithCorrelation(ctx, protocol.N01, n01PrefrontalOptimizeCapability, payload, message.CorrelationID)
		if err != nil {
			return n01PrefrontalBridgeError(message, fmt.Errorf("N01_PREFRONTAL_MESH_DELEGATION_FAILED: %w", err))
		}
		if response["kind"] == "error" {
			return n01PrefrontalBridgeError(message, errors.New("N01_PREFRONTAL_REMOTE_RETURNED_MESH_ERROR"))
		}
		remote, ok := response["payload"].(map[string]any)
		if !ok {
			return n01PrefrontalBridgeError(message, errors.New("N01_PREFRONTAL_RESPONSE_PAYLOAD_INVALID"))
		}
		if remote["status"] != "completed" || remote["agentId"] != n01PrefrontalOptimizerID {
			return n01PrefrontalBridgeError(message, fmt.Errorf("N01_PREFRONTAL_EXECUTION_NOT_CONFIRMED: status=%v agent=%v", remote["status"], remote["agentId"]))
		}
		if remote["correlationId"] != message.CorrelationID {
			return n01PrefrontalBridgeError(message, errors.New("N01_PREFRONTAL_RESPONSE_CORRELATION_MISMATCH"))
		}
		execution, ok := remote["execution"].(map[string]any)
		if !ok || execution["nucleus"] != protocol.N01 || execution["backend"] != "IN_PROCESS" {
			return n01PrefrontalBridgeError(message, errors.New("N01_PREFRONTAL_EXECUTION_EVIDENCE_INVALID"))
		}
		rawResult, err := json.Marshal(remote)
		if err != nil {
			return n01PrefrontalBridgeError(message, err)
		}
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.prefrontal-bridge", Target: message.Source, Status: "ok",
			Metadata: map[string]string{
				"n01_agent":          n01PrefrontalOptimizerID,
				"n01_capability":     n01PrefrontalOptimizeCapability,
				"n01_result_json":    string(rawResult),
				"n01_trace_id":       fmt.Sprint(remote["traceId"]),
				"n01_correlation_id": fmt.Sprint(remote["correlationId"]),
				"n01_execution_ms":   fmt.Sprint(execution["latencyMs"]),
				"n07_parent_trace_id": strings.TrimSpace(message.Metadata["parent_trace_id"]),
			},
		}, nil
	})
}

func n01PrefrontalBridgeError(message protocol.Message, err error) (protocol.Result, error) {
	return protocol.Result{
		TraceID: message.TraceID, CorrelationID: message.CorrelationID,
		Source: "N07.prefrontal-bridge", Target: message.Source, Status: "error", Error: err.Error(),
	}, err
}
