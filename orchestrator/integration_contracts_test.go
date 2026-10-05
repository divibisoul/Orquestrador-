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
