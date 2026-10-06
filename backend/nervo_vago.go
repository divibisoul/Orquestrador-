package backend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	NervoVagoEventOperation   = "nervo-vago.event@1.0.0"
	NervoVagoDescribeOperation = "nervo-vago.describe@1.0.0"
)

type NervoVagoProvenance struct {
	TraceID       string `json:"trace_id"`
	CorrelationID string `json:"correlation_id"`
	MessageID     string `json:"message_id"`
	SequenceIndex uint64 `json:"sequence_index"`
	ParentHash    string `json:"parent_hash"`
	InputHash     string `json:"input_hash"`
	OutputHash    string `json:"output_hash"`
}

type NervoVagoEnvelope struct {
	VagusVersion  string `json:"vagus_version"`
	MessageID     string `json:"message_id"`
	CorrelationID string `json:"correlation_id"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Priority      int `json:"priority"`
	TTL           int64 `json:"ttl"`
	Type          string `json:"type"`
	Payload       map[string]any `json:"payload"`
	Provenance    NervoVagoProvenance `json:"provenance"`
}

type NervoVagoGateway struct {
	sara *SARAProxy
}

func NewNervoVagoGateway(sara *SARAProxy) (*NervoVagoGateway, error) {
	if sara == nil {
		return nil, errors.New("Nervo Vago requires SARA policy gateway")
	}
	return &NervoVagoGateway{sara: sara}, nil
}

func (g *NervoVagoGateway) Describe() map[string]any {
	state := "BLOCKED"
	if g != nil && g.sara != nil && g.sara.Configured() {
		state = "PROJECTED"
	}
	return map[string]any{
		"id":                    "nervo-vago",
		"logical_component":    true,
		"owner":                 "N07",
		"policy_authority":      "SARA/ETR",
		"event_protocol":        "soul.vagus.event.v1",
		"transport":             "SARA.VagusNerveBus",
		"durable_transport":     "NATS JetStream",
		"durable_transport_state": "PROJECTED",
		"canonical_request_transport": "soul-mesh/1",
		"no_second_vagus":       true,
		"no_second_mesh":        true,
		"etr_gate_before_delivery": true,
		"state":                 state,
	}
}

func (g *NervoVagoGateway) Publish(ctx context.Context, envelope NervoVagoEnvelope) (map[string]any, error) {
	if g == nil || g.sara == nil || !g.sara.Configured() {
		return nil, errors.New("NERVO_VAGO_BLOCKED:SARA_POLICY_GATE_UNCONFIGURED")
	}
	if ctx == nil {
		return nil, errors.New("NERVO_VAGO_BLOCKED:CONTEXT_REQUIRED")
	}
	if err := envelope.Validate(); err != nil {
		return nil, err
	}

	envelope.CorrelationID = strings.TrimSpace(envelope.CorrelationID)
	envelope.Provenance.CorrelationID = envelope.CorrelationID
	provenanceInput, err := canonicalEnvelopeForHash(envelope, true)
	if err != nil {
		return nil, fmt.Errorf("NERVO_VAGO_PROVENANCE_ENCODE:%w", err)
	}
	expectedOutputHash := sha256Hex(provenanceInput)
	if envelope.Provenance.OutputHash == "" {
		envelope.Provenance.OutputHash = expectedOutputHash
	} else if envelope.Provenance.OutputHash != expectedOutputHash {
		return nil, errors.New("NERVO_VAGO_BLOCKED:OUTPUT_HASH_MISMATCH")
	}

	raw, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("NERVO_VAGO_ENVELOPE_ENCODE:%w", err)
	}
	audit, err := g.sara.Audit(ctx, string(raw), envelope.CorrelationID)
	if err != nil {
		return nil, fmt.Errorf("NERVO_VAGO_BLOCKED:ETR_AUDIT_FAILED:%w", err)
	}
	ethical, ok := audit["ethical"].(map[string]any)
	if !ok {
		return nil, errors.New("NERVO_VAGO_BLOCKED:ETR_RESULT_MISSING")
	}
	approved, ok := ethical["approved"].(bool)
	if !ok {
		return nil, errors.New("NERVO_VAGO_BLOCKED:ETR_APPROVAL_MISSING")
	}
	if !approved {
		return nil, errors.New("NERVO_VAGO_BLOCKED:ETR_REJECTED")
	}

	published, err := g.sara.PublishVagus(ctx, map[string]any{
		"vagus_version":  envelope.VagusVersion,
		"message_id":     envelope.MessageID,
		"correlation_id": envelope.CorrelationID,
		"source":         envelope.Source,
		"target":         envelope.Target,
		"priority":       envelope.Priority,
		"ttl":            envelope.TTL,
		"type":           envelope.Type,
		"payload":        envelope.Payload,
		"status":         "EXECUTE",
		"provenance":     envelope.Provenance,
	}, envelope.CorrelationID)
	if err != nil {
		return nil, fmt.Errorf("NERVO_VAGO_BLOCKED:VAGUS_PUBLISH_FAILED:%w", err)
	}

	return map[string]any{
		"operation":        "nervo-vago.event",
		"state":            "REAL",
		"event":            published,
		"provenance":       envelope.Provenance,
		"etr":              ethical,
		"etr_approved":     true,
		"transport":        "SARA.VagusNerveBus",
		"canonical_control": "N07",
	}, nil
}

func (e NervoVagoEnvelope) Validate() error {
	if e.VagusVersion != "1.0" {
		return errors.New("NERVO_VAGO_INVALID_VERSION")
	}
	required := map[string]string{
		"message_id": e.MessageID,
		"correlation_id": e.CorrelationID,
		"source": e.Source,
		"target": e.Target,
		"type": e.Type,
	}
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("NERVO_VAGO_REQUIRED:%s", key)
		}
	}
	if e.Priority < 0 || e.Priority > 100 || e.TTL <= 0 {
		return errors.New("NERVO_VAGO_INVALID_PRIORITY_TTL")
	}
	if e.Payload == nil {
		return errors.New("NERVO_VAGO_PAYLOAD_REQUIRED")
	}
	if e.Provenance.TraceID == "" || e.Provenance.MessageID != e.MessageID || e.Provenance.CorrelationID != e.CorrelationID {
		return errors.New("NERVO_VAGO_PROVENANCE_IDENTITY_INVALID")
	}
	if e.Provenance.SequenceIndex == 0 || e.Provenance.ParentHash == "" || e.Provenance.InputHash == "" {
		return errors.New("NERVO_VAGO_PROVENANCE_INCOMPLETE")
	}
	if !(e.Type == "gpu.submit" || e.Type == "gpu.result" || e.Type == "gpu.barrier" ||
		strings.HasPrefix(e.Type, "health.") || strings.HasPrefix(e.Type, "capability.") ||
		strings.HasPrefix(e.Type, "signal.") || strings.HasPrefix(e.Type, "sara.") ||
		strings.HasPrefix(e.Type, "session.") || strings.HasPrefix(e.Type, "research.") ||
		strings.HasPrefix(e.Type, "nervo.")) {
		return fmt.Errorf("NERVO_VAGO_TYPE_UNSUPPORTED:%s", e.Type)
	}
	return nil
}

func canonicalEnvelopeForHash(e NervoVagoEnvelope, withoutOutputHash bool) ([]byte, error) {
	if withoutOutputHash {
		e.Provenance.OutputHash = ""
	}
	return json.Marshal(e)
}

func sha256Hex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func RegisterNervoVagoOperations(e *orchestrator.Engine, sara *SARAProxy) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	gateway, err := NewNervoVagoGateway(sara)
	if err != nil {
		return err
	}

	if err := e.Register(NervoVagoDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		_ = ctx
		raw, err := json.Marshal(gateway.Describe())
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID:       m.TraceID,
			CorrelationID: m.CorrelationID,
			Source:        "N07.NervoVago",
			Target:        m.Source,
			Status:        "ok",
			Metadata:      map[string]string{"nervo_vago_json": string(raw)},
		}, nil
	}); err != nil {
		return err
	}

	if err := e.Register(NervoVagoEventOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		payloadRaw := strings.TrimSpace(m.Metadata["payload_json"])
		if payloadRaw == "" {
			payloadRaw = "{}"
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
			return protocol.Result{}, errors.New("NERVO_VAGO_PAYLOAD_JSON_INVALID")
		}
		sequenceRaw := strings.TrimSpace(m.Metadata["sequence_index"])
		var sequence uint64
		if _, err := fmt.Sscanf(sequenceRaw, "%d", &sequence); err != nil || sequence == 0 {
			return protocol.Result{}, errors.New("NERVO_VAGO_SEQUENCE_INDEX_INVALID")
		}
		source := strings.TrimSpace(m.Metadata["source"])
		if source == "" {
			source = strings.TrimSpace(m.Source)
		}
		envelope := NervoVagoEnvelope{
			VagusVersion:  "1.0",
			MessageID:     strings.TrimSpace(m.Metadata["message_id"]),
			CorrelationID: strings.TrimSpace(m.CorrelationID),
			Source:        source,
			Target:        strings.TrimSpace(m.Metadata["target"]),
			Priority:      parseInt(m.Metadata["priority"], 50),
			TTL:           parseInt64(m.Metadata["ttl"], 5000),
			Type:          strings.TrimSpace(m.Metadata["type"]),
			Payload:       payload,
			Provenance: NervoVagoProvenance{
				TraceID:       strings.TrimSpace(m.TraceID),
				CorrelationID: strings.TrimSpace(m.CorrelationID),
				MessageID:     strings.TrimSpace(m.Metadata["message_id"]),
				SequenceIndex: sequence,
				ParentHash:    strings.TrimSpace(m.Metadata["parent_hash"]),
				InputHash:     strings.TrimSpace(m.Metadata["input_hash"]),
				OutputHash:    strings.TrimSpace(m.Metadata["output_hash"]),
			},
		}
		out, err := gateway.Publish(ctx, envelope)
		if err != nil {
			return protocol.Result{
				TraceID:       m.TraceID,
				CorrelationID: m.CorrelationID,
				Source:        "N07.NervoVago",
				Target:        m.Source,
				Status:        "blocked",
				Error:         err.Error(),
			}, err
		}
		raw, err := json.Marshal(out)
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{
			TraceID:       m.TraceID,
			CorrelationID: m.CorrelationID,
			Source:        "N07.NervoVago",
			Target:        m.Source,
			Status:        "ok",
			Metadata:      map[string]string{
				"nervo_vago_json": rawString(raw),
				"epistemic_state": "REAL",
			},
		}, nil
	}); err != nil {
		return err
	}
	return nil
}

func parseInt(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	var v int
	if _, err := fmt.Sscanf(raw, "%d", &v); err != nil {
		return fallback
	}
	return v
}

func parseInt64(raw string, fallback int64) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	var v int64
	if _, err := fmt.Sscanf(raw, "%d", &v); err != nil {
		return fallback
	}
	return v
}

func rawString(raw []byte) string {
	return string(raw)
}

var _ = sort.Strings
