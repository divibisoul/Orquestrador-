package orchestrator

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSOULIntegrationContracts(t *testing.T) {
	providerData, err := os.ReadFile("../integrations/soul-provider-contracts.json")
	if err != nil {
		t.Fatal(err)
	}
	var providers struct {
		TargetSemantics string `json:"target_semantics"`
		Providers []struct {
			ID           string   `json:"id"`
			Intent       string   `json:"intent"`
			Boundary     string   `json:"activation_boundary"`
			State        string   `json:"adapter_state"`
			Verification string   `json:"verification_state"`
			Hosts        []string `json:"directAdapterHosts"`
			Targets      []string `json:"targets"`
			FailClosed   bool     `json:"fail_closed"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(providerData, &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers.Providers) != 17 {
		t.Fatalf("providers=%d want=17", len(providers.Providers))
	}
	if providers.TargetSemantics == "" {
		t.Fatal("provider target semantics are not documented")
	}

	augmentationData, err := os.ReadFile("../integrations/soul-25-augmentation.json")
	if err != nil {
		t.Fatal(err)
	}
	var augmentation struct {
		Providers []struct {
			ID                  string   `json:"id"`
			DirectAffinity      bool     `json:"direct_affinity"`
			DirectAdapterHost   string   `json:"direct_adapter_host"`
			DirectAdapterHosts  []string `json:"direct_adapter_hosts"`
		} `json:"external_providers"`
	}
	if err := json.Unmarshal(augmentationData, &augmentation); err != nil {
		t.Fatal(err)
	}
	affinityByID := make(map[string]struct {
		DirectAffinity bool
		PreferredHost  string
		Hosts          []string
	}, len(augmentation.Providers))
	for _, p := range augmentation.Providers {
		affinityByID[p.ID] = struct {
			DirectAffinity bool
			PreferredHost  string
			Hosts          []string
		}{p.DirectAffinity, p.DirectAdapterHost, p.DirectAdapterHosts}
	}
	if len(affinityByID) != 16 {
		t.Fatalf("primary upstreams in augmentation=%d want=16", len(affinityByID))
	}

	registryData, err := os.ReadFile("../integrations/external-capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Repositories []struct {
			ID      string   `json:"id"`
			Targets []string `json:"targets"`
		} `json:"repositories"`
	}
	if err := json.Unmarshal(registryData, &registry); err != nil {
		t.Fatal(err)
	}
	registryTargets := make(map[string][]string, len(registry.Repositories))
	for _, p := range registry.Repositories {
		registryTargets[p.ID] = p.Targets
	}

	fabricData, err := os.ReadFile("../integrations/soul-25-capability-fabric.json")
	if err != nil {
		t.Fatal(err)
	}
	var fabric struct {
		Edges []struct {
			From string `json:"from"`
			To   string `json:"to"`
			Mode string `json:"mode"`
		} `json:"edges"`
	}
	if err := json.Unmarshal(fabricData, &fabric); err != nil {
		t.Fatal(err)
	}
	graphHosts := make(map[string][]string)
	for _, edge := range fabric.Edges {
		if edge.Mode == "direct-affinity" {
			graphHosts[edge.To] = append(graphHosts[edge.To], edge.From)
		}
	}

	for _, p := range providers.Providers {
		if p.ID == "" || p.Intent == "" || p.Boundary == "" || p.State == "" || p.Verification == "" || len(p.Hosts) == 0 || len(p.Targets) == 0 || !p.FailClosed {
			t.Fatalf("incomplete provider contract: %+v", p)
		}
		switch p.State {
		case "CATALOGED", "ADAPTER_BOUND":
			if p.Verification == "REAL" {
				t.Fatalf("provider adapter boundary cannot claim REAL verification without runtime evidence: %+v", p)
			}
		case "IMPLEMENTED_AT_N07":
			if !contains(p.Hosts, "N07") {
				t.Fatalf("N07 implementation without N07 adapter host: %+v", p)
			}
		default:
			t.Fatalf("unknown adapter state: %q", p.State)
		}
		if p.Verification == "REAL" && p.State == "CATALOGED" {
			t.Fatalf("provider execution overclaim: %+v", p)
		}

		affinity, hasAugmentation := affinityByID[p.ID]
		if !hasAugmentation {
			// agent-arsenal is a native N07 bridge, not one of the 16 public
			// SOUL-25 upstreams, so it has no external-provider affinity row.
			continue
		}
		hosts := affinity.Hosts
		if len(hosts) == 0 {
			t.Fatalf("direct_adapter_hosts missing: %s", p.ID)
		}
		if !contains(hosts, affinity.PreferredHost) {
			t.Fatalf("legacy preferred host is not in the complete host set: %s:%s", p.ID, affinity.PreferredHost)
		}
		if affinity.DirectAffinity != contains(hosts, "N07") {
			t.Fatalf("direct_affinity/N07 host mismatch: %s direct_affinity=%t hosts=%v", p.ID, affinity.DirectAffinity, hosts)
		}
		if !sameStringSet(p.Hosts, hosts) || !sameStringSet(p.Targets, hosts) {
			t.Fatalf("provider contract host mismatch: %s contract_hosts=%v targets=%v augmentation_hosts=%v", p.ID, p.Hosts, p.Targets, hosts)
		}
		if !sameStringSet(registryTargets[p.ID], hosts) {
			t.Fatalf("source registry host mismatch: %s registry=%v expected=%v", p.ID, registryTargets[p.ID], hosts)
		}
		if !sameStringSet(graphHosts[p.ID], hosts) {
			t.Fatalf("SOUL-25 graph direct-affinity mismatch: %s graph=%v expected=%v", p.ID, graphHosts[p.ID], hosts)
		}
	}

	componentData, err := os.ReadFile("../integrations/soul-component-contracts.json")
	if err != nil {
		t.Fatal(err)
	}
	var components struct {
		Components []struct {
			ID string `json:"id"`
		} `json:"components"`
	}
	if err := json.Unmarshal(componentData, &components); err != nil {
		t.Fatal(err)
	}
	if len(components.Components) != 9 {
		t.Fatalf("components=%d want=9", len(components.Components))
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left)+len(right))
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}
