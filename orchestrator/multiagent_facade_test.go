package orchestrator

import (
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

func TestMetaGPTRoutingRemainsN06Owned(t *testing.T) {
	m := protocol.Message{
		CorrelationID: "corr-metagpt",
		Metadata: map[string]string{
			"provider": "metagpt",
			"goal":     "must not execute in N07",
		},
	}
	r, err := multiAgentFail(m, "MULTIAGENT_NATIVE_OWNER_N06")
	if err == nil || r.Status != "error" || r.Error != "MULTIAGENT_NATIVE_OWNER_N06" {
		t.Fatalf("unexpected fail-closed result: %#v err=%v", r, err)
	}
	if r.CorrelationID != m.CorrelationID {
		t.Fatal("correlation lost during authority boundary")
	}
}
