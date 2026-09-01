package protocol

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const SoulMeshVersion = "1.0"
const (
	N01 = "N01"
	N02 = "N02"
	N03 = "N03"
	N04 = "N04"
	N05 = "N05"
	N06 = "N06"
	N07 = "N07"
)

var validNuclei = map[string]struct{}{N01: {}, N02: {}, N03: {}, N04: {}, N05: {}, N06: {}, N07: {}}

type MeshEnvelope struct {
	Version         string            `json:"version"`
	ContractVersion string            `json:"contractVersion"`
	MessageID       string            `json:"messageId"`
	Source          string            `json:"source"`
	Target          string            `json:"target"`
	Timestamp       int64             `json:"timestamp"`
	Nonce           string            `json:"nonce"`
	CorrelationID   string            `json:"correlationId"`
	Type            string            `json:"type"`
	TTL             *int64            `json:"ttl,omitempty"`
	HMAC            string            `json:"hmac"`
	Payload         map[string]any    `json:"payload"`
	Operation       string            `json:"operation,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

var nonceMu sync.Mutex
var seenNonces = map[string]int64{}

func (m MeshEnvelope) Validate() error {
	if m.Version != SoulMeshVersion {
		return errors.New("unsupported Mesh envelope version")
	}
	if m.ContractVersion != SoulMeshContractVersion {
		return errors.New("unsupported Mesh contract version")
	}
	if strings.TrimSpace(m.MessageID) == "" {
		return errors.New("messageId is required")
	}
	if strings.TrimSpace(m.CorrelationID) == "" {
		return errors.New("correlationId is required")
	}
	if strings.TrimSpace(m.Source) == "" || strings.TrimSpace(m.Target) == "" {
		return errors.New("source and target are required")
	}
	if _, ok := validNuclei[m.Source]; !ok {
		return errors.New("unknown source nucleus")
	}
	if m.Target != "BROADCAST" {
		if _, ok := validNuclei[m.Target]; !ok {
			return errors.New("unknown target nucleus")
		}
	}
	if m.Source == m.Target && m.Target != "BROADCAST" {
		return errors.New("source and target cannot be identical")
	}
	if strings.TrimSpace(m.Nonce) == "" {
		return errors.New("nonce is required")
	}
	if m.Timestamp <= 0 {
		return errors.New("timestamp is required")
	}
	if strings.TrimSpace(m.Type) == "" {
		return errors.New("type is required")
	}
	if m.Payload == nil {
		return errors.New("payload is required")
	}
	if m.TTL != nil && *m.TTL < 0 {
		return errors.New("ttl cannot be negative")
	}
	return nil
}

func (m MeshEnvelope) Capability() string {
	if v, ok := m.Payload["capability"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func (m MeshEnvelope) NestedPayload() map[string]any {
	if p, ok := m.Payload["payload"].(map[string]any); ok {
		return p
	}
	return m.Payload
}

func (m MeshEnvelope) NestedMetadata() map[string]string {
	if metadata, ok := m.Payload["metadata"].(map[string]any); ok {
		out := make(map[string]string, len(metadata))
		for key, value := range metadata {
			if text, ok := value.(string); ok {
				out[key] = text
			}
		}
		return out
	}
	return m.Metadata
}

func canonicalUnsignedEnvelope(m MeshEnvelope) ([]byte, error) {
	return json.Marshal(struct {
		Version         string            `json:"version"`
		ContractVersion string            `json:"contractVersion"`
		MessageID       string            `json:"messageId"`
		Source          string            `json:"source"`
		Target          string            `json:"target"`
		Timestamp       int64             `json:"timestamp"`
		Nonce           string            `json:"nonce"`
		CorrelationID   string            `json:"correlationId"`
		Type            string            `json:"type"`
		TTL             *int64            `json:"ttl,omitempty"`
		Payload         map[string]any    `json:"payload"`
		Operation       string            `json:"operation,omitempty"`
		Metadata        map[string]string `json:"metadata,omitempty"`
	}{m.Version, m.ContractVersion, m.MessageID, m.Source, m.Target, m.Timestamp, m.Nonce, m.CorrelationID, m.Type, m.TTL, m.Payload, m.Operation, m.Metadata})
}

func SignHMAC(m *MeshEnvelope, secret string) error {
	if m == nil {
		return errors.New("Mesh envelope is nil")
	}
	if err := m.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(secret) == "" {
		return errors.New("Mesh HMAC secret is not configured")
	}
	if len(secret) < 16 {
		return errors.New("Mesh HMAC secret must contain at least 16 characters")
	}
	unsigned, err := canonicalUnsignedEnvelope(*m)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(unsigned)
	m.HMAC = hex.EncodeToString(mac.Sum(nil))
	return nil
}

func VerifyHMAC(m MeshEnvelope, secret string, now time.Time) error {
	if strings.TrimSpace(secret) == "" {
		return errors.New("Mesh HMAC secret is not configured")
	}
	if err := m.Validate(); err != nil {
		return err
	}
	if len(secret) < 16 {
		return errors.New("Mesh HMAC secret must contain at least 16 characters")
	}
	if delta := now.UnixMilli() - m.Timestamp; delta > 30000 || delta < -30000 {
		return errors.New("envelope timestamp outside accepted clock skew")
	}
	unsigned, err := canonicalUnsignedEnvelope(m)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(unsigned)
	expected, err := hex.DecodeString(m.HMAC)
	if err != nil {
		return fmt.Errorf("invalid HMAC-SHA256 encoding: %w", err)
	}
	if len(expected) != sha256.Size {
		return fmt.Errorf("invalid HMAC-SHA256 value: expected %d bytes, got %d", sha256.Size, len(expected))
	}
	if !hmac.Equal(mac.Sum(nil), expected) {
		return errors.New("invalid Mesh HMAC")
	}
	key := m.Source + "\x00" + m.Nonce
	nonceMu.Lock()
	defer nonceMu.Unlock()
	nowMS := now.UnixMilli()
	for nonce, expiry := range seenNonces {
		if expiry <= nowMS {
			delete(seenNonces, nonce)
		}
	}
	if _, exists := seenNonces[key]; exists {
		return errors.New("replay detected")
	}
	seenNonces[key] = nowMS + 120000
	return nil
}

func EncodeMesh(m MeshEnvelope) ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func DecodeMesh(b []byte) (MeshEnvelope, error) {
	var m MeshEnvelope
	if err := json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	return m, m.Validate()
}
