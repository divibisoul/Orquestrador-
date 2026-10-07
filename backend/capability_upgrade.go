package backend

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

type capabilityUpgradeRequest struct {
	Component  string `json:"component"`
	Capability string `json:"capability"`
}

func (s *Server) capabilityUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "GET or POST required"})
		return
	}

	component := strings.TrimSpace(r.URL.Query().Get("component"))
	capability := strings.TrimSpace(r.URL.Query().Get("capability"))
	if r.Method == http.MethodPost {
		var req capabilityUpgradeRequest
		if err := decodeJSON(r, s.Config.MaxRequestBytes, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		if component == "" {
			component = strings.TrimSpace(req.Component)
		}
		if capability == "" {
			capability = strings.TrimSpace(req.Capability)
		}
	}

	plan, err := orchestrator.ResolveCapabilityUpgrade(r.Context(), component, capability)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status":                  "BLOCKED",
			"error":                   err.Error(),
			"component":               component,
			"capability":              capability,
			"no_fake_runtime_success": true,
		})
		return
	}
	raw, _ := json.Marshal(plan)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":             "PROJECTED",
		"plan":               json.RawMessage(raw),
		"runtime_activation": "requires explicit provider adapter, configuration and real verification evidence",
	})
}
