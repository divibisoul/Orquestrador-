package mesh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

const ClareiraCapability = "clareira.ingest"

type PeerCaller interface {
	CallWithCorrelation(context.Context, string, string, map[string]any, string) (map[string]any, error)
}

type ClareiraReporter struct {
	peers PeerCaller
}

func NewClareiraReporter(peers PeerCaller) (*ClareiraReporter, error) {
	if peers == nil {
		return nil, errors.New("Clareira reporter requires N07 Mesh peer client")
	}
	return &ClareiraReporter{peers: peers}, nil
}

func (r *ClareiraReporter) Report(ctx context.Context, event supergpu.ExecutionEvent) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	correlationID := strings.TrimSpace(event.CorrelationID)
	if correlationID == "" {
		correlationID = protocol.NewTraceID()
	}

	raw, err := json.Marshal(map[string]any{
		"phase":          event.Phase,
		"operation":      event.Operation,
		"device_id":     event.DeviceID,
		"backend":        event.Backend,
		"input_size":    event.InputSize,
		"output_size":   event.OutputSize,
		"correlation_id": correlationID,
		"error":         event.Error,
	})
	if err != nil {
		return fmt.Errorf("clareira event encode: %w", err)
	}

	packet := map[string]any{
		"id":                 protocol.NewTraceID(),
		"data":               string(raw),
		"informationalValue": float64(event.OutputSize),
		"criticality":        criticalityForPhase(event.Phase),
		"packetType":         "StateReport",
		"sourceId":           "N07.SuperGPU",
		"timestamp":          time.Now().UnixMilli(),
		"correlationId":      correlationID,
		"metadata": map[string]any{
			"component": "supergpu",
			"phase":     event.Phase,
			"operation": event.Operation,
			"device":    event.DeviceID,
			"backend":   event.Backend,
		},
	}

	_, err = r.peers.CallWithCorrelation(ctx, "N01", ClareiraCapability, map[string]any{"packet": packet}, correlationID)
	return err
}

func criticalityForPhase(phase string) float64 {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "failed":
		return 1
	case "started":
		return 0.6
	default:
		return 0.4
	}
}

type CapabilityExecutionEvent struct {
	Phase         string
	Operation     string
	Provider      string
	Model         string
	Source        string
	Owner         string
	CorrelationID string
	InputSize     int
	OutputSize    int
	Error         string
}

func (r *ClareiraReporter) ReportCapability(ctx context.Context, event CapabilityExecutionEvent) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	correlationID := strings.TrimSpace(event.CorrelationID)
	if correlationID == "" {
		correlationID = protocol.NewTraceID()
	}
	raw, err := json.Marshal(map[string]any{
		"phase":          event.Phase,
		"operation":      event.Operation,
		"provider":       event.Provider,
		"model":          event.Model,
		"source":         event.Source,
		"owner":          event.Owner,
		"input_size":     event.InputSize,
		"output_size":    event.OutputSize,
		"correlation_id": correlationID,
		"error":          event.Error,
	})
	if err != nil {
		return fmt.Errorf("clareira capability event encode: %w", err)
	}
	packet := map[string]any{
		"id":                 protocol.NewTraceID(),
		"data":               string(raw),
		"informationalValue": float64(event.OutputSize),
		"criticality":        criticalityForCapabilityPhase(event.Phase),
		"packetType":         "StateReport",
		"sourceId":           "N07.Orchestrator",
		"timestamp":          time.Now().UnixMilli(),
		"correlationId":      correlationID,
		"metadata": map[string]any{
			"component": "gemini",
			"operation": event.Operation,
			"provider":   event.Provider,
			"model":      event.Model,
			"owner":      event.Owner,
			"source":     event.Source,
			"phase":      event.Phase,
		},
	}
	_, err = r.peers.CallWithCorrelation(ctx, "N01", ClareiraCapability, map[string]any{"packet": packet}, correlationID)
	return err
}

func criticalityForCapabilityPhase(phase string) float64 {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "failed":
		return 1
	case "started":
		return 0.6
	default:
		return 0.4
	}
}
