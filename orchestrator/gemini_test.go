package orchestrator

import (
	"math"
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestGeminiPolicyValidation(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid", raw: `{"id":"g1","cost":0.1,"risk":0.2,"urgency":1,"impact":2}`},
		{name: "missing-id", raw: `{"cost":0.1,"risk":0.2,"urgency":1,"impact":2}`, wantErr: true},
		{name: "risk-over-one", raw: `{"id":"g1","cost":0.1,"risk":1.1,"urgency":1,"impact":2}`, wantErr: true},
		{name: "negative-cost", raw: `{"id":"g1","cost":-0.1,"risk":0.2,"urgency":1,"impact":2}`, wantErr: true},
		{name: "invalid-json", raw: `{"id":`, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseGeminiPolicy(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got=%v", tc.wantErr, err)
			}
		})
	}
}

func TestGeminiPayloadPreservesNativeInputs(t *testing.T) {
	message := protocol.NewMessage("N01", "N07", "command", "gemini.delegate.text@1.0.0", nil)
	message.Metadata["text"] = "hello"
	message.Metadata["system_instruction"] = "system"
	message.Metadata["temperature"] = "0.7"
	message.Metadata["max_output_tokens"] = "512"
	message.Metadata["use_web_search"] = "true"

	payload := geminiPayload(message, "hello")
	if payload["text"] != "hello" || payload["input"] != "hello" {
		t.Fatalf("text propagation failed: %#v", payload)
	}
	if payload["systemInstruction"] != "system" || payload["temperature"] != 0.7 || payload["maxOutputTokens"] != 512 || payload["useWebSearch"] != true {
		t.Fatalf("metadata propagation failed: %#v", payload)
	}
}

func TestGeminiResponseNormalizationRejectsMissingProviderOutput(t *testing.T) {
	if _, _, err := normalizeGeminiResponse(geminiText, map[string]any{"payload": map[string]any{}}); err == nil {
		t.Fatal("expected missing Gemini text to fail")
	}
	if _, _, err := normalizeGeminiResponse(geminiAudioTranscribe, map[string]any{"payload": map[string]any{"transcript": ""}}); err == nil {
		t.Fatal("expected missing transcript to fail")
	}
}

func TestGeminiFloatSliceRejectsNonFiniteValues(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := floatSlice([]float64{0.1, value, 0.2}); err == nil {
			t.Fatalf("expected non-finite value to be rejected: %v", value)
		}
	}
	values, err := floatSlice([]any{0.1, 0.2, 0.3})
	if err != nil || len(values) != 3 {
		t.Fatalf("expected valid numeric vector, values=%v err=%v", values, err)
	}
}
