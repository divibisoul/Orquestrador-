package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNervoVagoGRCEFeedbackAuditsHashLinkedSummaryAndPublishesFullEvidence(t *testing.T) {
	var auditedInput string
	var deliveredPayload map[string]any
	var calls []string
	longInput := strings.Repeat("a", 11000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/audit":
			calls = append(calls, "audit")
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			auditedInput, _ = body["input"].(string)
			_, _ = w.Write([]byte("{\"ethical\":{\"approved\":true},\"operation\":\"audit\"}"))
		case "/v1/vagus":
			calls = append(calls, "vagus")
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			deliveredPayload, _ = body["payload"].(map[string]any)
			_, _ = w.Write([]byte("{\"operation\":\"vagus\",\"state\":\"REAL\"}"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	gateway, err := NewNervoVagoGateway(NewSARAProxy(Config{
		SARAServiceURL: server.URL, SARAServiceToken: "test-token",
	}))
	if err != nil {
		t.Fatal(err)
	}
	envelope := NervoVagoEnvelope{
		VagusVersion: "1.0", MessageID: "grce-message-1", CorrelationID: "grce-correlation-1",
		Source: "N07.GRCE", Target: "AETERNUM_HORTACORE", Priority: 100, TTL: 5000,
		Type: "nervo.grce.feedback.horta",
		Payload: map[string]any{
			"kind": "horta", "state_id": "state-grce-1", "feedback": "GRCE",
			"provenance": []map[string]any{{
				"parent_hash": "parent-hash", "input_hash": "input-hash", "output_hash": "output-hash",
				"sequence_index": 3, "stage": "ARA", "causal_failure_id": "failure-1",
			}},
			"evidence": []map[string]any{{
				"id": "evidence-1", "failure_id": "failure-1", "state": "PROJECTED",
				"hash": "evidence-hash", "input_hash": "input-hash", "output_hash": "output-hash",
				"sequence_index": 3, "payload": map[string]any{"input": longInput, "detail": "full original evidence"},
			}},
			"capabilities": []map[string]any{{"id": "capability-1", "state": "PROJECTED"}},
		},
		Provenance: NervoVagoProvenance{
			TraceID: "grce-trace-1", CorrelationID: "grce-correlation-1", MessageID: "grce-message-1",
			SequenceIndex: 4, ParentHash: "parent-hash", InputHash: "input-hash",
		},
	}
	_, err = gateway.PublishGRCEFeedback(context.Background(), envelope)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(calls, ","), "audit,vagus"; got != want {
		t.Fatalf("delivery order = %q, want %q", got, want)
	}
	if !strings.Contains(auditedInput, "payload_sha256") || !strings.Contains(auditedInput, "envelope_sha256") {
		t.Fatalf("ETR audit did not receive payload and envelope hashes: %s", auditedInput)
	}
	if strings.Contains(auditedInput, longInput) {
		t.Fatal("ETR summary must not duplicate long user content into the transport-policy review")
	}
	if deliveredPayload == nil {
		t.Fatal("full GRCE payload was not delivered")
	}
	deliveredEvidence, ok := deliveredPayload["evidence"].([]any)
	if !ok || len(deliveredEvidence) != 1 {
		t.Fatalf("full evidence list was not preserved: %#v", deliveredPayload["evidence"])
	}
	item, ok := deliveredEvidence[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected evidence payload type: %#v", deliveredEvidence[0])
	}
	preserved, ok := item["payload"].(map[string]any)
	if !ok || preserved["input"] != longInput {
		t.Fatal("full original evidence payload was not preserved through delivery")
	}
}

func TestNervoVagoGRCEFeedbackStillBlocksWhenETRRejects(t *testing.T) {
	var vagusCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/audit":
			_, _ = w.Write([]byte("{\"ethical\":{\"approved\":false},\"operation\":\"audit\"}"))
		case "/v1/vagus":
			vagusCalls++
			_, _ = w.Write([]byte("{\"operation\":\"vagus\"}"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	gateway, err := NewNervoVagoGateway(NewSARAProxy(Config{
		SARAServiceURL: server.URL, SARAServiceToken: "test-token",
	}))
	if err != nil {
		t.Fatal(err)
	}
	envelope := NervoVagoEnvelope{
		VagusVersion: "1.0", MessageID: "grce-message-reject", CorrelationID: "grce-correlation-reject",
		Source: "N07.GRCE", Target: "AETERNUM_HORTACORE", Priority: 100, TTL: 5000,
		Type: "nervo.grce.feedback.horta",
		Payload: map[string]any{
			"kind": "horta", "state_id": "state-grce-reject", "feedback": "GRCE",
			"provenance": []map[string]any{{"parent_hash": "p", "input_hash": "i", "output_hash": "o", "sequence_index": 1, "stage": "ARA"}},
			"evidence": []map[string]any{{"id": "e", "failure_id": "f", "hash": "h", "input_hash": "i", "output_hash": "o", "state": "PROJECTED", "sequence_index": 1}},
			"capabilities": []map[string]any{},
		},
		Provenance: NervoVagoProvenance{
			TraceID: "grce-trace-reject", CorrelationID: "grce-correlation-reject", MessageID: "grce-message-reject",
			SequenceIndex: 2, ParentHash: "p", InputHash: "i",
		},
	}
	_, err = gateway.PublishGRCEFeedback(context.Background(), envelope)
	if err == nil || !strings.Contains(err.Error(), "NERVO_VAGO_BLOCKED:ETR_REJECTED") {
		t.Fatalf("expected ethical rejection, got %v", err)
	}
	if vagusCalls != 0 {
		t.Fatalf("Vagus delivered an ethically rejected event %d times", vagusCalls)
	}
}

func TestNervoVagoPublishRunsETRBeforeVagusDelivery(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/audit":
			calls = append(calls, "audit")
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			input, ok := body["input"].(string)
			if !ok || input == "" {
				t.Fatal("audit input missing or not a string")
			}
			for _, required := range []string{
				"Princípios do ciclo GRCE",
				"Finalidade da avaliação",
				"evento do Nervo Vago",
				`"message_id":"m-1"`,
			} {
				if !strings.Contains(input, required) {
					t.Fatalf("ETR input does not preserve required context %q", required)
				}
			}
			_, _ = w.Write([]byte(`{"ethical":{"approved":true},"operation":"audit"}`))
		case "/v1/vagus":
			calls = append(calls, "vagus")
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["message_id"] != "m-1" {
				t.Fatalf("unexpected message_id: %#v", body["message_id"])
			}
			_, _ = w.Write([]byte(`{"operation":"vagus","event":{"message_id":"m-1"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	proxy := NewSARAProxy(Config{
		SARAServiceURL:   server.URL,
		SARAServiceToken: "test-token",
	})
	gateway, err := NewNervoVagoGateway(proxy)
	if err != nil {
		t.Fatal(err)
	}

	envelope := NervoVagoEnvelope{
		VagusVersion:  "1.0",
		MessageID:     "m-1",
		CorrelationID: "c-1",
		Source:        "N07",
		Target:        "SARA",
		Priority:      100,
		TTL:           5000,
		Type:          "nervo.event",
		Payload:       map[string]any{"kind": "test"},
		Provenance: NervoVagoProvenance{
			TraceID:       "t-1",
			CorrelationID: "c-1",
			MessageID:     "m-1",
			SequenceIndex: 1,
			ParentHash:    "parent",
			InputHash:     "input",
		},
	}

	out, err := gateway.Publish(context.Background(), envelope)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(calls, ","), "audit,vagus"; got != want {
		t.Fatalf("delivery order = %q, want %q", got, want)
	}
	if out["state"] != "REAL" {
		t.Fatalf("unexpected state: %#v", out["state"])
	}
	if out["etr_approved"] != true {
		t.Fatalf("ETR approval not recorded: %#v", out["etr_approved"])
	}
	prov, ok := out["provenance"].(NervoVagoProvenance)
	if !ok {
		t.Fatalf("unexpected provenance type: %#v", out["provenance"])
	}
	if prov.OutputHash == "" {
		t.Fatal("output hash not generated")
	}
}

func TestNervoVagoBlocksDeliveryWhenETRRejects(t *testing.T) {
	var vagusCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/audit":
			_, _ = w.Write([]byte(`{"ethical":{"approved":false},"operation":"audit"}`))
		case "/v1/vagus":
			vagusCalls++
			_, _ = w.Write([]byte(`{"operation":"vagus"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	proxy := NewSARAProxy(Config{SARAServiceURL: server.URL, SARAServiceToken: "test-token"})
	gateway, err := NewNervoVagoGateway(proxy)
	if err != nil {
		t.Fatal(err)
	}
	envelope := NervoVagoEnvelope{
		VagusVersion: "1.0", MessageID: "m-2", CorrelationID: "c-2",
		Source: "N07", Target: "N01", Priority: 50, TTL: 1000,
		Type: "nervo.event", Payload: map[string]any{"kind": "reject"},
		Provenance: NervoVagoProvenance{
			TraceID: "t-2", CorrelationID: "c-2", MessageID: "m-2",
			SequenceIndex: 1, ParentHash: "parent", InputHash: "input",
		},
	}
	_, err = gateway.Publish(context.Background(), envelope)
	if err == nil || !strings.Contains(err.Error(), "NERVO_VAGO_BLOCKED:ETR_REJECTED") {
		t.Fatalf("expected ETR rejection, got %v", err)
	}
	if vagusCalls != 0 {
		t.Fatalf("Vagus delivery occurred after ETR rejection: %d", vagusCalls)
	}
}

func TestNervoVagoBlocksWhenSARAIsUnconfigured(t *testing.T) {
	gateway, err := NewNervoVagoGateway(&SARAProxy{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = gateway.Publish(context.Background(), NervoVagoEnvelope{})
	if err == nil || !strings.Contains(err.Error(), "NERVO_VAGO_BLOCKED:SARA_POLICY_GATE_UNCONFIGURED") {
		t.Fatalf("expected unconfigured gate block, got %v", err)
	}
}

func TestNervoVagoAcceptsSuperGPUEvents(t *testing.T) {
	for _, eventType := range []string{"supergpu.submit", "supergpu.result", "supergpu.barrier"} {
		envelope := NervoVagoEnvelope{
			VagusVersion: "1.0", MessageID: "m-" + eventType, CorrelationID: "c-" + eventType,
			Source: "N07.SuperGPU", Target: "SARA", Priority: 100, TTL: 1000,
			Type: eventType, Payload: map[string]any{"kind": eventType},
			Provenance: NervoVagoProvenance{TraceID: "t-" + eventType, CorrelationID: "c-" + eventType, MessageID: "m-" + eventType, SequenceIndex: 1, ParentHash: "parent", InputHash: "input"},
		}
		if err := envelope.Validate(); err != nil {
			t.Fatalf("event type %q rejected: %v", eventType, err)
		}
	}
}

func TestNervoVagoStrictNumericParsing(t *testing.T) {
	if _, err := parseIntStrict("50x", 0); err == nil {
		t.Fatal("priority parser accepted trailing non-numeric data")
	}
	if _, err := parseInt64Strict("5000x", 0); err == nil {
		t.Fatal("ttl parser accepted trailing non-numeric data")
	}
}


func TestHashNervoVagoValueDeterministic(t *testing.T) {
	first := map[string]any{"b": 2, "a": 1}
	second := map[string]any{}
	second["a"] = 1
	second["b"] = 2

	firstHash, err := HashNervoVagoValue(first)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := HashNervoVagoValue(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash == "" || firstHash != secondHash {
		t.Fatalf("equivalent JSON objects must produce the same hash: %q != %q", firstHash, secondHash)
	}

	changedHash, err := HashNervoVagoValue(map[string]any{"a": 1, "b": 3})
	if err != nil {
		t.Fatal(err)
	}
	if changedHash == firstHash {
		t.Fatal("different JSON values must not produce the same hash")
	}
}

func TestHashNervoVagoValueRejectsUnsupportedValue(t *testing.T) {
	if _, err := HashNervoVagoValue(map[string]any{"unsupported": make(chan int)}); err == nil {
		t.Fatal("unsupported JSON value must fail closed")
	}
}
