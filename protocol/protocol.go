package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const Version = "N07.v1"
const SoulMeshContractVersion = "1.1.0"

type Message struct {
	Version         string            `json:"version"`
	ContractVersion string            `json:"contract_version"`
	TraceID         string            `json:"trace_id"`
	CorrelationID   string            `json:"correlation_id"`
	Source          string            `json:"source"`
	Target          string            `json:"target"`
	Kind            string            `json:"kind"`
	Operation       string            `json:"operation"`
	Payload         []float64         `json:"payload,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Priority        int               `json:"priority"`
	Deadline        time.Time         `json:"deadline,omitempty"`
	Sequence        uint64            `json:"sequence"`
}
type Result struct {
	TraceID       string            `json:"trace_id"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Source        string            `json:"source"`
	Target        string            `json:"target"`
	Status        string            `json:"status"`
	Payload       []float64         `json:"payload,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Error         string            `json:"error,omitempty"`
}

func NewTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("n07-%d", time.Now().UnixNano())
	}
	return "n07-" + hex.EncodeToString(b)
}
func NewMessage(source, target, kind, operation string, payload []float64) Message {
	return Message{Version: Version, ContractVersion: SoulMeshContractVersion, TraceID: NewTraceID(), CorrelationID: NewTraceID(), Source: source, Target: target, Kind: kind, Operation: operation, Payload: append([]float64(nil), payload...), Metadata: map[string]string{}, Priority: 0, Sequence: uint64(time.Now().UnixNano())}
}
func (m Message) Validate() error {
	if m.Version != Version {
		return errors.New("unsupported protocol version")
	}
	if m.ContractVersion != SoulMeshContractVersion {
		return errors.New("unsupported Mesh contract version")
	}
	if strings.TrimSpace(m.TraceID) == "" {
		return errors.New("trace_id is required")
	}
	if strings.TrimSpace(m.CorrelationID) == "" {
		return errors.New("correlation_id is required")
	}
	if strings.TrimSpace(m.Source) == "" || strings.TrimSpace(m.Target) == "" {
		return errors.New("source and target are required")
	}
	if strings.TrimSpace(m.Kind) == "" || strings.TrimSpace(m.Operation) == "" {
		return errors.New("kind and operation are required")
	}
	if m.Priority < 0 || m.Priority > 100 {
		return errors.New("priority must be between 0 and 100")
	}
	if !m.Deadline.IsZero() && time.Now().After(m.Deadline) {
		return contextDeadlineError{}
	}
	for _, v := range m.Payload {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("payload contains non-finite number")
		}
	}
	return nil
}
func Encode(m Message) ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}
func Decode(data []byte) (Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return Message{}, err
	}
	if err := m.Validate(); err != nil {
		return Message{}, err
	}
	return m, nil
}
func Propagate(parent Message, source, target, operation string, payload []float64) Message {
	m := parent
	m.Source, m.Target, m.Operation = source, target, operation
	m.Payload = append([]float64(nil), payload...)
	m.Sequence++
	return m
}

type contextDeadlineError struct{}

func (contextDeadlineError) Error() string { return "message deadline exceeded" }
