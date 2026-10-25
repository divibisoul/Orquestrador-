package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newExternalTestEngine(t *testing.T) (*Engine, ExternalAdapterRegistry) {
	t.Helper()
	n, _ := neural.New(4, .05)
	c, _ := prefrontal.New(.1, 8)
	g := supergpu.New(nil)
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewExternalAdapterRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterExternalAdapterOperations(e, r); err != nil {
		t.Fatal(err)
	}
	return e, r
}
func TestExternalAdapterManifestHasTwentyTwoProviders(t *testing.T) {
	_, r := newExternalTestEngine(t)
	if len(r.byID) != 22 {
		t.Fatalf("adapter-backed providers=%d want=22", len(r.byID))
	}
	if len(r.expansionProviders) != 10 {
		t.Fatalf("projected expansions=%d want=10", len(r.expansionProviders))
	}
}
func TestExternalAdapterPrimarySixteenRemainIntact(t *testing.T) {
	_, r := newExternalTestEngine(t)
	primary := []string{
		"superpowers","superagi","langgraph","crewai","microsoft-agent-framework","openhands",
		"metagpt","agentscope","letta-code","browser-use","smolagents","pydantic-ai",
		"llama-index","dspy","whisper","kokoro",
	}
	for _, id := range primary {
		if _, ok := r.byID[id]; !ok {
			t.Fatalf("original primary provider missing: %s", id)
		}
	}
}
func TestExternalAdapterExecutionFailsClosedByDefault(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	old := os.Getenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH")
	t.Cleanup(func() { _ = os.Setenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH", old) })
	_ = os.Unsetenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH")
	_, err := e.Execute(context.Background(), "external.langgraph.execute@1.0.0", nil, map[string]string{"external_operation": "workflow.invoke", "input": "x"})
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_ADAPTER_DISABLED") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSourceWorktreeIsGitRejectsParentRepositoryFallback(t *testing.T) {
	// A normal directory below the N07 checkout is not a provider worktree.
	// Git may nevertheless find the parent repository when invoked with -C.
	root, err := os.MkdirTemp(repositoryRoot(), ".external-unmaterialized-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })

	if sourceWorktreeIsGit(root) {
		t.Fatalf("unmaterialized provider directory was mistaken for a Git worktree: %s", root)
	}
}

func TestExternalAdapterProbeRegistered(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	if _, err := e.Execute(context.Background(), "external.langgraph.probe@1.0.0", nil, nil); err != nil {
		t.Fatalf("probe failed: %v", err)
	}
}

func TestExternalAdapterProbeDoesNotClaimRuntimeWithoutSourceWorktree(t *testing.T) {
	e, r := newExternalTestEngine(t)
	result, err := e.Execute(context.Background(), "external.langgraph.probe@1.0.0", nil, nil)
	if err != nil {
		t.Fatalf("metadata probe failed: %v", err)
	}
	var details map[string]any
	if err := json.Unmarshal([]byte(result.Metadata["external_result_json"]), &details); err != nil {
		t.Fatalf("probe response is not valid JSON: %v", err)
	}
	if details["execution_proven"] != false {
		t.Fatalf("probe must never claim source execution: %#v", details)
	}
	root := filepath.Join(repositoryRoot(), r.byID["langgraph"].Root)
	if !sourceWorktreeIsGit(root) {
		if details["state"] != "PROJECTED" || details["source_present"] != false || result.Status != "degraded" {
			t.Fatalf("missing source must remain PROJECTED/degraded, not PASS: result=%#v details=%#v", result, details)
		}
	}
}

func TestExternalAdapterAllTwentyTwoProvidersProbe(t *testing.T) {
	e, r := newExternalTestEngine(t)
	for id := range r.byID {
		operation := "external." + id + ".probe@1.0.0"
		if _, err := e.Execute(context.Background(), operation, nil, nil); err != nil {
			t.Fatalf("provider %s probe failed: %v", id, err)
		}
	}
}

func TestProjectedExpansionsAreMetadataOnly(t *testing.T) {
	e, r := newExternalTestEngine(t)
	if len(r.expansionProviders) != 10 {
		t.Fatalf("projected expansions=%d want=10", len(r.expansionProviders))
	}
	for id, p := range r.expansionProviders {
		if p.State != "PROJECTED" || p.ExecutionEnabled || p.ExecutionProven {
			t.Fatalf("expansion %s was prematurely promoted: %+v", id, p)
		}
		if len(p.RuntimeRouteOperations) != 0 {
			t.Fatalf("expansion %s has runtime routes: %v", id, p.RuntimeRouteOperations)
		}
		if !containsString(p.MetadataOperations, "external."+id+".describe@1.0.0") ||
			!containsString(p.MetadataOperations, "external."+id+".probe@1.0.0") {
			t.Fatalf("expansion %s lacks its metadata-only endpoints", id)
		}

		result, err := e.Execute(context.Background(), "external."+id+".probe@1.0.0", nil, nil)
		if err != nil {
			t.Fatalf("projected expansion %s metadata probe failed: %v", id, err)
		}
		var details map[string]any
		if err := json.Unmarshal([]byte(result.Metadata["external_result_json"]), &details); err != nil {
			t.Fatalf("projected expansion %s returned invalid JSON: %v", id, err)
		}
		if details["executionProven"] != false || details["executionEnabled"] != false {
			t.Fatalf("projected expansion %s claimed execution proof: %#v", id, details)
		}
		state, _ := details["state"].(string)
		if state != "PROJECTED" && state != "BLOCKED" {
			t.Fatalf("projected expansion %s returned unexpected state %q", id, state)
		}
		if _, err := e.Execute(context.Background(), "external."+id+".execute@1.0.0", nil, nil); err == nil {
			t.Fatalf("projected expansion %s unexpectedly exposes an execute route", id)
		}

		// Even a caller-supplied activation flag cannot bypass the adapter registry.
		envKey := "SOUL_EXTERNAL_EXECUTE_" + strings.ToUpper(strings.ReplaceAll(id, "-", "_"))
		oldValue, wasSet := os.LookupEnv(envKey)
		if err := os.Setenv(envKey, "true"); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if wasSet {
				_ = os.Setenv(envKey, oldValue)
			} else {
				_ = os.Unsetenv(envKey)
			}
		})
		_, err = e.Execute(context.Background(), ExternalFederationExecuteOperation, nil, map[string]string{
			"provider": id,
			"external_operation": firstNonEmptySlice(p.LegacyOperations),
		})
		if err == nil || !strings.Contains(err.Error(), "EXTERNAL_PROVIDER_UNKNOWN") {
			t.Fatalf("projected expansion %s bypassed core adapter registry with activation flag: %v", id, err)
		}
	}
}

func TestExternalAdapterDegradedExecutionFailsClosed(t *testing.T) {
	m := protocol.Message{CorrelationID: "corr-degraded"}
	result, err := externalAdapterExecutionResult(m, map[string]any{
		"state": "DEGRADED",
		"provider": "langgraph",
		"code": "TEST_DEGRADED",
	})
	if err == nil || !strings.Contains(err.Error(), "TEST_DEGRADED") {
		t.Fatalf("degraded execution unexpectedly succeeded: result=%#v err=%v", result, err)
	}
	if result.Status != "error" || result.Error != "TEST_DEGRADED" {
		t.Fatalf("degraded execution was not fail-closed: %#v", result)
	}
}

func TestExternalAdapterTransportAuthorityCannotBeOverridden(t *testing.T) {
	req := buildExternalAdapterRequest(
		"langgraph",
		"execute",
		"workflow.invoke",
		"/app/integrations/external/langgraph",
		map[string]string{
			"provider":  "crewai",
			"operation": "arbitrary.operation",
			"root":      "/tmp/escape",
			"mode":      "probe",
		},
	)
	for key, expected := range map[string]string{
		"provider":  "langgraph",
		"mode":      "execute",
		"operation": "workflow.invoke",
		"root":      "/app/integrations/external/langgraph",
	} {
		if got := req[key]; got != expected {
			t.Fatalf("%s override succeeded: got=%v want=%s", key, got, expected)
		}
	}
}

func TestExternalComplementaryDescribeContractsAreRegistered(t *testing.T) {
	e,r:=newExternalTestEngine(t)
	for _,id:=range []string{"autogenesis","octos","hora-graph-core","mycelium","prime-agent","cuda-oxide"} {
		result,err:=e.Execute(context.Background(),"external."+id+".describe@1.0.0",nil,nil)
		if err!=nil {
			t.Fatalf("complementary describe failed for %s: %v",id,err)
		}
		root:=filepath.Join(repositoryRoot(),r.byID[id].Root)
		if !sourceWorktreeIsGit(root) && result.Status!="degraded" {
			t.Fatalf("unmaterialized complementary source %s must remain degraded, got %q",id,result.Status)
		}
	}
}

func firstNonEmptySlice(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "source.verify"
}

