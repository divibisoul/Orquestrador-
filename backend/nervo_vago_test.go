package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
			if _, ok := body["input"]; !ok {
				t.Fatal("audit input missing")
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
