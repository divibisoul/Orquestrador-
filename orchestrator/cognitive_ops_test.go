package orchestrator

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestCognitiveGoalDecodeDefaultsCorrelationFromMessage(t *testing.T) {
	message := protocol.NewMessage("N01", "N07", "command", "cognitive.goal.plan@1.0.0", nil)
	message.Metadata["cognitive_goal_json"] = `{"goal_id":"g1","objective":"recover","capabilities":["ai.infer"]}`

	goal, err := decodeCognitiveGoal(message)
	if err != nil {
		t.Fatal(err)
	}
	if goal.CorrelationID != message.CorrelationID {
		t.Fatalf("expected message correlation propagation, got %q", goal.CorrelationID)
	}
}

func TestCognitiveGoalDecodeRejectsIncompleteGoal(t *testing.T) {
	message := protocol.NewMessage("N01", "N07", "command", "cognitive.goal.plan@1.0.0", nil)
	message.Metadata["cognitive_goal_json"] = `{"goal_id":"g1"}`

	if _, err := decodeCognitiveGoal(message); err == nil {
		t.Fatal("expected incomplete cognitive goal to fail")
	}
}

func TestGeminiDelegateMappingPreservesCanonicalN02Capabilities(t *testing.T) {
	cases := map[string]string{
		"gemini.text.generate":       "gemini.delegate.text@1.0.0",
		"gemini.multimodal.generate": "gemini.delegate.multimodal@1.0.0",
		"gemini.audio.transcribe":    "gemini.delegate.audio.transcribe@1.0.0",
		"gemini.audio.analyze":       "gemini.delegate.audio.analyze@1.0.0",
		"gemini.speech.synthesize":   "gemini.delegate.speech.synthesize@1.0.0",
	}
	for canonical, delegate := range cases {
		got, ok := geminiDelegateFor(canonical)
		if !ok || got != delegate {
			t.Fatalf("mapping %s => %q, %v; want %s", canonical, got, ok, delegate)
		}
	}
}

func TestMapPayloadToMetadataPreservesStructuredValues(t *testing.T) {
	values, err := mapPayloadToMetadata(map[string]any{
		"temperature":   0.7,
		"useWebSearch":  true,
		"candidateJson": map[string]any{"risk": 0.1},
	}, "corr-1")
	if err != nil {
		t.Fatal(err)
	}
	if values["temperature"] != "0.7" || values["use_web_search"] != "true" || values["candidate_json"] != `{"risk":0.1}` {
		t.Fatalf("unexpected metadata mapping: %#v", values)
	}
}
