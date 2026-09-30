package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
)

// GeminiPeerCaller is deliberately narrower than mesh.PeerClient so N07 keeps
// orchestration ownership while N02 remains the canonical Gemini provider.
type GeminiPeerCaller interface {
	CallWithCorrelation(context.Context, string, string, map[string]any, string) (map[string]any, error)
}

// GeminiExecutionEvent represents a real Gemini execution lifecycle event.
// A reporter may forward it to Clareira/Vagus without creating a second bus.
type GeminiExecutionEvent struct {
	Phase         string
	Capability    string
	Provider      string
	Model         string
	Source        string
	Owner         string
	CorrelationID string
	InputSize     int
	OutputSize    int
	Error         string
	At            time.Time
}

type GeminiExecutionReporter func(context.Context, GeminiExecutionEvent) error

type geminiPolicy struct {
	ID      string  `json:"id"`
	Cost    float64 `json:"cost"`
	Risk    float64 `json:"risk"`
	Urgency float64 `json:"urgency"`
	Impact  float64 `json:"impact"`
}

const (
	geminiOwner           = "N02"
	geminiProvider        = "google-gemini"
	geminiText            = "gemini.text.generate"
	geminiMultimodal      = "gemini.multimodal.generate"
	geminiAudioTranscribe = "gemini.audio.transcribe"
	geminiAudioAnalyze    = "gemini.audio.analyze"
	geminiSpeechSynthesize = "gemini.speech.synthesize"
)

func RegisterGeminiOperations(e *Engine, peer GeminiPeerCaller, reporter GeminiExecutionReporter) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if peer == nil {
		return errors.New("Gemini gateway requires N07 Mesh peer caller")
	}

	operations := []struct {
		name       string
		capability string
	}{
		{"gemini.delegate.text@1.0.0", geminiText},
		{"gemini.delegate.multimodal@1.0.0", geminiMultimodal},
		{"gemini.delegate.audio.transcribe@1.0.0", geminiAudioTranscribe},
		{"gemini.delegate.audio.analyze@1.0.0", geminiAudioAnalyze},
		{"gemini.delegate.speech.synthesize@1.0.0", geminiSpeechSynthesize},
	}

	for _, item := range operations {
		capability := item.capability
		if err := e.Register(item.name, func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			return executeGemini(e, peer, reporter, capability, message)
		}); err != nil {
			return err
		}
	}
	return nil
}

func executeGemini(
	e *Engine,
	peer GeminiPeerCaller,
	reporter GeminiExecutionReporter,
	capability string,
	message protocol.Message,
) (protocol.Result, error) {
	if ctxErr := contextError(message); ctxErr != nil {
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.gemini", Target: message.Source, Status: "rejected", Error: ctxErr.Error()}, ctxErr
	}

	policy, err := parseGeminiPolicy(message.Metadata["candidate_json"])
	if err != nil {
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.prefrontal", Target: message.Source, Status: "rejected", Error: err.Error()}, err
	}

	semanticText := firstNonEmpty(
		message.Metadata["text"],
		message.Metadata["input"],
		message.Metadata["speech_text"],
		message.Metadata["transcript"],
	)

	ctx := context.Background()
	if !message.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, message.Deadline)
		defer cancel()
	}

	neuralInput, err := resolveGeminiNeuralInput(ctx, peer, message, semanticText)
	if err != nil {
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.prefrontal", Target: message.Source, Status: "rejected", Error: err.Error()}, err
	}

	neocortex, err := prefrontal.NewNeocortex(e.cortex, e.neural)
	if err != nil {
		return protocol.Result{}, err
	}
	candidate, err := neocortex.Evaluate(ctx, policy.ID, neuralInput, policy.Risk, policy.Cost, policy.Urgency, policy.Impact)
	if err != nil {
		reportGeminiFailure(ctx, reporter, capability, message, err, len(neuralInput))
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.prefrontal", Target: message.Source, Status: "rejected", Error: err.Error()}, err
	}
	decision, err := neocortex.Commit(candidate, "gemini-provider-admission")
	if err != nil {
		reportGeminiFailure(ctx, reporter, capability, message, err, len(neuralInput))
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.prefrontal", Target: message.Source, Status: "rejected", Error: err.Error()}, err
	}

	startReportErr := reportGemini(ctx, reporter, GeminiExecutionEvent{
		Phase: "started", Capability: capability, Provider: geminiProvider, Model: geminiModelFor(capability),
		Source: "N07.Orchestrator", Owner: geminiOwner, CorrelationID: message.CorrelationID,
		InputSize: len(neuralInput), At: time.Now().UTC(),
	})

	payload := geminiPayload(message, semanticText)
	upstream, callErr := peer.CallWithCorrelation(ctx, geminiOwner, capability, payload, message.CorrelationID)
	if callErr != nil {
		reportGeminiFailure(ctx, reporter, capability, message, callErr, len(neuralInput))
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.gemini", Target: message.Source, Status: "error", Error: callErr.Error()}, callErr
	}

	metadata, outputSize, normalizeErr := normalizeGeminiResponse(capability, upstream)
	if normalizeErr != nil {
		reportGeminiFailure(ctx, reporter, capability, message, normalizeErr, len(neuralInput))
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.gemini", Target: message.Source, Status: "error", Error: normalizeErr.Error()}, normalizeErr
	}

	metadata["provider"] = geminiProvider
	metadata["owner"] = geminiOwner
	metadata["route"] = "N07.prefrontal>N02"
	metadata["correlation_id"] = message.CorrelationID
	metadata["prefrontal_decision_id"] = decision.ID
	metadata["prefrontal_score"] = floatString(decision.Score)
	metadata["prefrontal_candidate_id"] = candidate.ID
	metadata["prefrontal_neural_dimensions"] = itoa(len(neuralInput))
	if startReportErr != nil {
		metadata["clareira_start_report_error"] = startReportErr.Error()
	}

	reportErr := reportGemini(ctx, reporter, GeminiExecutionEvent{
		Phase: "completed", Capability: capability, Provider: geminiProvider, Model: geminiModelFor(capability),
		Source: "N07.Orchestrator", Owner: geminiOwner, CorrelationID: message.CorrelationID,
		InputSize: len(neuralInput), OutputSize: outputSize, At: time.Now().UTC(),
	})
	if reportErr != nil {
		metadata["clareira_report_error"] = reportErr.Error()
	}

	return protocol.Result{
		TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.gemini", Target: message.Source,
		Status: "ok", Metadata: metadata,
	}, nil
}

