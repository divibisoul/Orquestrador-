package grf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// FlexiblePayload carries either the canonical object form or the SOUL-28
// byte form. Raw bytes are retained unchanged until a caller explicitly asks
// to encode a map; invalid JSON is never silently converted to a fake object.
type FlexiblePayload struct {
	raw    []byte
	object map[string]any
	isMap  bool
}

// NewFlexiblePayloadFromMap constructs a payload from the canonical N07 form.
func NewFlexiblePayloadFromMap(value map[string]any) FlexiblePayload {
	object := make(map[string]any, len(value))
	for key, item := range value {
		object[key] = item
	}
	return FlexiblePayload{object: object, isMap: true}
}

// NewFlexiblePayloadFromBytes constructs a payload from the legacy SOUL-28 form.
func NewFlexiblePayloadFromBytes(value []byte) FlexiblePayload {
	return FlexiblePayload{raw: append([]byte(nil), value...)}
}

// ToMap obtains an independent JSON object. It returns an error for invalid JSON,
// a non-object JSON value, an empty byte slice, or trailing non-JSON content.
func (p FlexiblePayload) ToMap() (map[string]any, error) {
	var raw []byte
	if p.isMap {
		encoded, err := json.Marshal(p.object)
		if err != nil {
			return nil, fmt.Errorf("FLEXIBLE_PAYLOAD_MAP_ENCODE_FAILED: %w", err)
		}
		raw = encoded
	} else {
		raw = append([]byte(nil), p.raw...)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, errors.New("FLEXIBLE_PAYLOAD_EMPTY")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		return nil, fmt.Errorf("FLEXIBLE_PAYLOAD_JSON_INVALID: %w", err)
	}
	if object == nil {
		return nil, errors.New("FLEXIBLE_PAYLOAD_NOT_JSON_OBJECT")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("FLEXIBLE_PAYLOAD_TRAILING_JSON")
		}
		return nil, fmt.Errorf("FLEXIBLE_PAYLOAD_TRAILING_CONTENT: %w", err)
	}
	return object, nil
}

// ToBytes returns the original byte sequence when the payload originated as
// bytes. Map-origin payloads are encoded as JSON and encoding errors are surfaced.
func (p FlexiblePayload) ToBytes() ([]byte, error) {
	if !p.isMap {
		return append([]byte(nil), p.raw...), nil
	}
	raw, err := json.Marshal(p.object)
	if err != nil {
		return nil, fmt.Errorf("FLEXIBLE_PAYLOAD_MAP_ENCODE_FAILED: %w", err)
	}
	return raw, nil
}

// IsEmpty reports whether the selected representation has no payload data.
func (p FlexiblePayload) IsEmpty() bool {
	if p.isMap {
		return len(p.object) == 0
	}
	return len(p.raw) == 0
}
