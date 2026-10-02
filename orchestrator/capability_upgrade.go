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
	SOULCapabilityUpgradeDescribeOperation = "soul.capability.upgrade.describe@1.0.0"
	SOULCapabilityUpgradeResolveOperation  = "soul.capability.upgrade.resolve@1.0.0"
)

type CapabilityUpgradeProvider struct {
	ID           string   `json:"id"`
	Source       string   `json:"source"`
	Revision     string   `json:"revision"`
	Capabilities []string `json:"capabilities"`
	Targets      []string `json:"targets"`
	AdapterState string   `json:"adapter_state"`
	Evidence     string   `json:"evidence_state"`
}

type CapabilityUpgradePlan struct {
	Component             string                     `json:"component"`
	RequestedCapability   string                     `json:"requested_capability,omitempty"`
	NativeAuthority       string                     `json:"native_authority"`
	DirectAugmenters      []CapabilityUpgradeProvider `json:"direct_augmenters"`
	FederatedSources      []CapabilityUpgradeProvider `json:"federated_sources"`
	ExecutionBoundary     string                     `json:"execution_boundary"`
	RequiresExplicitAdapter bool                     `json:"requires_explicit_adapter"`
	NoFakeRuntimeSuccess  bool                       `json:"no_fake_runtime_success"`
}

type externalProviderManifest struct {
	Repositories []struct {
		ID           string   `json:"id"`
		Source       string   `json:"source"`
		Revision     string   `json:"revision"`
		Targets      []string `json:"targets"`
		Capabilities []string `json:"capabilities"`
	} `json:"repositories"`
}

var soulComponentAuthorities = map[string]string{
	"N01": "N01",
	"N02": "N02",
	"N03": "N03",
	"N04": "N04",
	"N05": "N05",
	"N06": "N06",
	"N07": "N07",
	"SARA": "SARA",
	"JEV": "JEV",
}

func loadExternalProviderManifest() (externalProviderManifest, error) {
	candidates := []string{}
	if p := strings.TrimSpace(os.Getenv("SOUL_EXTERNAL_CAPABILITY_REGISTRY_PATH")); p != "" {
		candidates = append(candidates, p)
	}
	if cwd, err := os.Getwd(); err == nil {
		for dir := filepath.Clean(cwd); ; dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, "integrations", "external-capabilities.json"))
			parent := filepath.Dir(dir)
			if parent == dir { break }
		}
	}
	var last error
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil { last = err; continue }
		var manifest externalProviderManifest
		if err := json.Unmarshal(data, &manifest); err != nil { last = err; continue }
		if len(manifest.Repositories) != 16 { last = errors.New("expected exactly 16 external providers"); continue }
		return manifest, nil
	}
	if last == nil { last = errors.New("external capability registry not found") }
	return externalProviderManifest{}, last
}

func ResolveCapabilityUpgrade(ctx context.Context, component, capability string) (CapabilityUpgradePlan, error) {
	if ctx == nil { return CapabilityUpgradePlan{}, errors.New("context is nil") }
	if err := ctx.Err(); err != nil { return CapabilityUpgradePlan{}, err }
	component = strings.ToUpper(strings.TrimSpace(component))
	capability = strings.TrimSpace(capability)
	if _, ok := soulComponentAuthorities[component]; !ok { return CapabilityUpgradePlan{}, errors.New("unknown SOUL component: " + component) }

	manifest, err := loadExternalProviderManifest()
	if err != nil { return CapabilityUpgradePlan{}, err }

	plan := CapabilityUpgradePlan{
		Component: component,
		RequestedCapability: capability,
		NativeAuthority: soulComponentAuthorities[component],
		DirectAugmenters: make([]CapabilityUpgradeProvider, 0),
		FederatedSources: make([]CapabilityUpgradeProvider, 0, len(manifest.Repositories)),
		ExecutionBoundary: "canonical soul-mesh/1 via N07; provider runtime requires explicit adapter/configuration",
		RequiresExplicitAdapter: true,
		NoFakeRuntimeSuccess: true,
	}
	for _, p := range manifest.Repositories {
		state := "PROJECTED"
		adapterState := "CATALOGED"
		if p.ID == "superpowers" {
			adapterState = "IMPLEMENTED_AT_N07"
		}
		entry := CapabilityUpgradeProvider{ID:p.ID, Source:p.Source, Revision:p.Revision, Capabilities:p.Capabilities, Targets:p.Targets, AdapterState:adapterState, Evidence:state}
		plan.FederatedSources = append(plan.FederatedSources, entry)

		direct := false
		for _, target := range p.Targets {
			if strings.EqualFold(target, component) { direct = true; break }
		}
		if direct {
			if capability == "" || providerSupportsCapability(p, capability) {
				plan.DirectAugmenters = append(plan.DirectAugmenters, entry)
			}
		}
	}
	return plan, nil
}

func providerSupportsCapability(p struct{ ID string `json:"id"`; Source string `json:"source"`; Revision string `json:"revision"`; Targets []string `json:"targets"`; Capabilities []string `json:"capabilities"` }, capability string) bool {
	query := strings.ToLower(strings.TrimSpace(capability))
	for _, candidate := range p.Capabilities {
		c := strings.ToLower(strings.TrimSpace(candidate))
		if c == query || strings.Contains(c, query) || strings.Contains(query, c) { return true }
	}
	return false
}

func RegisterCapabilityUpgradeOperations(e *Engine) error {
	if e == nil { return errors.New("engine is nil") }
	if err := e.Register(SOULCapabilityUpgradeDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil { return protocol.Result{}, err }
		plan, err := ResolveCapabilityUpgrade(ctx, "N07", "")
		if err != nil { return capabilityUpgradeFailure(m, err) }
		raw, err := json.Marshal(map[string]any{"operation":SOULCapabilityUpgradeDescribeOperation,"provider_count":len(plan.FederatedSources),"components":[]string{"N01","N02","N03","N04","N05","N06","N07","SARA","JEV"},"rule":"all 16 upstreams remain discoverable; direct affinity controls normal routing"})
		if err != nil { return capabilityUpgradeFailure(m, err) }
		return protocol.Result{TraceID:m.TraceID,CorrelationID:m.CorrelationID,Source:"N07.capability-upgrade",Target:m.Source,Status:"ok",Metadata:map[string]string{"description_json":string(raw)}}, nil
	}); err != nil { return err }

	return e.Register(SOULCapabilityUpgradeResolveOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		component := strings.TrimSpace(m.Metadata["component"])
		capability := strings.TrimSpace(m.Metadata["capability"])
		plan, err := ResolveCapabilityUpgrade(ctx, component, capability)
		if err != nil { return capabilityUpgradeFailure(m, err) }
		raw, err := json.Marshal(plan)
		if err != nil { return capabilityUpgradeFailure(m, err) }
		return protocol.Result{TraceID:m.TraceID,CorrelationID:m.CorrelationID,Source:"N07.capability-upgrade",Target:m.Source,Status:"ok",Metadata:map[string]string{"plan_json":string(raw)}}, nil
	})
}

func capabilityUpgradeFailure(m protocol.Message, err error) (protocol.Result, error) {
	return protocol.Result{TraceID:m.TraceID,CorrelationID:m.CorrelationID,Source:"N07.capability-upgrade",Target:m.Source,Status:"error",Error:"CAPABILITY_UPGRADE_RESOLUTION_FAILED:"+err.Error()}, err
}
