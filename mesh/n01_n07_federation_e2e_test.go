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

func TestN07FederatesToRealN04N05N06NativeCapabilities(t *testing.T) {
	secret := requiredEnv(t, "SOUL_MESH_HMAC_SECRET")
	_ = secret
	requiredEnv(t, "SOUL_MESH_N04_URL")
	requiredEnv(t, "SOUL_MESH_N05_URL")
	requiredEnv(t, "SOUL_MESH_N06_URL")

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

			description, err := client.Discover(ctx, tc.nucleus)
			if err != nil {
				t.Fatalf("REAL_MESH_DISCOVERY_FAILED nucleus=%s: %v", tc.nucleus, err)
			}
			if !supportsExecutableCapability(description, tc.capability) {
				t.Fatalf("REAL_MESH_CAPABILITY_NOT_EXECUTABLE nucleus=%s capability=%s description=%#v", tc.nucleus, tc.capability, description)
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