func contextError(message protocol.Message) error {
	if strings.TrimSpace(message.CorrelationID) == "" {
		return errors.New("correlation id is required")
	}
	if !message.Deadline.IsZero() && time.Now().After(message.Deadline) {
		return errors.New("message deadline exceeded")
	}
	return nil
}

func parseGeminiPolicy(raw string) (geminiPolicy, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return geminiPolicy{}, errors.New("candidate_json is required for prefrontal Gemini admission")
	}
	var policy geminiPolicy
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return geminiPolicy{}, errors.New("candidate_json is invalid")
	}
	if strings.TrimSpace(policy.ID) == "" {
		return geminiPolicy{}, errors.New("candidate_json.id is required")
	}
	for name, value := range map[string]float64{
		"cost": policy.Cost, "risk": policy.Risk, "urgency": policy.Urgency, "impact": policy.Impact,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return geminiPolicy{}, fmt.Errorf("candidate_json.%s must be finite and non-negative", name)
		}
	}
	if policy.Risk > 1 {
		return geminiPolicy{}, errors.New("candidate_json.risk must be <= 1")
	}
	return policy, nil
}

func resolveGeminiNeuralInput(ctx context.Context, peer GeminiPeerCaller, message protocol.Message, semanticText string) ([]float64, error) {
	if len(message.Payload) > 0 {
		return append([]float64(nil), message.Payload...), nil
	}
	if raw := strings.TrimSpace(message.Metadata["neural_input_json"]); raw != "" {
		var values []float64
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return nil, errors.New("neural_input_json is invalid")
		}
		if len(values) == 0 {
			return nil, errors.New("neural_input_json cannot be empty")
		}
		return values, nil
	}
	if semanticText == "" {
		return nil, errors.New("prefrontal neural input is unavailable; provide payload, neural_input_json or text")
	}

	upstream, err := peer.CallWithCorrelation(ctx, geminiOwner, "neural.bnc_v2", map[string]any{"text": semanticText}, message.CorrelationID)
	if err != nil {
		return nil, fmt.Errorf("N02 BNCv2 unavailable for prefrontal admission: %w", err)
	}
	payload := responsePayload(upstream)
	bnc, ok := payload["bnc"].(map[string]any)
	if !ok {
		return nil, errors.New("N02 BNCv2 response did not expose bnc")
	}
	values, err := floatSlice(bnc["vector"])
	if err != nil {
		return nil, fmt.Errorf("N02 BNCv2 vector invalid: %w", err)
	}
	return values, nil
}

