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
func TestExternalExpansionProvidersRemainProjectedAndNonExecutable(t *testing.T) {
	_, r := newExternalTestEngine(t)
	if len(r.expansionProviders) != 9 {
		t.Fatalf("projected expansion providers=%d, want 9", len(r.expansionProviders))
	}
	expected := map[string]bool{
		"fedml": true, "hivemind": true, "temporal": true,
		"cognitive-workspace": true, "ravana": true, "ray": true,
		"bijux-dag-runtime": true, "ouro-loop": true, "recurs": true,
	}
	for _, p := range r.expansionProviders {
		if !expected[p.ID] {
			t.Fatalf("unexpected projected expansion provider %q", p.ID)
		}
		if p.State != "PROJECTED" || p.IntegrationState == "" || p.RuntimeRequirement == "" {
			t.Fatalf("expansion evidence boundary missing for %s: %#v", p.ID, p)
		}
		if _, routeable := r.byID[p.ID]; routeable {
			t.Fatalf("projected provider %s became runtime-routeable", p.ID)
		}
		for _, operation := range p.Operations {
			if r.operationAllowed(p.ID, "execute", operation) {
				t.Fatalf("projected operation unexpectedly executable: %s:%s", p.ID, operation)
			}
		}
		delete(expected, p.ID)
	}
	if len(expected) != 0 {
		t.Fatalf("projected expansion sources missing: %v", expected)
	}
}

func TestExternalFederationDescribeIncludesProjectedSourcesWithoutPromotingThem(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	result, err := e.Execute(context.Background(), ExternalFederationDescribeOperation, nil, nil)
	if err != nil {
		t.Fatalf("federation describe failed: %v", err)
	}
	raw := result.Metadata["federation_json"]
	var report struct {
		State string `json:"state"`
		CatalogState string `json:"catalog_state"`
		RuntimeEvidence string `json:"runtime_evidence"`
		ProviderCount int `json:"provider_count"`
		ExpansionProviderCount int `json:"expansion_provider_count"`
		DeclaredExternalRepositoryCount int `json:"declared_external_repository_count"`
		ExpansionProviders []struct {
			ID string `json:"id"`
			State string `json:"state"`
			ExecutionEnabled bool `json:"executionEnabled"`
		} `json:"expansion_providers"`
	}
	if err := json.Unmarshal([]byte(raw), &report); err != nil {
		t.Fatalf("federation descriptor JSON invalid: %v; raw=%s", err, raw)
	}
	if report.State != "PASS" || report.CatalogState != "DECLARED_WITH_PROJECTED_EXPANSIONS" || report.RuntimeEvidence != "UNVERIFIED" {
		t.Fatalf("catalog/runtime evidence states conflated: %#v", report)
	}
	if report.ProviderCount != 22 || report.ExpansionProviderCount != 9 || report.DeclaredExternalRepositoryCount != 31 {
		t.Fatalf("unexpected federated inventory counts: %#v", report)
	}
	if len(report.ExpansionProviders) != 9 {
		t.Fatalf("expansion inventory length=%d", len(report.ExpansionProviders))
	}
	for _, p := range report.ExpansionProviders {
		if p.State != "PROJECTED" || p.ExecutionEnabled {
			t.Fatalf("projected source was promoted in descriptor: %#v", p)
		}
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


func TestInspectExternalSourceUsesBuildAttestationOnlyForNonGitRuntimeRoot(t *testing.T) {
	repoRoot := t.TempDir()
	sourceRoot := filepath.Join(repoRoot, "integrations", "external", "langgraph")
	if err := os.MkdirAll(sourceRoot, 0700); err != nil {
		t.Fatal(err)
	}
	revision := strings.Repeat("a", 40)
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("verified runtime source"), 0600); err != nil {
		t.Fatal(err)
	}
	attestation := ExternalRuntimeAttestation{
		SchemaVersion: "soul.external.runtime-attestation.v1",
		GeneratedFromMaterializedGit: true,
		Repositories: map[string]string{"langgraph": revision},
	}
	raw, err := json.Marshal(attestation)
	if err != nil {
		t.Fatal(err)
	}
	attestationPath := filepath.Join(repoRoot, "config", "soul-external-runtime-attestation.json")
	if err := os.MkdirAll(filepath.Dir(attestationPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attestationPath, raw, 0600); err != nil {
		t.Fatal(err)
	}

	matched := inspectExternalSourceWithRepositoryRoot(sourceRoot, revision, repoRoot)
	if !matched.Present || !matched.RevisionMatched || matched.State != "PIN_MATCH" || matched.ObservedRevision != revision {
		t.Fatalf("valid build attestation was not accepted in runtime root: %#v", matched)
	}
	mismatch := inspectExternalSourceWithRepositoryRoot(sourceRoot, strings.Repeat("b", 40), repoRoot)
	if !mismatch.Present || mismatch.RevisionMatched || mismatch.State != "REVISION_MISMATCH" || mismatch.ObservedRevision != revision {
		t.Fatalf("attested revision mismatch was not preserved: %#v", mismatch)
	}

	emptyRoot := filepath.Join(repoRoot, "integrations", "external", "empty-provider")
	if err := os.MkdirAll(emptyRoot, 0700); err != nil {
		t.Fatal(err)
	}
	attestation.Repositories["empty-provider"] = revision
	raw, err = json.Marshal(attestation)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attestationPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	empty := inspectExternalSourceWithRepositoryRoot(emptyRoot, revision, repoRoot)
	if empty.Present || empty.RevisionMatched || empty.State != "MISSING_WORKTREE" {
		t.Fatalf("attestation overrode an empty source directory: %#v", empty)
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
