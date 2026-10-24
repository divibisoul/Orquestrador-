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
			ExecutionProven bool `json:"executionProven"`
			MeshMetadataOperations []string `json:"meshMetadataOperations"`
			RuntimeRouteOperations []string `json:"runtimeRouteOperations"`
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
		if p.State != "PROJECTED" || p.ExecutionEnabled || p.ExecutionProven {
			t.Fatalf("projected source was promoted in descriptor: %#v", p)
		}
		if len(p.MeshMetadataOperations) != 2 || len(p.RuntimeRouteOperations) != 0 {
			t.Fatalf("projected source metadata/execution route split invalid: %#v", p)
		}
	}
}

func TestProjectedExpansionSourcesExposeMeshMetadataWithoutExecuteRoutes(t *testing.T) {
	e, r := newExternalTestEngine(t)
	if len(r.expansionProviders) != 9 {
		t.Fatalf("projected expansion providers=%d, want 9", len(r.expansionProviders))
	}
	for _, p := range r.expansionProviders {
		for _, kind := range []string{"describe", "probe"} {
			operation := "external." + p.ID + "." + kind + "@1.0.0"
			result, err := e.Execute(context.Background(), operation, nil, nil)
			if err != nil {
				t.Fatalf("Mesh metadata route %s failed to execute: %v", operation, err)
			}
			raw := result.Metadata["external_result_json"]
			var evidence struct {
				State            string `json:"state"`
				Provider          string `json:"provider"`
				ExecutionEnabled  bool   `json:"executionEnabled"`
				ExecutionProven   bool   `json:"executionProven"`
				RuntimeEvidence   string `json:"runtimeEvidence"`
				RuntimeRouteOperations []string `json:"runtimeRouteOperations"`
				RevisionMatched   bool   `json:"revisionMatched"`
				SourcePresent     bool   `json:"sourcePresent"`
			}
			if err := json.Unmarshal([]byte(raw), &evidence); err != nil {
				t.Fatalf("Mesh metadata route %s returned invalid evidence: %v; raw=%s", operation, err, raw)
			}
			if evidence.Provider != p.ID {
				t.Fatalf("Mesh metadata route %s returned provider=%q", operation, evidence.Provider)
			}
			if evidence.State != "PROJECTED" && evidence.State != "BLOCKED" {
				t.Fatalf("projected source %s reported an invalid evidence state: %#v", p.ID, evidence)
			}
			if evidence.ExecutionEnabled || evidence.ExecutionProven || evidence.RuntimeEvidence != "UNVERIFIED" || len(evidence.RuntimeRouteOperations) != 0 {
				t.Fatalf("projected source %s was promoted by a metadata route: %#v", p.ID, evidence)
			}
			if evidence.State == "PROJECTED" && (!evidence.SourcePresent || !evidence.RevisionMatched || result.Status != "degraded") {
				t.Fatalf("projected source %s lacks exact-pin evidence or degraded status: result=%#v evidence=%#v", p.ID, result, evidence)
			}
			if evidence.State == "BLOCKED" && result.Status != "error" {
				t.Fatalf("blocked source %s did not fail closed: result=%#v evidence=%#v", p.ID, result, evidence)
			}
		}
		result, err := e.Execute(context.Background(), "external."+p.ID+".execute@1.0.0", nil, map[string]string{"external_operation": "execute"})
		if err == nil && result.Status != "error" {
			t.Fatalf("projected source %s unexpectedly exposes an execute route: %#v", p.ID, result)
		}
	}
}

func TestCanonicalFederationSourceMetadataContractCoversAllThirtyOneSources(t *testing.T) {
	e, r := newExternalTestEngine(t)
	ids := make([]string, 0, len(r.byID)+len(r.expansionProviders))
	for id := range r.byID {
		ids = append(ids, id)
	}
	for _, p := range r.expansionProviders {
		ids = append(ids, p.ID)
	}
	if len(ids) != 31 || len(r.revisions) != 31 {
		t.Fatalf("canonical external inventory mismatch: ids=%d revisions=%d", len(ids), len(r.revisions))
	}

	for _, id := range ids {
		result, err := e.Execute(context.Background(), ExternalFederationSourceDescribeOperation, nil, map[string]string{"provider": id})
		if err != nil {
			t.Fatalf("canonical source describe failed for %s: %v", id, err)
		}
		raw := result.Metadata["external_result_json"]
		var evidence struct {
			State                  string   `json:"state"`
			Provider               string   `json:"provider"`
			Owner                  string   `json:"owner"`
			ExpectedRevision       string   `json:"expectedRevision"`
			ExecutionEnabled       bool     `json:"executionEnabled"`
			ExecutionProven        bool     `json:"executionProven"`
			RuntimeEvidence        string   `json:"runtimeEvidence"`
			RegisteredOperations   []string `json:"registeredOperations"`
			MeshMetadataOperations []string `json:"meshMetadataOperations"`
			RuntimeRouteOperations []string `json:"runtimeRouteOperations"`
		}
		if err := json.Unmarshal([]byte(raw), &evidence); err != nil {
			t.Fatalf("canonical source describe returned invalid JSON for %s: %v; raw=%s", id, err, raw)
		}
		if evidence.Provider != id || evidence.Owner == "" || evidence.ExpectedRevision == "" {
			t.Fatalf("canonical source describe omitted identity/pin/owner for %s: %#v", id, evidence)
		}
		if evidence.ExecutionProven || evidence.RuntimeEvidence != "UNVERIFIED" {
			t.Fatalf("metadata route claimed execution proof for %s: %#v", id, evidence)
		}
		if len(evidence.MeshMetadataOperations) < 2 {
			t.Fatalf("canonical metadata operations missing for %s: %#v", id, evidence)
		}
		if evidence.State != "PROJECTED" && evidence.State != "UNMEASURABLE" && evidence.State != "BLOCKED" {
			t.Fatalf("canonical source describe returned ambiguous state for %s: %#v", id, evidence)
		}
		if result.Status != "degraded" && result.Status != "error" {
			t.Fatalf("metadata route must not imply execution success for %s: result=%#v evidence=%#v", id, result, evidence)
		}
		if evidence.State == "PROJECTED" && evidence.ExecutionEnabled {
			t.Fatalf("projected source %s unexpectedly enabled execution", id)
		}
	}

	for _, id := range []string{"ravana", "langgraph"} {
		result, err := e.Execute(context.Background(), ExternalFederationSourceProbeOperation, nil, map[string]string{"provider": id})
		if err != nil {
			t.Fatalf("canonical source probe failed for %s: %v", id, err)
		}
		var evidence map[string]any
		if err := json.Unmarshal([]byte(result.Metadata["external_result_json"]), &evidence); err != nil {
			t.Fatalf("canonical source probe returned invalid JSON for %s: %v", id, err)
		}
		if evidence["provider"] != id || evidence["executionProven"] != false || evidence["runtimeEvidence"] != "UNVERIFIED" {
			t.Fatalf("canonical source probe reported inconsistent evidence for %s: %#v", id, evidence)
		}
	}
}

func TestCanonicalFederationSourceMetadataRejectsUnknownSource(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	_, err := e.Execute(context.Background(), ExternalFederationSourceDescribeOperation, nil, map[string]string{"provider": "not-registered"})
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_PROVIDER_UNKNOWN:not-registered") {
		t.Fatalf("unknown source was not rejected fail-closed: %v", err)
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
