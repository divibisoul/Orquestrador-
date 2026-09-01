package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// RegisterSuperGPUOperations exposes the existing SuperGPU runtime through the
// same operation registry used by neural, compute and cognitive capabilities.
// It does not duplicate the runtime or invent a second execution path.
func RegisterSuperGPUOperations(e *Engine) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	registrations := []struct {
		name    string
		handler Handler
	}{
		{
			name: "supergpu.describe@1.0.0",
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				select {
				case <-ctx.Done():
					return protocol.Result{}, ctx.Err()
				default:
				}
				devices := e.compute.Discover()
				encoded, err := json.Marshal(devices)
				if err != nil {
					return protocol.Result{}, err
				}
				return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.supergpu", Target: message.Source, Status: "ok", Metadata: map[string]string{"devices_json": string(encoded), "device_count": strconv.Itoa(len(devices))}}, nil
			},
		},
		{
			name: "supergpu.execute@1.0.0",
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				device, err := e.compute.Select(strings.TrimSpace(message.Metadata["device"]))
				if err != nil {
					return protocol.Result{}, err
				}
				if err := e.compute.Reserve(device.ID, message.TraceID); err != nil {
					return protocol.Result{}, err
				}
				defer e.compute.Release(device.ID, message.TraceID)
				op := strings.TrimSpace(message.Metadata["operation"])
				if op == "" {
					op = "identity"
				}
				values, err := e.compute.Execute(ctx, device, op, message.Payload)
				return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.supergpu", Target: message.Source, Status: status(err), Payload: values, Metadata: map[string]string{"device": device.ID, "operation": op}, Error: errorText(err)}, err
			},
		},
		{
			name: "supergpu.parallel@1.0.0",
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				device, err := e.compute.Select(strings.TrimSpace(message.Metadata["device"]))
				if err != nil {
					return protocol.Result{}, err
				}
				if err := e.compute.Reserve(device.ID, message.TraceID); err != nil {
					return protocol.Result{}, err
				}
				defer e.compute.Release(device.ID, message.TraceID)
				var inputs [][]float64
				if err := json.Unmarshal([]byte(message.Metadata["inputs_json"]), &inputs); err != nil || len(inputs) == 0 {
					return protocol.Result{}, errors.New("metadata.inputs_json must contain a non-empty numeric batch")
				}
				op := strings.TrimSpace(message.Metadata["operation"])
				if op == "" {
					op = "identity"
				}
				workers := 1
				if raw := strings.TrimSpace(message.Metadata["workers"]); raw != "" {
					if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
						workers = parsed
					}
				}
				values, err := e.compute.BatchParallel(ctx, device, op, inputs, workers)
				encoded, encodeErr := json.Marshal(values)
				if err != nil {
					return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.supergpu", Target: message.Source, Status: status(err), Metadata: map[string]string{"device": device.ID, "operation": op}}, err
				}
				if encodeErr != nil {
					return protocol.Result{}, encodeErr
				}
				return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.supergpu", Target: message.Source, Status: "ok", Metadata: map[string]string{"device": device.ID, "operation": op, "workers": strconv.Itoa(workers), "batch_json": string(encoded)}}, nil
			},
		},
		{
			name: "supergpu.memory@1.0.0",
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				select {
				case <-ctx.Done():
					return protocol.Result{}, ctx.Err()
				default:
				}
				device, err := e.compute.Select(strings.TrimSpace(message.Metadata["device"]))
				if err != nil {
					return protocol.Result{}, err
				}
				stats := e.compute.MemoryStats(device)
				encoded, err := json.Marshal(stats)
				if err != nil {
					return protocol.Result{}, err
				}
				return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.supergpu", Target: message.Source, Status: "ok", Metadata: map[string]string{"memory_json": string(encoded), "device": device.ID}}, nil
			},
		},
	}
	for _, item := range registrations {
		if err := e.Register(item.name, item.handler); err != nil {
			if strings.Contains(err.Error(), "already registered") {
				continue
			}
			return err
		}
	}
	return nil
}
