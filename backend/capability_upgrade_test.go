package backend

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

func TestCapabilityUpgradeRejectsUnknownComponent(t *testing.T) {
	s := &Server{Engine: nil, Config: Config{MaxRequestBytes: 1 << 20}}
	req := httptest.NewRequest("GET", "/v1/capability-upgrade?component=UNKNOWN", nil)
	req.Header.Set("Authorization", "Bearer test")
	rec := httptest.NewRecorder()

	s.capabilityUpgrade(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status=%d want=400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "BLOCKED") {
		t.Fatalf("response=%s", rec.Body.String())
	}
}

func TestCapabilityUpgradeEndpointUsesResolverContract(t *testing.T) {
	if _, err := orchestrator.ResolveCapabilityUpgrade(nil, "N07", ""); err == nil {
		t.Fatal("nil context must be rejected")
	}
}
