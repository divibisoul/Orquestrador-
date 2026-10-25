package orchestrator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectExternalSourceUsesBuildAttestationWhenImageOmitsGitMetadata(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "integrations", "external", "langgraph")
	if err := os.MkdirAll(sourceRoot, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("materialized external source\n"), 0600); err != nil {
		t.Fatal(err)
	}

	revision := strings.Repeat("a", 40)
	attestationPath := filepath.Join(root, "config", "soul-external-runtime-attestation.json")
	if err := os.MkdirAll(filepath.Dir(attestationPath), 0700); err != nil {
		t.Fatal(err)
	}
	attestation := ExternalRuntimeAttestation{
		SchemaVersion: "soul.external.runtime-attestation.v1",
		GeneratedFromMaterializedGit: true,
		Repositories:  map[string]string{"langgraph": revision},
	}
	raw, err := json.Marshal(attestation)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attestationPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SOUL_EXTERNAL_RUNTIME_ATTESTATION", attestationPath)

	got := inspectAttestedExternalSource(sourceRoot, revision, attestationPath)
	if !got.Present || !got.RevisionMatched || got.State != "ATTESTED_PIN_MATCH" ||
		got.EvidenceMode != "build-attestation" || got.ObservedRevision != revision {
		t.Fatalf("attested source not recognized: %#v", got)
	}
}

func TestInspectExternalSourceAttestationMustMatchExpectedPin(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "integrations", "external", "langgraph")
	if err := os.MkdirAll(sourceRoot, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("materialized external source\n"), 0600); err != nil {
		t.Fatal(err)
	}

	attestedRevision := strings.Repeat("a", 40)
	expectedRevision := strings.Repeat("b", 40)
	attestationPath := filepath.Join(root, "config", "soul-external-runtime-attestation.json")
	if err := os.MkdirAll(filepath.Dir(attestationPath), 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(ExternalRuntimeAttestation{
		SchemaVersion: "soul.external.runtime-attestation.v1",
		GeneratedFromMaterializedGit: true,
		Repositories:  map[string]string{"langgraph": attestedRevision},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attestationPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SOUL_EXTERNAL_RUNTIME_ATTESTATION", attestationPath)

	got := inspectAttestedExternalSource(sourceRoot, expectedRevision, attestationPath)
	if !got.Present || got.RevisionMatched || got.State != "ATTESTATION_REVISION_MISMATCH" ||
		got.EvidenceMode != "build-attestation" || got.ObservedRevision != attestedRevision {
		t.Fatalf("mismatched attestation incorrectly accepted: %#v", got)
	}
}

func TestInspectExternalSourceDoesNotAcceptGitMetadataOnlyDirectory(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "integrations", "external", "langgraph")
	if err := os.MkdirAll(sourceRoot, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, ".git"), []byte("gitdir: ../.git/modules/langgraph\n"), 0600); err != nil {
		t.Fatal(err)
	}
	attestationPath := filepath.Join(root, "config", "soul-external-runtime-attestation.json")
	if err := os.MkdirAll(filepath.Dir(attestationPath), 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(ExternalRuntimeAttestation{
		SchemaVersion: "soul.external.runtime-attestation.v1",
		GeneratedFromMaterializedGit: true,
		Repositories:  map[string]string{"langgraph": strings.Repeat("a", 40)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(attestationPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SOUL_EXTERNAL_RUNTIME_ATTESTATION", attestationPath)

	got := inspectExternalSource(sourceRoot, strings.Repeat("a", 40))
	if got.Present || got.RevisionMatched || got.State != "MISSING_WORKTREE" {
		t.Fatalf("metadata-only empty worktree misclassified: %#v", got)
	}
}


// The build marker is required before an attestation can stand in for a Git HEAD.
func TestInspectAttestedExternalSourceRejectsUntrustedBuildMarker(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("source files\n"), 0600); err != nil {
		t.Fatal(err)
	}
	revision := strings.Repeat("a", 40)
	attestationPath := filepath.Join(t.TempDir(), "attestation.json")
	raw, err := json.Marshal(ExternalRuntimeAttestation{
		SchemaVersion: "soul.external.runtime-attestation.v1",
		Repositories: map[string]string{"provider": revision},
	})
	if err != nil { t.Fatal(err) }
	if err := os.WriteFile(attestationPath, raw, 0600); err != nil { t.Fatal(err) }
	got := inspectAttestedExternalSource(root, revision, attestationPath)
	if !got.Present || got.RevisionMatched || got.ObservedRevision != "" || got.EvidenceMode != "filesystem-unattested" || got.State != "ATTESTATION_UNTRUSTED" {
		t.Fatalf("untrusted attestation promoted source: %#v", got)
	}
}
