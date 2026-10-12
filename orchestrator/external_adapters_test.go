package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
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
		t.Fatalf("providers=%d", len(r.byID))
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
func TestExternalAdapterProbeRegistered(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	if _, err := e.Execute(context.Background(), "external.langgraph.probe@1.0.0", nil, nil); err != nil {
		t.Fatalf("probe failed: %v", err)
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


func TestExternalFederationExecutionRequiresCanonicalNativeOwner(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	t.Setenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH", "")

	base := map[string]string{
		"provider":           "langgraph",
		"external_operation": "workflow.invoke",
	}
	_, err := e.Execute(context.Background(), ExternalFederationExecuteOperation, nil, base)
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_NATIVE_OWNER_REQUIRED") {
		t.Fatalf("missing native owner did not fail closed: %v", err)
	}

	wrongOwner := map[string]string{
		"provider":           "langgraph",
		"external_operation": "workflow.invoke",
		"native_owner":       "N01",
	}
	_, err = e.Execute(context.Background(), ExternalFederationExecuteOperation, nil, wrongOwner)
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_NATIVE_OWNER_MISMATCH") {
		t.Fatalf("wrong native owner did not fail closed: %v", err)
	}

	wrongRevision := map[string]string{
		"provider":           "langgraph",
		"external_operation": "workflow.invoke",
		"native_owner":       "N07",
		"provider_revision":  strings.Repeat("0", 40),
	}
	_, err = e.Execute(context.Background(), ExternalFederationExecuteOperation, nil, wrongRevision)
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_PROVIDER_REVISION_MISMATCH") {
		t.Fatalf("provider revision drift did not fail closed: %v", err)
	}

	canonicalOwner := map[string]string{
		"provider":           "langgraph",
		"external_operation": "workflow.invoke",
		"native_owner":       "N07",
		"provider_revision":  "2d942085e214ef6b99b6f54ed4d544a7c7c5ac56",
	}
	_, err = e.Execute(context.Background(), ExternalFederationExecuteOperation, nil, canonicalOwner)
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_ADAPTER_DISABLED") {
		t.Fatalf("canonical owner and revision must still respect the default-disabled adapter gate: %v", err)
	}
}

func TestExternalFederationDescribeIsAvailableWithoutExecutingProviders(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	result, err := e.Execute(context.Background(), ExternalFederationDescribeOperation, nil, nil)
	if err != nil {
		t.Fatalf("federation describe should not require provider execution: %v", err)
	}
	if result.Status != "ok" || result.Metadata["federation_json"] == "" {
		t.Fatalf("missing canonical federation observation: %#v", result)
	}
	var observation map[string]any
	if err := json.Unmarshal([]byte(result.Metadata["federation_json"]), &observation); err != nil {
		t.Fatalf("invalid federation description JSON: %v", err)
	}
	if observation["control_plane"] != "N07" {
		t.Fatalf("wrong control plane: %#v", observation["control_plane"])
	}
	if observation["provider_count"] != float64(22) {
		t.Fatalf("unexpected canonical adapter registry count: %#v", observation["provider_count"])
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
	e,_:=newExternalTestEngine(t)
	for _,id:=range []string{"autogenesis","octos","hora-graph-core","mycelium","prime-agent","cuda-oxide"} {
		if _,err:=e.Execute(context.Background(),"external."+id+".describe@1.0.0",nil,nil);err!=nil {
			t.Fatalf("complementary describe failed for %s: %v",id,err)
		}
	}
}


func TestInspectExternalSourceRequiresOwnGitWorktreeAndPinnedRevision(t *testing.T) {
	root := t.TempDir()
	runGit := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	if output, err := exec.Command("git", "init", root).CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, output)
	}
	runGit("config", "user.name", "SOUL source status test")
	runGit("config", "user.email", "soul-source-status@example.invalid")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("committed source\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "fixture repository for Git-worktree status")
	revision := runGit("rev-parse", "HEAD")

	matched := inspectExternalSource(root, revision)
	if !matched.Present || !matched.RevisionMatched || matched.State != "PIN_MATCH" || matched.ObservedRevision != revision {
		t.Fatalf("matching source worktree misclassified: %#v", matched)
	}
	mismatch := inspectExternalSource(root, strings.Repeat("0", 40))
	if !mismatch.Present || mismatch.RevisionMatched || mismatch.State != "REVISION_MISMATCH" {
		t.Fatalf("revision mismatch misclassified: %#v", mismatch)
	}

	emptySubdirectory := filepath.Join(root, "uninitialized-submodule")
	if err := os.Mkdir(emptySubdirectory, 0700); err != nil {
		t.Fatal(err)
	}
	missing := inspectExternalSource(emptySubdirectory, revision)
	if missing.Present || missing.RevisionMatched || missing.State != "MISSING_WORKTREE" || missing.ObservedRevision != "" {
		t.Fatalf("empty subdirectory was misclassified as a source checkout: %#v", missing)
	}
	if inspectExternalSource(filepath.Join(root, "does-not-exist"), revision).Present {
		t.Fatal("nonexistent external source reported present")
	}
}
