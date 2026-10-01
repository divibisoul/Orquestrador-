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

func requiredFederationEnv(t *testing.T, key string) string {
	t.Helper()
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		t.Fatalf("REAL_MESH_E2E_BLOCKED: required environment variable %s is not configured", key)
	}
	return value
}

func getenv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func TestN07FederatesToRealN04N05N06NativeCapabilities(t *testing.T) {
	requiredFederationEnv(t, "SOUL_MESH_HMAC_SECRET")
	nuclei := []string{"N04", "N05", "N06"}
	for _, nucleus := range nuclei {
		requiredFederationEnv(t, "SOUL_MESH_"+nucleus+"_URL")
	}

	client, err := NewPeerClient(&http.Client{Timeout: 15 * time.Second})
	if err != nil {
		t.Fatalf("peer client initialization failed: %v", err)
	}

	for _, nucleus := range nuclei {
		nucleus := nucleus
		t.Run(nucleus, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			describeCorrelation := fmt.Sprintf("real-n07-%s-describe", strings.ToLower(nucleus))
			description, err := client.CallWithCorrelation(
				ctx,
				nucleus,
				"mesh.describe",
				map[string]any{"from": "N07", "purpose": "real-federation-verification"},
				describeCorrelation,
			)
			if err != nil {
				t.Fatalf(
					"REAL_MESH_DISCOVERY_FAILED nucleus=%s correlation=%s: %v",
					nucleus, describeCorrelation, err,
				)
			}

			if got, _ := description["correlationId"].(string); got != describeCorrelation {
				t.Fatalf(
					"REAL_MESH_CORRELATION_MISMATCH nucleus=%s expected=%s got=%q",
					nucleus, describeCorrelation, got,
				)
			}
			if got, _ := description["source"].(string); got != nucleus {
				t.Fatalf("REAL_MESH_SOURCE_MISMATCH nucleus=%s got=%q", nucleus, got)
			}
			if got, _ := description["target"].(string); got != "N07" {
				t.Fatalf("REAL_MESH_TARGET_MISMATCH nucleus=%s got=%q", nucleus, got)
			}
			if got, _ := description["kind"].(string); got != "response" {
				t.Fatalf("REAL_MESH_KIND_MISMATCH nucleus=%s got=%q", nucleus, got)
			}

			pingCorrelation := fmt.Sprintf("real-n07-%s-ping", strings.ToLower(nucleus))
			ping, err := client.CallWithCorrelation(
				ctx,
				nucleus,
				"mesh.ping",
				map[string]any{"from": "N07", "correlation": pingCorrelation},
				pingCorrelation,
			)
			if err != nil {
				t.Fatalf(
					"REAL_MESH_EXECUTION_FAILED nucleus=%s capability=mesh.ping correlation=%s: %v",
					nucleus, pingCorrelation, err,
				)
			}

			if got, _ := ping["correlationId"].(string); got != pingCorrelation {
				t.Fatalf(
					"REAL_MESH_CORRELATION_MISMATCH nucleus=%s expected=%s got=%q",
					nucleus, pingCorrelation, got,
				)
			}
			if got, _ := ping["source"].(string); got != nucleus {
				t.Fatalf("REAL_MESH_SOURCE_MISMATCH nucleus=%s got=%q", nucleus, got)
			}
			if got, _ := ping["target"].(string); got != "N07" {
				t.Fatalf("REAL_MESH_TARGET_MISMATCH nucleus=%s got=%q", nucleus, got)
			}
		})
	}
}
