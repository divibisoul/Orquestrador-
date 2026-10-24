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


func TestDirectAffinityHostSetsAreConsistentAcrossSOULManifests(t *testing.T) {
	type providerContract struct {
		ID      string   `json:"id"`
		Hosts   []string `json:"directAdapterHosts"`
		Targets []string `json:"targets"`
	}
	type sourceRecord struct {
		ID      string   `json:"id"`
		Targets []string `json:"targets"`
	}
	type adapterRecord struct {
		ID    string `json:"id"`
		Owner string `json:"owner"`
	}
	type augmentationRecord struct {
		ID                string   `json:"id"`
		DirectAffinity    bool     `json:"direct_affinity"`
		PreferredHost     string   `json:"direct_adapter_host"`
		DirectAdapterHosts []string `json:"direct_adapter_hosts"`
	}

	contractsRaw, err := os.ReadFile("../integrations/soul-provider-contracts.json")
	if err != nil { t.Fatal(err) }
	var contracts struct { Providers []providerContract `json:"providers"` }
	if err := json.Unmarshal(contractsRaw, &contracts); err != nil { t.Fatal(err) }

	registryRaw, err := os.ReadFile("../integrations/external-capabilities.json")
	if err != nil { t.Fatal(err) }
	var registry struct { Repositories []sourceRecord `json:"repositories"` }
	if err := json.Unmarshal(registryRaw, &registry); err != nil { t.Fatal(err) }

	adaptersRaw, err := os.ReadFile("../integrations/external-adapters.json")
	if err != nil { t.Fatal(err) }
	var adapters struct { Providers []adapterRecord `json:"providers"` }
	if err := json.Unmarshal(adaptersRaw, &adapters); err != nil { t.Fatal(err) }

	augmentationRaw, err := os.ReadFile("../integrations/soul-25-augmentation.json")
	if err != nil { t.Fatal(err) }
	var augmentation struct { Providers []augmentationRecord `json:"external_providers"` }
	if err := json.Unmarshal(augmentationRaw, &augmentation); err != nil { t.Fatal(err) }

	fabricRaw, err := os.ReadFile("../integrations/soul-25-capability-fabric.json")
	if err != nil { t.Fatal(err) }
	var fabric struct {
		Edges []struct { From string `json:"from"`; To string `json:"to"`; Mode string `json:"mode"` } `json:"edges"`
	}
	if err := json.Unmarshal(fabricRaw, &fabric); err != nil { t.Fatal(err) }

	contractByID := make(map[string]providerContract)
	for _, p := range contracts.Providers { contractByID[p.ID] = p }
	targetsByID := make(map[string][]string)
	for _, p := range registry.Repositories { targetsByID[p.ID] = p.Targets }
	ownerByID := make(map[string]string)
	for _, p := range adapters.Providers { ownerByID[p.ID] = p.Owner }
	graphHosts := make(map[string][]string)
	for _, edge := range fabric.Edges {
		if edge.Mode == "direct-affinity" { graphHosts[edge.To] = append(graphHosts[edge.To], edge.From) }
	}
	eqSet := func(left, right []string) bool {
		if len(left) != len(right) { return false }
		counts := make(map[string]int, len(left)+len(right))
		for _, v := range left { counts[v]++ }
		for _, v := range right { counts[v]-- }
		for _, count := range counts { if count != 0 { return false } }
		return true
	}
	if len(augmentation.Providers) != 16 {
		t.Fatalf("augmentation providers=%d want=16", len(augmentation.Providers))
	}
	for _, p := range augmentation.Providers {
		hosts := p.DirectAdapterHosts
		if len(hosts) == 0 { t.Errorf("%s: declared direct_adapter_hosts is empty", p.ID); continue }
		if !contains(hosts, p.PreferredHost) { t.Errorf("%s: legacy preferred host %q not in hosts %v", p.ID, p.PreferredHost, hosts) }
		if p.DirectAffinity != contains(hosts, "N07") { t.Errorf("%s: direct_affinity disagrees with N07 host membership", p.ID) }
		if !eqSet(hosts, targetsByID[p.ID]) { t.Errorf("%s: registry targets %v != declared hosts %v", p.ID, targetsByID[p.ID], hosts) }
		if !eqSet(hosts, graphHosts[p.ID]) { t.Errorf("%s: graph affinity edges %v != declared hosts %v", p.ID, graphHosts[p.ID], hosts) }
		contract, ok := contractByID[p.ID]
		if !ok { t.Errorf("%s: provider contract missing", p.ID); continue }
		if !eqSet(hosts, contract.Hosts) || !eqSet(hosts, contract.Targets) { t.Errorf("%s: provider contract hosts/targets mismatch: %+v", p.ID, contract) }
		owner, ok := ownerByID[p.ID]
		if !ok { t.Errorf("%s: adapter registration missing", p.ID); continue }
		if !contains(hosts, owner) { t.Errorf("%s: primary adapter owner %s is outside declared affinity hosts %v", p.ID, owner, hosts) }
	}
}
