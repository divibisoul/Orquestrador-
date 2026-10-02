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
			ID         string `json:"id"`
			Intent     string `json:"intent"`
			Boundary   string `json:"activation_boundary"`
			State      string `json:"adapter_state"`
			FailClosed bool   `json:"fail_closed"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(providerData, &providers); err != nil {
		t.Fatal(err)
	}
	if len(providers.Providers) != 17 {
		t.Fatalf("providers=%d want=17", len(providers.Providers))
	}
	for _, p := range providers.Providers {
		if p.ID == "" || p.Intent == "" || p.Boundary == "" || p.State == "" || !p.FailClosed {
			t.Fatalf("incomplete provider contract: %+v", p)
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
