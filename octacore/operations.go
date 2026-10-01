package octacore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	OpDescribe = "octacore.describe@1.0.0"
	OpHealth   = "octacore.health@1.0.0"
	OpSubmit   = "octacore.submit@1.0.0"
	OpBatch    = "octacore.batch@1.0.0"
	OpSignal   = "octacore.signal@1.0.0"
	OpCoreDescribe = "octacore.core.describe@1.0.0"
	OpCoreHealth   = "octacore.core.health@1.0.0"
	OpCoreExecute  = "octacore.core.execute@1.0.0"
)

func RegisterOperations(engine *orchestrator.Engine, processor *Processor) error {
	if engine == nil {
		return errors.New("orchestrator engine is required")
	}
	if processor == nil {
		return errors.New("Octacore processor is required")
	}

	registrations := []struct {
		name    string
		handler orchestrator.Handler
	}{
		{
			name: OpDescribe,
			handler: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
				payload := map[string]any{
					"name":        "Octacore",
					"type":        "system_gpu_federated_processor",
					"silicon_gpu": false,
					"slots":       processor.Inventory(),
					"backends":    []string{"IN_PROCESS", "WEBASSEMBLY", "WEBGPU", "REMOTE_MESH"},
					"authority":   "N07 scheduling; SARA regeneration",
				}
				raw, err := json.Marshal(payload)
				return protocolResult(message, raw, err)
			},
		},
		{
			name: OpHealth,
			handler: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
				raw, err := json.Marshal(processor.Health())
				return protocolResult(message, raw, err)
			},
		},
		{
			name: OpSubmit,
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				job, err := decodeJob(message.Metadata)
				if err != nil {
					return protocolResult(message, nil, fmt.Errorf("INVALID_OCTACORE_JOB: %w", err))
				}
				if message.CorrelationID != "" && job.CorrelationID != message.CorrelationID {
					return protocolResult(message, nil, errors.New("CORRELATION_ID_MISMATCH"))
				}
				result := processor.Submit(ctx, job)
				raw, marshalErr := json.Marshal(result)
				if marshalErr != nil {
					return protocolResult(message, nil, marshalErr)
				}
				if !result.OK && result.Error != nil {
					return protocolResult(message, raw, fmt.Errorf("%s:%s", result.Error.Code, result.Error.Message))
				}
				return protocolResult(message, raw, nil)
			},
		},
		{
			name: OpBatch,
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				jobs, err := decodeJobs(message.Metadata)
				if err != nil {
					return protocolResult(message, nil, fmt.Errorf("INVALID_OCTACORE_BATCH: %w", err))
				}
				results := processor.Batch(ctx, jobs)
				raw, marshalErr := json.Marshal(results)
				if marshalErr != nil {
					return protocolResult(message, nil, marshalErr)
				}
				return protocolResult(message, raw, nil)
			},
		},

		{
			name: OpCoreDescribe,
			handler: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
				core := processor.ExecutiveCore()
				if core == nil {
					return protocolResult(message, nil, errors.New("executive core is not connected"))
				}
				raw, err := json.Marshal(core.Describe())
				return protocolResult(message, raw, err)
			},
		},
		{
			name: OpCoreHealth,
			handler: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
				core := processor.ExecutiveCore()
				if core == nil {
					return protocolResult(message, nil, errors.New("executive core is not connected"))
				}
				raw, err := json.Marshal(core.Health())
				return protocolResult(message, raw, err)
			},
		},
		{
			name: OpCoreExecute,
			handler: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
				req, err := DecodeExecutiveCoreRequest(message.Metadata)
				if err != nil {
					return protocolResult(message, nil, fmt.Errorf("INVALID_EXECUTIVE_CORE_REQUEST: %w", err))
				}
				if message.CorrelationID != "" && req.CorrelationID != message.CorrelationID {
					return protocolResult(message, nil, errors.New("CORRELATION_ID_MISMATCH"))
				}
				result, err := processor.ExecutiveExecute(ctx, req)
				raw, marshalErr := json.Marshal(result)
				if marshalErr != nil {
					return protocolResult(message, nil, marshalErr)
				}
				if err != nil {
					return protocolResult(message, raw, err)
				}
				return protocolResult(message, raw, nil)
			},
		},
		{
			name: OpSignal,
			handler: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
				signal := strings.ToLower(strings.TrimSpace(message.Metadata["signal"]))
				switch signal {
				case "throttle":
					var level int
					if _, err := fmt.Sscan(strings.TrimSpace(message.Metadata["level"]), &level); err != nil {
						return protocolResult(message, nil, errors.New("signal level must be integer"))
					}
					if err := processor.Throttle(level); err != nil {
						return protocolResult(message, nil, err)
					}
				case "halt":
					processor.Halt()
				case "resume":
					processor.Resume()
				default:
					return protocolResult(message, nil, errors.New("unsupported Octacore signal"))
				}
				raw, err := json.Marshal(processor.Health())
				return protocolResult(message, raw, err)
			},
		},
	}

	for _, registration := range registrations {
		if err := engine.Register(registration.name, registration.handler); err != nil {
			return err
		}
	}
	return nil
}

func decodeJob(metadata map[string]string) (Job, error) {
	raw := strings.TrimSpace(metadata["octacore_job_json"])
	if raw == "" {
		return Job{}, errors.New("metadata.octacore_job_json is required")
	}
	var job Job
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return Job{}, err
	}
	if err := job.Validate(); err != nil {
		return Job{}, err
	}
	return job, nil
}

func decodeJobs(metadata map[string]string) ([]Job, error) {
	raw := strings.TrimSpace(metadata["octacore_jobs_json"])
	if raw == "" {
		return nil, errors.New("metadata.octacore_jobs_json is required")
	}
	var jobs []Job
	if err := json.Unmarshal([]byte(raw), &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, errors.New("Octacore batch is empty")
	}
	for index, job := range jobs {
		if err := job.Validate(); err != nil {
			return nil, fmt.Errorf("job %d: %w", index, err)
		}
	}
	return jobs, nil
}

func protocolResult(message protocol.Message, raw []byte, err error) (protocol.Result, error) {
	metadata := map[string]string{"octacore_json": ""}
	if len(raw) > 0 {
		metadata["octacore_json"] = string(raw)
	}
	result := protocol.Result{
		TraceID:       message.TraceID,
		CorrelationID: message.CorrelationID,
		Source:        "N07.octacore",
		Target:        message.Source,
		Status:        "ok",
		Metadata:      metadata,
	}
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
	}
	return result, err
}
