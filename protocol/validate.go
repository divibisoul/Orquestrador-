package protocol

import (
	"fmt"
	"strings"
)

const InvalidMeshContract = "INVALID_MESH_CONTRACT"

type MeshContractError struct {
	Field  string
	Reason string
}

func (e *MeshContractError) Error() string {
	if e == nil {
		return InvalidMeshContract
	}
	return fmt.Sprintf("%s: field %s %s", InvalidMeshContract, e.Field, e.Reason)
}

func invalidMeshField(field, reason string) error {
	return &MeshContractError{Field: field, Reason: reason}
}

// ValidateMeshEnvelope applies the canonical structural contract. Signing paths
// may set requireHMAC=false until the signature is generated; receiving paths
// must require it.
func ValidateMeshEnvelope(m MeshEnvelope, requireHMAC bool) error {
	if m.Version != SoulMeshVersion {
		return invalidMeshField("version", "must equal "+SoulMeshVersion)
	}
	if m.ContractVersion != SoulMeshContractVersion {
		return invalidMeshField("contractVersion", "must equal "+SoulMeshContractVersion)
	}
	if strings.TrimSpace(m.MessageID) == "" {
		return invalidMeshField("id", "is required")
	}
	if strings.TrimSpace(m.CorrelationID) == "" {
		return invalidMeshField("correlationId", "is required")
	}
	if strings.TrimSpace(m.Source) == "" {
		return invalidMeshField("source", "is required")
	}
	if _, ok := validNuclei[m.Source]; !ok {
		return invalidMeshField("source", "contains an unknown nucleus")
	}
	if strings.TrimSpace(m.Target) == "" {
		return invalidMeshField("target", "is required")
	}
	if m.Target != "BROADCAST" {
		if _, ok := validNuclei[m.Target]; !ok {
			return invalidMeshField("target", "contains an unknown nucleus")
		}
	}
	if m.Source == m.Target && m.Target != "BROADCAST" {
		return invalidMeshField("target", "must differ from source")
	}
	if strings.TrimSpace(m.Type) == "" {
		return invalidMeshField("kind", "is required")
	}
	if strings.TrimSpace(m.Capability()) == "" {
		return invalidMeshField("capability", "is required")
	}
	if m.Payload == nil {
		return invalidMeshField("payload", "is required")
	}
	if m.Timestamp <= 0 {
		return invalidMeshField("timestamp", "must be positive")
	}
	if strings.TrimSpace(m.Nonce) == "" {
		return invalidMeshField("nonce", "is required")
	}
	if requireHMAC && strings.TrimSpace(m.HMAC) == "" {
		return invalidMeshField("hmac", "is required")
	}
	if m.TTL != nil && *m.TTL < 0 {
		return invalidMeshField("ttl", "cannot be negative")
	}
	return nil
}

// ValidateMeshWireResponse validates the JSON shape emitted by the federation
// peers before any cryptographic verification is attempted.
func ValidateMeshWireResponse(result map[string]any) error {
	if result == nil {
		return invalidMeshField("envelope", "is required")
	}
	for _, field := range []string{"id", "correlationId", "source", "target", "kind", "capability", "payload", "timestamp", "nonce", "hmac"} {
		if _, ok := result[field]; !ok {
			return invalidMeshField(field, "is required")
		}
	}
	protocolName, ok := result["protocol"].(string)
	if !ok || strings.TrimSpace(protocolName) != "soul-mesh/1" {
		return invalidMeshField("protocol", "must equal soul-mesh/1")
	}
	contractVersion, ok := result["contractVersion"].(string)
	if !ok || strings.TrimSpace(contractVersion) != SoulMeshContractVersion {
		return invalidMeshField("contractVersion", "must equal "+SoulMeshContractVersion)
	}
	id, ok := result["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return invalidMeshField("id", "must be a non-empty string")
	}
	correlationID, ok := result["correlationId"].(string)
	if !ok || strings.TrimSpace(correlationID) == "" {
		return invalidMeshField("correlationId", "must be a non-empty string")
	}
	source, ok := result["source"].(string)
	if !ok || strings.TrimSpace(source) == "" {
		return invalidMeshField("source", "must be a non-empty string")
	}
	target, ok := result["target"].(string)
	if !ok || strings.TrimSpace(target) == "" {
		return invalidMeshField("target", "must be a non-empty string")
	}
	kind, ok := result["kind"].(string)
	if !ok || (kind != "response" && kind != "error") {
		return invalidMeshField("kind", "must be response or error")
	}
	capability, ok := result["capability"].(string)
	if !ok || strings.TrimSpace(capability) == "" {
		return invalidMeshField("capability", "must be a non-empty string")
	}
	if _, ok := result["payload"].(map[string]any); !ok {
		return invalidMeshField("payload", "must be an object")
	}
	timestamp, ok := result["timestamp"].(float64)
	if !ok || timestamp <= 0 {
		return invalidMeshField("timestamp", "must be a positive number")
	}
	nonce, ok := result["nonce"].(string)
	if !ok || strings.TrimSpace(nonce) == "" {
		return invalidMeshField("nonce", "must be a non-empty string")
	}
	hmacValue, ok := result["hmac"].(string)
	if !ok || strings.TrimSpace(hmacValue) == "" {
		return invalidMeshField("hmac", "must be a non-empty string")
	}
	if _, ok := validNuclei[source]; !ok {
		return invalidMeshField("source", "contains an unknown nucleus")
	}
	if target != "BROADCAST" {
		if _, ok := validNuclei[target]; !ok {
			return invalidMeshField("target", "contains an unknown nucleus")
		}
	}
	if source == target && target != "BROADCAST" {
		return invalidMeshField("target", "must differ from source")
	}
	_ = id
	_ = correlationID
	return nil
}
