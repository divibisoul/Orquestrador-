package learning

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type ExternalFabric struct {
	machine *Machine
	peers   *mesh.PeerClient
	sara    *backend.SARAProxy
}

func NewExternalFabric(machine *Machine, peers *mesh.PeerClient, sara *backend.SARAProxy) *ExternalFabric {
	return &ExternalFabric{machine: machine, peers: peers, sara: sara}
}

func (f *ExternalFabric) command(provider string) string {
	switch strings.TrimSpace(provider) {
	case "FedML":
		return strings.TrimSpace(os.Getenv("SOUL_FEDML_COMMAND"))
	case "Hivemind":
		return strings.TrimSpace(os.Getenv("SOUL_HIVEMIND_COMMAND"))
	default:
		return ""
	}
}

func (f *ExternalFabric) Execute(ctx context.Context, provider, capability string, payload map[string]any, correlation string) (map[string]any, error) {
	if f == nil {
		return nil, errors.New("external learning fabric is unavailable")
	}
	provider = strings.TrimSpace(provider)
	capability = strings.TrimSpace(capability)
	correlation = strings.TrimSpace(correlation)
	command := f.command(provider)
	if provider == "" || capability == "" || correlation == "" {
		return nil, errors.New("learning provider, capability and correlation are required")
	}
	if command == "" {
		return nil, errors.New(provider + "_COMMAND_NOT_CONFIGURED")
	}

	request := map[string]any{
		"provider":       provider,
		"capability":     capability,
		"payload":        payload,
		"correlation_id": correlation,
	}
	raw, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.Stdin = bytes.NewReader(raw)
	output, err := cmd.CombinedOutput()
	outputText := strings.TrimSpace(string(output))

	result := map[string]any{
		"provider":         provider,
		"capability":       capability,
		"correlation_id":   correlation,
		"provider_output":  outputText,
		"provider_sha256":  sha256Hex([]byte(outputText)),
		"created_at_unixms": time.Now().UnixMilli(),
	}
	if err != nil {
		result["state"] = "BLOCKED"
		_ = f.observe(ctx, provider, capability, correlation, -1, "provider_execution_failed")
		return result, fmt.Errorf("%s provider execution: %w", provider, err)
	}

	if err := json.Unmarshal([]byte(outputText), &map[string]any{}); err == nil && strings.HasPrefix(outputText, "{") {
		var structured map[string]any
		if json.Unmarshal([]byte(outputText), &structured) == nil {
			result["provider_result"] = structured
		}
	}
	result["state"] = "REAL"
	result["persistence_boundary"] = "N01.HortaCore"
	result["strategy_boundary"] = "N07.Prefrontal"
	result["transport_boundary"] = "SARA.NervoVago"

	if f.peers == nil {
		return map[string]any{"state": "BLOCKED", "error": "N01_HORTACORE_PEER_UNCONFIGURED"}, errors.New("N01_HORTACORE_PEER_UNCONFIGURED")
	}
	if _, err := f.peers.CallWithCorrelation(ctx, "N01", "rgo.hortacore.store", result, correlation); err != nil {
		_ = f.observe(ctx, provider, capability, correlation, -1, "hortacore_persistence_failed")
		return map[string]any{"state": "BLOCKED", "error": err.Error()}, err
	}
	result["hortacore_persisted"] = true

	if f.sara == nil || !f.sara.Configured() {
		return map[string]any{"state": "BLOCKED", "error": "SARA_NERVOVAGO_UNCONFIGURED"}, errors.New("SARA_NERVOVAGO_UNCONFIGURED")
	}
	if _, err := f.sara.PublishVagus(ctx, map[string]any{
		"vagus_version":  "1.0",
		"message_id":     protocol.NewTraceID(),
		"correlation_id": correlation,
		"source":         "N07.ExternalLearningFabric",
		"target":         "NervoVago",
		"priority":       90,
		"ttl":            5000,
		"type":           capability,
		"payload":        result,
	}, correlation); err != nil {
		_ = f.observe(ctx, provider, capability, correlation, -1, "nervovago_publish_failed")
		return map[string]any{"state": "BLOCKED", "error": err.Error()}, err
	}
	result["nervovago_published"] = true

	if err := f.observe(ctx, provider, capability, correlation, 1, "provider_execution_success"); err != nil {
		return map[string]any{"state": "BLOCKED", "error": err.Error()}, err
	}
	return result, nil
}

func (f *ExternalFabric) observe(ctx context.Context, provider, capability, correlation string, reward float64, outcome string) error {
	if f.machine == nil {
		return errors.New("N07_Prefrontal_learning_machine_unconfigured")
	}
	return f.machine.Feedback(ctx, Experience{
		ID:            correlation + ":" + provider + ":" + capability,
		TraceID:       correlation,
		CorrelationID: correlation,
		Source:        "N07",
		Target:        provider,
		Capability:    capability,
		Outcome:       outcome,
		Reward:        reward,
		Confidence:    1,
		Provenance:    "external-learning-fabric",
		Metadata: map[string]string{
			"provider": provider,
			"transport_boundary": "SARA.NervoVago",
			"persistence_boundary": "N01.HortaCore",
			"strategy_boundary": "N07.Prefrontal",
		},
	})
}

func RegisterExternalFabricOperations(e *orchestrator.Engine, fabric *ExternalFabric) error {
	if e == nil || fabric == nil {
		return errors.New("external learning operations require engine and fabric")
	}
	register := func(provider, capability string) error {
		return e.Register(capability, func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			payload := map[string]any{
				"values": message.Payload,
				"metadata": message.Metadata,
			}
			out, err := fabric.Execute(ctx, provider, capability, payload, message.CorrelationID)
			raw, marshalErr := json.Marshal(out)
			if marshalErr != nil {
				return protocol.Result{}, marshalErr
			}
			status := "ok"
			if err != nil {
				status = "blocked"
			}
			return protocol.Result{
				TraceID: message.TraceID,
				CorrelationID: message.CorrelationID,
				Source: "N07.learning",
				Target: message.Source,
				Status: status,
				Metadata: map[string]string{
					"learning_provider": provider,
					"learning_state": providerState(out),
					"learning_evidence": string(raw),
				},
			}, err
		})
	}
	if err := register("FedML", "learning.federated.execute@1.0.0"); err != nil {
		return err
	}
	return register("Hivemind", "learning.decentralized.execute@1.0.0")
}

func providerState(value map[string]any) string {
	if value == nil {
		return "BLOCKED"
	}
	if state, ok := value["state"].(string); ok && strings.TrimSpace(state) != "" {
		return state
	}
	return "BLOCKED"
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
