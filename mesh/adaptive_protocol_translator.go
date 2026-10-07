package mesh

import (
	"errors"
	"fmt"
	"strings"
)

const (
	adaptiveMeshProtocol  = "soul-mesh/1"
	canonicalMeshContract = "1.1.0"
	legacyMeshContract    = "1.2.0"
)

// AdaptiveProtocolTranslator is a compatibility capability for legacy Mesh
// envelopes. It normalizes 1.2.0 wire fields to the live 1.1.0 contract without
// bypassing the canonical Mesh gateway. Legacy authenticated messages MUST be
// re-signed by the caller after translation; authentication is never preserved
// implicitly across protocol versions.
type AdaptiveProtocolTranslator struct{}

type TranslatedEnvelope struct {
	Protocol          string
	ContractVersion   string
	CorrelationID     string
	Source            string
	Target            string
	Payload           map[string]any
	SourceVersion     string
	NeedsResign       bool
	TranslationReason string
}

func (AdaptiveProtocolTranslator) Translate(packet map[string]any) (TranslatedEnvelope, error) {
	if packet == nil {
		return TranslatedEnvelope{}, errors.New("Mesh packet is nil")
	}
	if protocol := stringValue(packet, "protocol"); protocol != adaptiveMeshProtocol {
		return TranslatedEnvelope{}, fmt.Errorf("unsupported Mesh protocol: %q", protocol)
	}

	version := firstString(packet, "contractVersion", "contract_version")
	if version != canonicalMeshContract && version != legacyMeshContract {
		return TranslatedEnvelope{}, fmt.Errorf("unsupported Mesh contract version: %q", version)
	}

	correlationID := firstString(packet, "correlationId", "correlation_id", "id", "trace_id")
	source := firstString(packet, "source", "origin")
	target := firstString(packet, "target", "destination")
	if correlationID == "" {
		return TranslatedEnvelope{}, errors.New("correlationId is required for adaptive translation")
	}
	if source == "" || target == "" {
		return TranslatedEnvelope{}, errors.New("source and target are required for adaptive translation")
	}

	payload := mapValue(packet, "payload")
	if payload == nil {
		payload = mapValue(packet, "data")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload = cloneAdaptivePayload(payload)

	if version == legacyMeshContract {
		delete(payload, "_upgraded")
		delete(payload, "ttl")
		return TranslatedEnvelope{
			Protocol:          adaptiveMeshProtocol,
			ContractVersion:   canonicalMeshContract,
			CorrelationID:     correlationID,
			Source:            source,
			Target:            target,
			Payload:           payload,
			SourceVersion:     version,
			NeedsResign:       true,
			TranslationReason: "legacy-1.2.0-to-canonical-1.1.0",
		}, nil
	}

	return TranslatedEnvelope{
		Protocol:          adaptiveMeshProtocol,
		ContractVersion:   canonicalMeshContract,
		CorrelationID:     correlationID,
		Source:            source,
		Target:            target,
		Payload:           payload,
		SourceVersion:     version,
		NeedsResign:       false,
		TranslationReason: "canonical-1.1.0",
	}, nil
}

func firstString(packet map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(packet, key); value != "" {
			return value
		}
	}
	return ""
}

func stringValue(packet map[string]any, key string) string {
	value, ok := packet[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func mapValue(packet map[string]any, key string) map[string]any {
	value, ok := packet[key].(map[string]any)
	if !ok {
		return nil
	}
	return value
}

func cloneAdaptivePayload(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
