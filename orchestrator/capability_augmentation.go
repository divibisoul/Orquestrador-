package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	CapabilityAugmentationResolveOperation  = "soul.capability.augment.resolve@1.0.0"
	CapabilityAugmentationDescribeOperation = "soul.capability.augment.describe@1.0.0"
)

type capabilityAugmentationProvider struct {
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities"`
	State        string   `json:"state"`
}

type capabilityAugmentationComponent struct {
	Repository         string   `json:"repository"`
	NativeCapabilities []string `json:"native_capabilities"`
	Augmenters         []string `json:"augmenters"`
}

type capabilityAugmentationManifest struct {
	NativeComponents map[string]capabilityAugmentationComponent `json:"native_components"`
	Providers        map[string]capabilityAugmentationProvider  `json:"providers"`
	RuntimePolicy    map[string]any                             `json:"runtime_policy"`
}

type CapabilityAugmentationPlan struct {
	Component          string   `json:"component"`
	NativeCapability   string   `json:"native_capability,omitempty"`
	NativeOwner        string   `json:"native_owner"`
	Augmenters         []string `json:"augmenters"`
	AvailableProviders []string `json:"available_providers"`
	BlockedProviders   []string `json:"blocked_providers"`
	ExecutionPolicy    string   `json:"execution_policy"`
	EvidenceState      string   `json:"evidence_state"`
}

func loadCapabilityAugmentationManifest() (capabilityAugmentationManifest, error) {
	candidates := []string{}
	if p := strings.TrimSpace(os.Getenv("SOUL_NATIVE_AUGMENTATION_PATH")); p != "" {
		candidates = append(candidates, p)
	}
	if cwd, err := os.Getwd(); err == nil {
		for dir := filepath.Clean(cwd); ; dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, "integrations", "native-capability-augmentation.json"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	var last error
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			last = err
			continue
		}
		var m capabilityAugmentationManifest
		if err := json.Unmarshal(data, &m); err != nil {
			last = err
			continue
		}
		if len(m.NativeComponents) != 9 || len(m.Providers) != 25 {
			last = errors.New("native augmentation manifest must contain 9 SOUL components and 25 providers")
			continue
		}
		return m, nil
	}
	if last == nil {
		last = errors.New("native augmentation manifest not found")
	}
	return capabilityAugmentationManifest{}, last
}

func ResolveCapabilityAugmentation(ctx context.Context, component, capability string) (CapabilityAugmentationPlan, error) {
	if ctx == nil {
		return CapabilityAugmentationPlan{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return CapabilityAugmentationPlan{}, err
	}
	component = strings.ToUpper(strings.TrimSpace(component))
	capability = strings.TrimSpace(capability)

	manifest, err := loadCapabilityAugmentationManifest()
	if err != nil {
		return CapabilityAugmentationPlan{}, err
	}

	c, ok := manifest.NativeComponents[component]
	if !ok {
		return CapabilityAugmentationPlan{}, errors.New("unknown SOUL component: " + component)
	}

	plan := CapabilityAugmentationPlan{
		Component:          component,
		NativeOwner:        component,
		Augmenters:         []string{},
		AvailableProviders: []string{},
		BlockedProviders:   []string{},
		ExecutionPolicy:    "native owner executes first; external provider requires explicit adapter/configuration/evidence",
		EvidenceState:      "PROJECTED",
	}

	for _, providerID := range c.Augmenters {
		p, exists := manifest.Providers[providerID]
		if !exists {
			plan.BlockedProviders = append(plan.BlockedProviders, providerID)
			continue
		}
		plan.Augmenters = append(plan.Augmenters, providerID)
		if strings.EqualFold(p.State, "ADAPTER_BOUND") {
			plan.AvailableProviders = append(plan.AvailableProviders, providerID)
		} else {
			plan.BlockedProviders = append(plan.BlockedProviders, providerID)
		}
	}

	if capability != "" {
		foundNative := false
		for _, c := range c.NativeCapabilities {
			if c == capability {
				foundNative = true
				break
			}
		}
		if !foundNative {
			plan.EvidenceState = "BLOCKED"
			return plan, errors.New("native capability is not registered: " + capability)
		}
		plan.NativeCapability = capability
	}

	return plan, nil
}

func RegisterCapabilityAugmentationOperations(e *Engine) error {
	if e == nil {
		return errors.New("engine is nil")
	}
	if err := e.Register(CapabilityAugmentationDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		manifest, err := loadCapabilityAugmentationManifest()
		if err != nil {
			return protocol.Result{}, err
		}
		raw, err := json.Marshal(map[string]any{
			"components": len(manifest.NativeComponents),
			"providers":  len(manifest.Providers),
			"operation":  CapabilityAugmentationResolveOperation,
			"policy":     manifest.RuntimePolicy,
		})
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.capability-augmentation", Target: m.Source, Status: "ok", Metadata: map[string]string{"description_json": string(raw)}}, nil
	}); err != nil {
		return err
	}

	return e.Register(CapabilityAugmentationResolveOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		plan, err := ResolveCapabilityAugmentation(ctx, m.Metadata["component"], m.Metadata["capability"])
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.capability-augmentation", Target: m.Source, Status: "error", Error: err.Error(), Metadata: map[string]string{"evidence_state": plan.EvidenceState}}, err
		}
		raw, err := json.Marshal(plan)
		if err != nil {
			return protocol.Result{}, err
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.capability-augmentation", Target: m.Source, Status: "ok", Metadata: map[string]string{"plan_json": string(raw)}}, nil
	})
}
