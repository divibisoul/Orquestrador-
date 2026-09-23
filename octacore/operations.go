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
	OpDescribe              = "octacore.describe@1.0.0"
	OpHealth                = "octacore.health@1.0.0"
	OpSubmit                = "octacore.submit@1.0.0"
	OpBatch                 = "octacore.batch@1.0.0"
	OpSignal                = "octacore.signal@1.0.0"
	OpFederatedContextCycle = "octacore.federated_context_cycle@1.0.0"
)

func RegisterOctaCoreOperations(engine *orchestrator.Engine, processor *Processor) error {
	if engine == nil {
		return errors.New("orchestrator engine is required")
	}
	if processor == nil {
		return errors.New("Octacore processor is required")
	}
	registrations := map[string]orchestrator.Handler{
		OpDescribe: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
			raw, err := json.Marshal(map[string]any{"name": "Octacore", "type": "system_gpu_federated_processor", "silicon_gpu": false, "slots": processor.Inventory(), "backends": []string{"IN_PROCESS", "WEBASSEMBLY", "WEBGPU", "REMOTE_MESH"}})
			return octaProtocolResult(message, raw, err)
		},
		OpHealth: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
			raw, err := json.Marshal(processor.Health())
			return octaProtocolResult(message, raw, err)
		},
		OpSubmit: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			job, err := decodeJob(message.Metadata)
			if err != nil {
				return octaProtocolResult(message, nil, fmt.Errorf("INVALID_OCTACORE_JOB: %w", err))
			}
			result := processor.Submit(ctx, job)
			raw, marshalErr := json.Marshal(result)
			if marshalErr != nil {
				return octaProtocolResult(message, nil, marshalErr)
			}
			if !result.OK {
				return octaProtocolResult(message, raw, errors.New(result.Error.Code+":"+result.Error.Message))
			}
			return octaProtocolResult(message, raw, nil)
		},
		OpBatch: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			jobs, err := decodeJobs(message.Metadata)
			if err != nil {
				return octaProtocolResult(message, nil, fmt.Errorf("INVALID_OCTACORE_BATCH: %w", err))
			}
			results := processor.Batch(ctx, jobs)
			raw, marshalErr := json.Marshal(results)
			if marshalErr != nil {
				return octaProtocolResult(message, nil, marshalErr)
			}
			return octaProtocolResult(message, raw, nil)
		},
		OpFederatedContextCycle: func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			input := strings.TrimSpace(message.Metadata["octacore_input"])
			allowResearchSkip := message.Metadata["allow_research_skip"] == "true"
			var envelopePayload map[string]any
			if rawPayload := strings.TrimSpace(message.Metadata["local_payload_json"]); rawPayload != "" {
				if err := json.Unmarshal([]byte(rawPayload), &envelopePayload); err != nil {
					return octaProtocolResult(message, nil, fmt.Errorf("invalid local Octacore payload: %w", err))
				}
				if input == "" {
					if value, ok := envelopePayload["input"].(string); ok {
						input = strings.TrimSpace(value)
					}
				}
				if value, ok := envelopePayload["allow_research_skip"].(bool); ok {
					allowResearchSkip = value
				}
			}
			if input == "" {
				return octaProtocolResult(message, nil, errors.New("octacore_input is required"))
			}
			request := FederatedContextInput{
				CorrelationID:     message.CorrelationID,
				Input:             input,
				AllowResearchSkip: allowResearchSkip,
				Priority:          90,
				TTLMS:             30_000,
			}
			if raw := strings.TrimSpace(message.Metadata["research_payload_json"]); raw != "" {
				if err := json.Unmarshal([]byte(raw), &request.ResearchPayload); err != nil {
					return octaProtocolResult(message, nil, fmt.Errorf("invalid research payload: %w", err))
				}
			}
			if raw := strings.TrimSpace(message.Metadata["perception_payload_json"]); raw != "" {
				if err := json.Unmarshal([]byte(raw), &request.PerceptionPayload); err != nil {
					return octaProtocolResult(message, nil, fmt.Errorf("invalid perception payload: %w", err))
				}
			}
			if envelopePayload != nil {
				if request.ResearchPayload == nil {
					if value, ok := envelopePayload["research_payload"].(map[string]any); ok {
						request.ResearchPayload = value
					}
				}
				if request.PerceptionPayload == nil {
					if value, ok := envelopePayload["perception_payload"].(map[string]any); ok {
						request.PerceptionPayload = value
					}
				}
			}
			raw, err := json.Marshal(processor.ExecuteFederatedContextCycle(ctx, request))
			return octaProtocolResult(message, raw, err)
		},
		OpSignal: func(_ context.Context, message protocol.Message) (protocol.Result, error) {
			signal := strings.ToLower(strings.TrimSpace(message.Metadata["signal"]))
			switch signal {
			case "throttle":
				levelText := strings.TrimSpace(message.Metadata["level"])
				var level int
				if _, err := fmt.Sscan(levelText, &level); err != nil {
					return octaProtocolResult(message, nil, errors.New("signal level must be integer 0..3"))
				}
				if err := processor.SetThrottle(level); err != nil {
					return octaProtocolResult(message, nil, err)
				}
			case "halt":
				processor.Halt()
			case "resume":
				processor.Resume()
			default:
				return octaProtocolResult(message, nil, errors.New("unsupported Octacore signal"))
			}
			raw, err := json.Marshal(processor.Health())
			return octaProtocolResult(message, raw, err)
		},
	}
	for name, handler := range registrations {
		if err := engine.Register(name, handler); err != nil {
			return err
		}
	}
	return nil
}

func decodeJob(metadata map[string]string) (OctaCoreJob, error) {
	raw := strings.TrimSpace(metadata["octacore_job_json"])
	if raw == "" {
		return OctaCoreJob{}, errors.New("metadata.octacore_job_json is required")
	}
	var job OctaCoreJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return OctaCoreJob{}, err
	}
	if err := job.Validate(); err != nil {
		return OctaCoreJob{}, err
	}
	return job, nil
}

func decodeJobs(metadata map[string]string) ([]OctaCoreJob, error) {
	raw := strings.TrimSpace(metadata["octacore_jobs_json"])
	if raw == "" {
		return nil, errors.New("metadata.octacore_jobs_json is required")
	}
	var jobs []OctaCoreJob
	if err := json.Unmarshal([]byte(raw), &jobs); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, errors.New("Octacore batch is empty")
	}
	for i, job := range jobs {
		if err := job.Validate(); err != nil {
			return nil, fmt.Errorf("job %d: %w", i, err)
		}
	}
	return jobs, nil
}

func octaProtocolResult(message protocol.Message, raw []byte, err error) (protocol.Result, error) {
	metadata := map[string]string{"octacore_json": ""}
	if len(raw) > 0 {
		metadata["octacore_json"] = string(raw)
	}
	result := protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07", Target: message.Source, Status: "ok", Metadata: metadata}
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result, err
	}
	return result, nil
}