func geminiPayload(message protocol.Message, semanticText string) map[string]any {
	payload := map[string]any{}
	if semanticText != "" {
		payload["text"] = semanticText
		payload["input"] = semanticText
	}
	copyMetadata := func(key string, target string) {
		if value := strings.TrimSpace(message.Metadata[key]); value != "" {
			payload[target] = value
		}
	}
	for _, item := range [][2]string{
		{"system_instruction", "systemInstruction"},
		{"audio_base64", "audioBase64"},
		{"mime_type", "mimeType"},
		{"image_base64", "imageBase64"},
		{"image_mime_type", "imageMimeType"},
		{"media_base64", "mediaBase64"},
		{"media_mime_type", "mediaMimeType"},
		{"audio_mime_type", "audioMimeType"},
		{"speech_text", "speechText"},
		{"voice", "voice"},
		{"mode", "mode"},
	}{
		copyMetadata(item[0], item[1])
	}
	if value := strings.TrimSpace(message.Metadata["use_web_search"]); value != "" {
		payload["useWebSearch"] = parseBool(value)
	}
	if value := strings.TrimSpace(message.Metadata["temperature"]); value != "" {
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			payload["temperature"] = number
		}
	}
	if value := strings.TrimSpace(message.Metadata["max_output_tokens"]); value != "" {
		if number, err := strconv.Atoi(value); err == nil {
			payload["maxOutputTokens"] = number
		}
	}
	if value := strings.TrimSpace(message.Metadata["contents_json"]); value != "" {
		var contents []any
		if err := json.Unmarshal([]byte(value), &contents); err != nil {
			// Invalid structured contents must fail at the N02 provider boundary,
			// not be replaced with a fabricated request.
			payload["contents_json"] = value
		} else {
			payload["contents"] = contents
		}
	}
	return payload
}

func normalizeGeminiResponse(capability string, upstream map[string]any) (map[string]string, int, error) {
	payload := responsePayload(upstream)
	metadata := map[string]string{}
	switch capability {
	case geminiText, geminiMultimodal:
		textValue, ok := payload["text"].(string)
		if !ok || strings.TrimSpace(textValue) == "" {
			return nil, 0, errors.New("Gemini response did not expose text")
		}
		metadata["output_text"] = textValue
		return metadata, len([]rune(textValue)), nil
	case geminiAudioTranscribe:
		value, ok := payload["transcript"].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, 0, errors.New("Gemini response did not expose transcript")
		}
		metadata["transcript"] = value
		return metadata, len([]rune(value)), nil
	case geminiAudioAnalyze:
		value, ok := payload["analysis"].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, 0, errors.New("Gemini response did not expose analysis")
		}
		metadata["analysis"] = value
		return metadata, len([]rune(value)), nil
	case geminiSpeechSynthesize:
		value, ok := payload["audio"].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, 0, errors.New("Gemini TTS response did not expose audio")
		}
		metadata["audio_base64"] = value
		return metadata, len(value), nil
	default:
		return nil, 0, fmt.Errorf("unsupported Gemini capability: %s", capability)
	}
}

func responsePayload(response map[string]any) map[string]any {
	if response == nil {
		return map[string]any{}
	}
	if payload, ok := response["payload"].(map[string]any); ok {
		return payload
	}
	return response
}

func floatSlice(value any) ([]float64, error) {
	switch values := value.(type) {
	case []any:
		out := make([]float64, len(values))
		for i, item := range values {
			number, ok := item.(float64)
			if !ok || math.IsNaN(number) || math.IsInf(number, 0) {
				return nil, errors.New("vector contains non-finite or non-numeric values")
			}
			out[i] = number
		}
		return out, nil
	case []float64:
		return append([]float64(nil), values...), nil
	default:
		return nil, errors.New("vector is not numeric")
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func geminiModelFor(capability string) string {
	switch capability {
	case geminiText, geminiMultimodal, geminiAudioAnalyze:
		return "gemini-3.8-flash"
	case geminiAudioTranscribe:
		return "gemini-3.5-transcribe"
	case geminiSpeechSynthesize:
		return "gemini-3.8-flash-tts"
	default:
		return "configured-by-N02"
	}
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func reportGemini(ctx context.Context, reporter GeminiExecutionReporter, event GeminiExecutionEvent) error {
	if reporter == nil {
		return nil
	}
	return reporter(ctx, event)
}

func reportGeminiFailure(ctx context.Context, reporter GeminiExecutionReporter, capability string, message protocol.Message, err error, inputSize int) {
	_ = reportGemini(ctx, reporter, GeminiExecutionEvent{
		Phase: "failed", Capability: capability, Provider: geminiProvider, Model: geminiModelFor(capability),
		Source: "N07.Orchestrator", Owner: geminiOwner, CorrelationID: message.CorrelationID,
		InputSize: inputSize, Error: err.Error(), At: time.Now().UTC(),
	})
}
