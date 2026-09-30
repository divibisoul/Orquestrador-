package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/cognitive"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	cognitiveGoalRun  = "cognitive.goal.run@1.0.0"
	cognitiveGoalPlan = "cognitive.goal.plan@1.0.0"
	cognitiveHealth   = "cognitive.health@1.0.0"
)

type cognitiveMeshAdapter struct {
	peers *mesh.PeerClient
}

func (a cognitiveMeshAdapter) ConfiguredPeers() []cognitive.PeerDescriptor {
	if a.peers == nil {
		return nil
	}
	peers := a.peers.ConfiguredPeers()
	out := make([]cognitive.PeerDescriptor, 0, len(peers))
	for _, peer := range peers {
		out = append(out, cognitive.PeerDescriptor{Nucleus: peer.Nucleus})
	}
	return out
}

func (a cognitiveMeshAdapter) Discover(ctx context.Context, nucleus string) (map[string]any, error) {
	if a.peers == nil {
		return nil, errors.New("mesh peer client unavailable")
	}
	return a.peers.Discover(ctx, nucleus)
}

func (a cognitiveMeshAdapter) CallBestDynamic(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
	if a.peers == nil {
		return nil, "", errors.New("mesh peer client unavailable")
	}
	return a.peers.CallBestDynamic(ctx, capability, payload, correlation)
}

func RegisterCognitiveOperations(
	e *Engine,
	peers *mesh.PeerClient,
	sara *backend.SARAProxy,
	store *backend.SupabaseStore,
) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if peers == nil {
		return errors.New("cognitive operations require Mesh peer client")
	}

	cfg := cognitive.DefaultConfig()
	local := func(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
		delegate, ok := geminiDelegateFor(capability)
		if !ok {
			return nil, "", errors.New("local cognitive executor only supports Gemini gateway capabilities")
		}
		metadata, err := mapPayloadToMetadata(payload, correlation)
		if err != nil {
			return nil, "", err
		}
		result, err := e.Execute(ctx, delegate, nil, metadata)
		out := map[string]any{
			"trace_id":       result.TraceID,
			"correlation_id": result.CorrelationID,
			"status":         result.Status,
			"metadata":       result.Metadata,
		}
		if result.Error != "" {
			out["error"] = result.Error
		}
		return out, "N07.gemini-gateway", err
	}

	loop, err := cognitive.Build(cfg, e.cortex, cognitiveMeshAdapter{peers: peers}, sara, store, local)
	if err != nil {
		return err
	}

	if err := e.Register(cognitiveGoalRun, func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		goal, err := decodeCognitiveGoal(message)
		if err != nil {
			return cognitiveResult(message, nil, err)
		}
		observations, err := loop.RunPlanned(ctx, goal)
		raw, marshalErr := json.Marshal(map[string]any{
			"goal": goal, "observations": observations,
		})
		if marshalErr != nil {
			return cognitiveResult(message, nil, marshalErr)
		}
		return cognitiveResult(message, raw, err)
	}); err != nil {
		return err
	}

	if err := e.Register(cognitiveGoalPlan, func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		goal, err := decodeCognitiveGoal(message)
		if err != nil {
			return cognitiveResult(message, nil, err)
		}
		planner, err := cognitive.NewPlanner(cognitiveMeshAdapter{peers: peers})
		if err != nil {
			return cognitiveResult(message, nil, err)
		}
		steps, err := planner.Plan(ctx, goal)
		raw, marshalErr := json.Marshal(map[string]any{"goal": goal, "steps": steps})
		if marshalErr != nil {
			return cognitiveResult(message, nil, marshalErr)
		}
		return cognitiveResult(message, raw, err)
	}); err != nil {
		return err
	}

	return e.Register(cognitiveHealth, func(_ context.Context, message protocol.Message) (protocol.Result, error) {
		raw, err := json.Marshal(map[string]any{
			"status":  "READY",
			"enabled": loop.Enabled(),
		})
		return cognitiveResult(message, raw, err)
	})
}

func decodeCognitiveGoal(message protocol.Message) (cognitive.Goal, error) {
	raw := strings.TrimSpace(message.Metadata["cognitive_goal_json"])
	if raw == "" {
		return cognitive.Goal{}, errors.New("metadata.cognitive_goal_json is required")
	}
	var goal cognitive.Goal
	if err := json.Unmarshal([]byte(raw), &goal); err != nil {
		return cognitive.Goal{}, fmt.Errorf("INVALID_COGNITIVE_GOAL: %w", err)
	}
	if strings.TrimSpace(goal.CorrelationID) == "" {
		goal.CorrelationID = message.CorrelationID
	}
	if strings.TrimSpace(goal.CorrelationID) == "" {
		return cognitive.Goal{}, errors.New("cognitive goal correlation_id is required")
	}
	if strings.TrimSpace(goal.ID) == "" {
		return cognitive.Goal{}, errors.New("cognitive goal id is required")
	}
	if strings.TrimSpace(goal.Objective) == "" {
		return cognitive.Goal{}, errors.New("cognitive goal objective is required")
	}
	return goal, nil
}

func cognitiveResult(message protocol.Message, raw []byte, err error) (protocol.Result, error) {
	metadata := map[string]string{}
	if len(raw) > 0 {
		metadata["cognitive_json"] = string(raw)
	}
	result := protocol.Result{
		TraceID:       message.TraceID,
		CorrelationID: message.CorrelationID,
		Source:        "N07.cognitive",
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

func geminiDelegateFor(capability string) (string, bool) {
	switch strings.TrimSpace(capability) {
	case geminiText:
		return "gemini.delegate.text@1.0.0", true
	case geminiMultimodal:
		return "gemini.delegate.multimodal@1.0.0", true
	case geminiAudioTranscribe:
		return "gemini.delegate.audio.transcribe@1.0.0", true
	case geminiAudioAnalyze:
		return "gemini.delegate.audio.analyze@1.0.0", true
	case geminiSpeechSynthesize:
		return "gemini.delegate.speech.synthesize@1.0.0", true
	default:
		return "", false
	}
}

func mapPayloadToMetadata(payload map[string]any, correlation string) (map[string]string, error) {
	metadata := map[string]string{"correlation_id": correlation}
	for key, value := range payload {
		raw, err := valueToMetadata(value)
		if err != nil {
			return nil, fmt.Errorf("cognitive payload %s: %w", key, err)
		}
		metadata[normalizeCognitiveMetadataKey(key)] = raw
	}
	return metadata, nil
}

func normalizeCognitiveMetadataKey(key string) string {
	switch strings.TrimSpace(key) {
	case "systemInstruction":
		return "system_instruction"
	case "audioBase64":
		return "audio_base64"
	case "imageBase64":
		return "image_base64"
	case "imageMimeType":
		return "image_mime_type"
	case "mediaBase64":
		return "media_base64"
	case "mediaMimeType":
		return "media_mime_type"
	case "audioMimeType":
		return "audio_mime_type"
	case "speechText":
		return "speech_text"
	case "useWebSearch":
		return "use_web_search"
	case "maxOutputTokens":
		return "max_output_tokens"
	case "candidateJson":
		return "candidate_json"
	default:
		return key
	}
}

func valueToMetadata(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return fmt.Sprintf("%t", typed), nil
	case float64:
		return fmt.Sprintf("%v", typed), nil
	case float32:
		return fmt.Sprintf("%v", typed), nil
	case int:
		return fmt.Sprintf("%d", typed), nil
	case int64:
		return fmt.Sprintf("%d", typed), nil
	case nil:
		return "", nil
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
}
