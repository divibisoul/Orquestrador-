package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

const (
	SuperGPUMeshDescribeOperation = "mesh.supergpu.describe@1.0.0"
	SuperGPUMeshExecuteOperation  = "mesh.supergpu.execute@1.0.0"
	SuperGPUMeshParallelOperation = "mesh.supergpu.parallel@1.0.0"
)

func RegisterSuperGPUMeshOperations(e *Engine, federation *supergpu.Federation) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if federation == nil {
		return errors.New("supergpu federation is required")
	}

	if err := e.Register(SuperGPUMeshDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		raw, err := json.Marshal(federation.Health())
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.supergpu.mesh", Target: m.Source, Status: "ok",
			Metadata: map[string]string{"health_json": string(raw)},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(SuperGPUMeshExecuteOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if !isFederatedNucleus(m.Source) {
			return superGPUMeshFail(m, "SUPERGPU_MESH_SOURCE_INVALID")
		}
		op := strings.TrimSpace(m.Metadata["operation"])
		if op == "" {
			return superGPUMeshFail(m, "SUPERGPU_OPERATION_REQUIRED")
		}
		requireAccelerator := !strings.EqualFold(strings.TrimSpace(m.Metadata["require_accelerator"]), "false")
		fResult, err := federation.Execute(ctx, supergpu.FederatedRequest{
			Nucleus: m.Source, Operation: op, Payload: append([]float64(nil), m.Payload...),
			Device: strings.TrimSpace(m.Metadata["device"]), RequireAccelerator: requireAccelerator,
		})
		if err != nil {
			return superGPUMeshFail(m, classifySuperGPUError(err))
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.supergpu.mesh", Target: m.Source, Status: "ok",
			Payload: fResult.Output,
			Metadata: map[string]string{
				"device": fResult.Device.ID,
				"backend": fResult.Device.Backend,
				"accelerator": strconv.FormatBool(fResult.Device.Backend != "cpu"),
				"operation": op,
			},
		}, nil
	}); err != nil {
		return err
	}

	return e.Register(SuperGPUMeshParallelOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if !isFederatedNucleus(m.Source) {
			return superGPUMeshFail(m, "SUPERGPU_MESH_SOURCE_INVALID")
		}
		op := strings.TrimSpace(m.Metadata["operation"])
		if op == "" {
			return superGPUMeshFail(m, "SUPERGPU_OPERATION_REQUIRED")
		}
		var inputs [][]float64
		if err := json.Unmarshal([]byte(m.Metadata["inputs_json"]), &inputs); err != nil || len(inputs) == 0 {
			return superGPUMeshFail(m, "SUPERGPU_BATCH_INPUTS_INVALID")
		}
		workers := 1
		if raw := strings.TrimSpace(m.Metadata["workers"]); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				workers = parsed
			}
		}
		requireAccelerator := !strings.EqualFold(strings.TrimSpace(m.Metadata["require_accelerator"]), "false")
		result, err := federation.ExecuteBatch(ctx, supergpu.FederatedBatchRequest{
			Nucleus: m.Source, Operation: op, Inputs: inputs, Device: strings.TrimSpace(m.Metadata["device"]),
			Workers: workers, RequireAccelerator: requireAccelerator,
		})
		if err != nil {
			return superGPUMeshFail(m, classifySuperGPUError(err))
		}
		raw, err := json.Marshal(result.Outputs)
		if err != nil {
			return superGPUMeshFail(m, "SUPERGPU_BATCH_ENCODE_FAILED")
		}
		return protocol.Result{
			TraceID: m.TraceID, CorrelationID: m.CorrelationID,
			Source: "N07.supergpu.mesh", Target: m.Source, Status: "ok",
			Metadata: map[string]string{
				"device": result.Device.ID,
				"backend": result.Device.Backend,
				"accelerator": strconv.FormatBool(result.Device.Backend != "cpu"),
				"operation": op, "workers": strconv.Itoa(workers), "batch_json": string(raw),
			},
		}, nil
	})
}

func isFederatedNucleus(source string) bool {
	switch strings.TrimSpace(source) {
	case protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06:
		return true
	default:
		return false
	}
}

func classifySuperGPUError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if strings.Contains(msg, "SUPERGPU_ACCELERATOR") {
		return "SUPERGPU_ACCELERATOR_UNAVAILABLE"
	}
	return msg
}

func superGPUMeshFail(m protocol.Message, code string) (protocol.Result, error) {
	if strings.TrimSpace(code) == "" {
		code = "SUPERGPU_MESH_EXECUTION_FAILED"
	}
	err := errors.New(code)
	return protocol.Result{
		TraceID: m.TraceID, CorrelationID: m.CorrelationID,
		Source: "N07.supergpu.mesh", Target: m.Source, Status: "error", Error: code,
	}, err
}
