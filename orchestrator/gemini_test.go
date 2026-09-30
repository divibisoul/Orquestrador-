package orchestrator

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type geminiPeerCall struct {
	Nucleus     string
	Capability  string
	Correlation string
	Payload     map[string]any
}

type geminiFakePeer struct {
	calls []geminiPeerCall
}

func (p *geminiFakePeer) CallWithCorrelation(_ context.Context, nucleus, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	p.calls = append(p.calls, geminiPeerCall{Nucleus: nucleus, Capability: capability, Correlation: correlation, Payload: payload})
	switch capability {
	case "neural.bnc_v2":
		return map[string]any{"payload": map[string]any{"bnc": map[string]any{"vector": []any{0.1, 0.2, 0.3, 0.4, 0.2, 0.1, 0.2, 0.3}}}}, nil
	case "gemini.text.generate":
		return map[string]any{"payload": map[string]any{"text": "Gemini real output"}}, nil
	default:
		return map[string]any{"payload": map[string]any{}}, nil
	}
}

func newGeminiEngine(t *testing.T) (*Engine, *prefrontal.Cortex) {
	t.Helper()
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.01, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	return e, c
}

func TestGeminiGatewayRoutesThroughPrefrontalToN02AndClareira(t *testing.T) {
	e, c := newGeminiEngine(t)
	peer := &geminiFakePeer{}
	var events []GeminiExecutionEvent
	if err := RegisterGeminiOperations(e, peer, func(_ context.Context, event GeminiExecutionEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	candidateJSON, _ := json.Marshal(map[string]any{"id": "gemini-safe-action", "cost": .01, "risk": .01, "urgency": .1, "impact": .2})
	message := protocol.NewMessage("N01", "N07", "command", "gemini.delegate.text@1.0.0", nil)
	message.CorrelationID = "corr-gemini-test"
	message.Metadata["candidate_json"] = string(candidateJSON)
	message.Metadata["text"] = "analisar esta solicitação"

	result, err := e.Submit(context.Background(), message)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" {
		t.Fatalf("unexpected status: %#v", result)
	}
	if result.Metadata["route"] != "N07.prefrontal>N02" || result.Metadata["provider"] != "google-gemini" || result.Metadata["owner"] != "N02" {
		t.Fatalf("unexpected routing metadata: %#v", result.Metadata)
	}
	if result.Metadata["output_text"] != "Gemini real output" {
		t.Fatalf("unexpected output: %#v", result.Metadata)
	}
	if len(c.Recall(1)) != 1 {
		t.Fatal("expected prefrontal decision to be committed")
	}
	if len(peer.calls) != 2 || peer.calls[0].Capability != "neural.bnc_v2" || peer.calls[1].Capability != "gemini.text.generate" {
		t.Fatalf("unexpected N02 call chain: %#v", peer.calls)
	}
	for _, call := range peer.calls {
		if call.Nucleus != "N02" || call.Correlation != message.CorrelationID {
			t.Fatalf("invalid peer provenance: %#v", peer.calls)
		}
	}
	if len(events) != 2 || events[0].Phase != "started" || events[1].Phase != "completed" {
		t.Fatalf("unexpected lifecycle: %#v", events)
	}
}

func TestGeminiGatewayRejectsMalformedStructuredContentWithoutProviderCall(t *testing.T) {
	e, _ := newGeminiEngine(t)
	peer := &geminiFakePeer{}
	if err := RegisterGeminiOperations(e, peer, nil); err != nil {
		t.Fatal(err)
	}
	message := protocol.NewMessage("N01", "N07", "command", "gemini.delegate.multimodal@1.0.0", nil)
	message.Metadata["candidate_json"] = `{"id":"bad-content","cost":0.01,"risk":0.01,"urgency":0,"impact":0}`
	message.Metadata["text"] = "texto válido"
	message.Metadata["contents_json"] = "{malformed"

	_, err := e.Submit(context.Background(), message)
	if err == nil || !strings.Contains(err.Error(), "contents_json is invalid") {
		t.Fatalf("expected explicit malformed contents error, got %v", err)
	}
	if len(peer.calls) != 1 || peer.calls[0].Capability != "neural.bnc_v2" {
		t.Fatalf("provider must not execute: %#v", peer.calls)
	}
}

func TestGeminiGatewayRejectsRiskBeforeProvider(t *testing.T) {
	e, _ := newGeminiEngine(t)
	peer := &geminiFakePeer{}
	var events []GeminiExecutionEvent
	if err := RegisterGeminiOperations(e, peer, func(_ context.Context, event GeminiExecutionEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	message := protocol.NewMessage("N01", "N07", "command", "gemini.delegate.text@1.0.0", []float64{.1, .1, .1, .1, .1, .1, .1, .1})
	message.Metadata["candidate_json"] = `{"id":"risky","cost":0.01,"risk":1,"urgency":0,"impact":0}`
	message.Metadata["text"] = "não deve chegar ao provider"

	_, err := e.Submit(context.Background(), message)
	if err == nil {
		t.Fatal("expected risk gate rejection")
	}
	for _, call := range peer.calls {
		if strings.HasPrefix(call.Capability, "gemini.") {
			t.Fatalf("Gemini provider must not execute: %#v", peer.calls)
		}
	}
	if len(events) != 1 || events[0].Phase != "failed" {
		t.Fatalf("expected failed lifecycle event: %#v", events)
	}
}
