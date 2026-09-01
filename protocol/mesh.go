package protocol

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	SoulMeshVersion         = "1.0"
	SoulMeshContractVersion = "1.1.0"
)

type MeshEnvelope struct {
	Version        string         `json:"version"`
	ContractVersion string         `json:"contractVersion"`
	MessageID      string         `json:"messageId"`
	Source         string         `json:"source"`
	Target         string         `json:"target"`
	Timestamp      int64          `json:"timestamp"`
	Nonce          string         `json:"nonce"`
	CorrelationID  string         `json:"correlationId"`
	Type           string         `json:"type"`
	TTL            *int64         `json:"ttl,omitempty"`
	HMAC           string         `json:"hmac"`
	Payload        map[string]any `json:"payload"`
	// Operation is legacy-only. Canonical Mesh carries capability inside payload.
	Operation string `json:"operation,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

var nonceMu sync.Mutex
var seenNonces = map[string]int64{}

func (m MeshEnvelope) Validate() error {
	if m.Version != SoulMeshVersion { return errors.New("unsupported Mesh envelope version") }
	if m.ContractVersion != SoulMeshContractVersion { return errors.New("unsupported Mesh contract version") }
	if strings.TrimSpace(m.MessageID) == "" { return errors.New("messageId is required") }
	if strings.TrimSpace(m.CorrelationID) == "" { return errors.New("correlationId is required") }
	if strings.TrimSpace(m.Source) == "" || strings.TrimSpace(m.Target) == "" { return errors.New("source and target are required") }
	if m.Source == m.Target && m.Target != "BROADCAST" { return errors.New("source and target cannot be identical") }
	if strings.TrimSpace(m.Nonce) == "" { return errors.New("nonce is required") }
	if m.Timestamp <= 0 { return errors.New("timestamp is required") }
	if m.Type == "" { return errors.New("type is required") }
	if m.Payload == nil { return errors.New("payload is required") }
	return nil
}

func (m MeshEnvelope) Capability() string {
	if v, ok := m.Payload["capability"].(string); ok { return strings.TrimSpace(v) }
	return strings.TrimSpace(m.Operation)
}

func (m MeshEnvelope) NestedPayload() map[string]any {
	if p, ok := m.Payload["payload"].(map[string]any); ok { return p }
	return m.Payload
}

func canonicalUnsignedEnvelope(m MeshEnvelope) ([]byte, error) {
	return json.Marshal(struct {
		Version string `json:"version"`
		ContractVersion string `json:"contractVersion"`
		MessageID string `json:"messageId"`
		Source string `json:"source"`
		Target string `json:"target"`
		Timestamp int64 `json:"timestamp"`
		Nonce string `json:"nonce"`
		CorrelationID string `json:"correlationId"`
		Type string `json:"type"`
		TTL *int64 `json:"ttl,omitempty"`
		Payload map[string]any `json:"payload"`
	}{m.Version,m.ContractVersion,m.MessageID,m.Source,m.Target,m.Timestamp,m.Nonce,m.CorrelationID,m.Type,m.TTL,m.Payload})
}

func VerifyHMAC(m MeshEnvelope, secret string, now time.Time) error {
	if strings.TrimSpace(secret) == "" { return errors.New("Mesh HMAC secret is not configured") }
	if err := m.Validate(); err != nil { return err }
	if len(secret) < 16 { return errors.New("Mesh HMAC secret must contain at least 16 characters") }
	if delta := now.UnixMilli() - m.Timestamp; delta > 30000 || delta < -30000 { return errors.New("envelope timestamp outside accepted clock skew") }
	nonceMu.Lock()
	for nonce, expiry := range seenNonces { if expiry <= now.UnixMilli() { delete(seenNonces, nonce) } }
	if _, exists := seenNonces[m.Nonce]; exists { nonceMu.Unlock(); return errors.New("replay detected") }
	nonceMu.Unlock()
	unsigned, err := canonicalUnsignedEnvelope(m)
	if err != nil { return err }
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(unsigned)
	expected := hex.EncodeToString(mac.Sum(nil))
	provided, err := hex.DecodeString(m.HMAC)
	if err != nil || len(provided) != sha256.Size { return errors.New("invalid HMAC-SHA256 value") }
	if !hmac.Equal([]byte(expected), []byte(m.HMAC)) { return errors.New("invalid Mesh HMAC") }
	nonceMu.Lock(); seenNonces[m.Nonce] = now.UnixMilli() + 120000; nonceMu.Unlock()
	return nil
}

func EncodeMesh(m MeshEnvelope) ([]byte,error){if err:=m.Validate();err!=nil{return nil,err};return json.Marshal(m)}
func DecodeMesh(b []byte)(MeshEnvelope,error){var m MeshEnvelope;if err:=json.Unmarshal(b,&m);err!=nil{return m,err};return m,m.Validate()}
