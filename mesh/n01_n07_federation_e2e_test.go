//go:build integration

package mesh

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func requiredEnv(t *testing.T, key string) string {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		t.Fatalf("REAL_MESH_E2E_BLOCKED: required environment variable %s is not configured", key)
	}
	return value
}

func advertisedCapability(description map[string]any, capability string) bool {
	capability = strings.SplitN(strings.TrimSpace(capability), "@", 2)[0]
	for _, key := range []string{"executableCapabilities", "declaredCapabilities", "capabilities"} {
		raw, ok := description[key]
		if !ok {
			if nested, ok := description["payload"].(map[string]any); ok {
				raw = nested[key]
			}
		}
		switch items := raw.(type) {
		case []any:
			for _, item := range items {
				switch value := item.(type) {
				case string:
					if strings.SplitN(strings.TrimSpace(value), "@", 2)[0] == capability {
						return true
					}
				case map[string]any:
					if id, _ := value["id"].(string); strings.TrimSpace(id) == capability {
						return true
					}
				}
			}
		case []string:
			for _, value := range items {
				if strings.SplitN(strings.TrimSpace(value), "@", 2)[0] == capability {
					return true
				}
			}
		}
	}
	return false
}

func TestN07FederatesToRealN04N05N06NativeCapabilities(t *testing.T) {
	requiredEnv(t, "SOUL_MESH_HMAC_SECRET")
	for _, nucleus := range []string{"N04", "N05", "N06"} {
		requiredEnv(t, "SOUL_MESH_"+nucleus+"_URL")
	}

	client, err := NewPeerClient(&http.Client{Timeout: 15 * time.Second})
	if err != nil {
		t.Fatalf("peer client initialization failed: %v", err)
	}

	cases := []struct {
		nucleus    string
		capability string
		payload    map[string]any
		validate   func(t *testing.T, payload map[string]any)
	}{
		{
			nucleus:    "N04",
			capability: "core.health",
			payload:    map[string]any{"probe": "phase1", "requestedBy": "N07"},
		},
		{
			nucleus:    "N05",
			capability: "core.health",
			payload:    map[string]any{"probe": "phase1", "requestedBy": "N07"},
		},
		{
			nucleus:    "N06",
			capability: "support.mesh",
			payload:    map[string]any{"probe": "phase1", "requestedBy": "N07"},
			validate: func(t *testing.T, payload map[string]any) {
				t.Helper()
				accepted, ok := payload["accepted"].(bool)
				if !ok || !accepted {
					t.Fatalf("N06 support.mesh did not return native acceptance: %#v", payload)
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%s_%s", tc.nucleus, strings.ReplaceAll(tc.capability, ".", "_")), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			describeCorrelation := fmt.Sprintf("phase1-n07-%s-describe", strings.ToLower(tc.nucleus))
			description, err := client.CallWithCorrelation(ctx, tc.nucleus, "mesh.describe", map[string]any{"from": "N07"}, describeCorrelation)
			if err != nil {
				t.Fatalf("REAL_MESH_DISCOVERY_FAILED nucleus=%s correlation=%s: %v", tc.nucleus, describeCorrelation, err)
			}
			if !advertisedCapability(description, tc.capability) {
				t.Fatalf("REAL_MESH_CAPABILITY_NOT_ADVERTISED nucleus=%s capability=%s description=%#v", tc.nucleus, tc.capability, description)
			}

			correlation := fmt.Sprintf("phase1-n07-%s", strings.ToLower(tc.nucleus))
			result, err := client.CallWithCorrelation(ctx, tc.nucleus, tc.capability, tc.payload, correlation)
			if err != nil {
				t.Fatalf("REAL_MESH_EXECUTION_FAILED nucleus=%s capability=%s correlation=%s: %v", tc.nucleus, tc.capability, correlation, err)
			}
			if got, _ := result["correlationId"].(string); got != correlation {
				t.Fatalf("REAL_MESH_CORRELATION_MISMATCH nucleus=%s expected=%s got=%q", tc.nucleus, correlation, got)
			}
			if got, _ := result["source"].(string); got != tc.nucleus {
				t.Fatalf("REAL_MESH_SOURCE_MISMATCH nucleus=%s got=%q", tc.nucleus, got)
			}
			if got, _ := result["target"].(string); got != "N07" {
				t.Fatalf("REAL_MESH_TARGET_MISMATCH nucleus=%s got=%q", tc.nucleus, got)
			}

			payload, ok := result["payload"].(map[string]any)
			if !ok {
				t.Fatalf("REAL_MESH_PAYLOAD_MISSING nucleus=%s result=%#v", tc.nucleus, result)
			}
			if tc.validate != nil {
				tc.validate(t, payload)
			}
		})
	}
}
func TestN07PersistsRGOStageToRealN01HortaCore(t *testing.T) {
	requiredEnv(t, "SOUL_MESH_HMAC_SECRET")
	requiredEnv(t, "SOUL_MESH_N01_URL")

	client, err := NewPeerClient(&http.Client{Timeout: 15 * time.Second})
	if err != nil {
		t.Fatalf("peer client initialization failed: %v", err)
	}

	correlation := "e2e-n07-n01-rgo-horta"
	stage := map[string]any{
		"sequence_index":           7,
		"stage":                    "ERU",
		"scale":                    "MACRO",
		"finding_id":               "e2e-rgo-horta-finding",
		"cycle_id":                 correlation,
		"parent_stage":             "TRINITY::VALIDATION",
		"parent_hash":              "sha256:parent-e2e",
		"input_hash":               "sha256:input-e2e",
		"output_hash":              "sha256:output-e2e",
		"status":                   "FINALIZED",
		"eru_snapshot_hash":        "sha256:eru-e2e",
		"rgo_evidence_chain_hash":  "sha256:rgo-e2e",
		"data": map[string]any{
			"source": "N07",
			"proof":  "real-soul-mesh-e2e",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	result, err := client.CallWithCorrelation(
		ctx,
		"N01",
		"rgo.hortacore.store",
		stage,
		correlation,
	)
	if err != nil {
		t.Fatalf("REAL_RGO_HORTA_E2E_FAILED correlation=%s: %v", correlation, err)
	}

	if got, _ := result["correlationId"].(string); got != correlation {
		t.Fatalf("REAL_RGO_HORTA_CORRELATION_MISMATCH expected=%s got=%q", correlation, got)
	}
	if got, _ := result["source"].(string); got != "N01" {
		t.Fatalf("REAL_RGO_HORTA_SOURCE_MISMATCH got=%q", got)
	}
	if got, _ := result["target"].(string); got != "N07" {
		t.Fatalf("REAL_RGO_HORTA_TARGET_MISMATCH got=%q", got)
	}

	payload, ok := result["payload"].(map[string]any)
	if !ok {
		t.Fatalf("REAL_RGO_HORTA_PAYLOAD_MISSING result=%#v", result)
	}
	if persisted, _ := payload["persisted"].(bool); !persisted {
		t.Fatalf("REAL_RGO_HORTA_NOT_PERSISTED payload=%#v", payload)
	}
	if key, _ := payload["key"].(string); !strings.Contains(key, "e2e-rgo-horta-finding") ||
		!strings.Contains(key, correlation) ||
		!strings.Contains(key, "sha256:output-e2e") {
		t.Fatalf("REAL_RGO_HORTA_KEY_INTEGRITY_FAILED key=%q", key)
	}
}
