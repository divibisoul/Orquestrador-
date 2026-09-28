package rgo

import (
	"testing"
	"time"
)

func TestFromBugShieldPreservesDetectionAndDoesNotInventCorrection(t *testing.T) {
	payload := map[string]any{
		"schema_version":"1.2.0","scan_id":"scan-1","timestamp":time.Now().UTC().Format(time.RFC3339Nano),
		"scanner":map[string]any{"name":"BugShield","version":"1.2.0"},
		"scope":map[string]any{"input_hash":"sha256:input"},
		"finding":map[string]any{
			"id":"f-1","type":"BUG","category":"static","description":"finding",
			"epistemic_mode":"INSPECTION","verification_state":"UNVERIFIED",
			"evidence":[]any{map[string]any{"id":"ev-1","kind":"static","ref":"lint://1"}},
		},
	}
	env, err := FromBugShield(payload)
	if err != nil { t.Fatal(err) }
	if env.Dual.Status != DualUnresolved { t.Fatalf("dual was invented: %#v", env.Dual) }
	if env.Extensions["bugshield"] == nil { t.Fatal("source payload not preserved") }
	if err := env.Validate(); err != nil { t.Fatal(err) }
}
