package orchestrator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

func TestMultiAgentEvidenceFailClosed(t *testing.T) {
	t.Setenv("SOUL_N07_CREWAI_ENABLED", "")
	e := providerEvidence("crewai")
	if e.State != "DEGRADED" || e.Code != "MULTIAGENT_ADAPTER_DISABLED" {
		t.Fatalf("%#v", e)
	}
}

func TestCorrelationPreserved(t *testing.T) {
	m := protocol.Message{CorrelationID: "corr-test"}
	r, _ := multiAgentFail(m, "TEST")
	if r.CorrelationID != "corr-test" {
		t.Fatal("correlation lost")
	}
}

func TestMultiAgentRequestFromMetadata(t *testing.T) {
	m := protocol.Message{
		Metadata: map[string]string{
			"provider":  "crewai",
			"goal":      "test-goal",
			"roles":     `["builder"]`,
			"maxRounds": "4",
		},
	}
	q := multiAgentRequestFromMessage(m)
	if q.Provider != "crewai" || q.Goal != "test-goal" || len(q.Roles) != 1 || q.Roles[0] != "builder" || q.MaxRounds != 4 {
		t.Fatalf("unexpected request: %#v", q)
	}
}

func TestMetaGPTRoutingUsesRegisteredExternalOwner(t *testing.T) {
	t.Setenv("SOUL_N07_METAGPT_ENABLED", "")
	e := providerEvidence("metagpt")
	if e.State != "DEGRADED" || e.Code != "MULTIAGENT_ADAPTER_DISABLED" {
		t.Fatalf("unexpected MetaGPT evidence: %#v", e)
	}
	if e.Provider != "metagpt" || len(e.Revision) != 40 {
		t.Fatalf("MetaGPT must be resolved from the external owner registry: %#v", e)
	}
	if !strings.Contains(filepath.ToSlash(e.Root), "integrations/external/metagpt") {
		t.Fatalf("unexpected canonical MetaGPT root: %s", e.Root)
	}
}
