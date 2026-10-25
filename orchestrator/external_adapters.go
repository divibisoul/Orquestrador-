package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type ExternalAdapterSpec struct {
	ID                string   `json:"id"`
	Owner             string   `json:"owner"`
	Kind              string   `json:"kind"`
	Root              string   `json:"root"`
	Operations        []string `json:"operations"`
	Package           *string  `json:"package"`
	State             string   `json:"state"`
	IntegrationState  string   `json:"integration_state"`
	RuntimeRequirement string  `json:"runtime_requirement"`
}
type ExternalAdapterManifest struct {
	SchemaVersion         string `json:"schema_version"`
	CanonicalControlPlane string `json:"canonical_control_plane"`
	RuntimeDefaults       struct {
		Enabled               bool `json:"enabled"`
		ProbeWithoutExecution bool `json:"probe_without_execution"`
		FailClosed            bool `json:"fail_closed"`
		MaxInputBytes         int  `json:"max_input_bytes"`
		MaxOutputBytes        int  `json:"max_output_bytes"`
	} `json:"runtime_defaults"`
	Runner struct {
		Language       string `json:"language"`
		Path           string `json:"path"`
		InterpreterEnv string `json:"interpreter_env"`
	} `json:"runner"`
	Providers          []ExternalAdapterSpec `json:"providers"`
	ExpansionProviders []ExternalAdapterSpec `json:"expansion_providers"`
}

const (
	ExternalFederationDescribeOperation       = "external.federation.describe@1.0.0"
	ExternalFederationExecuteOperation        = "external.federation.execute@1.0.0"
	ExternalFederationSourceDescribeOperation = "external.federation.source.describe@1.0.0"
	ExternalFederationSourceProbeOperation   = "external.federation.source.probe@1.0.0"
)

type ExternalCapabilityRevision struct {
	ID       string   `json:"id"`
	Revision string   `json:"revision"`
	Source   string   `json:"source"`
	Targets  []string `json:"targets"`
	Target   string   `json:"target"`
}

type ExternalCapabilityRegistry struct {
	Repositories              []ExternalCapabilityRevision `json:"repositories"`
	ComplementaryRepositories []ExternalCapabilityRevision `json:"complementary_repositories"`
	ExpansionRepositories     []ExternalCapabilityRevision `json:"expansion_repositories"`
	GRCEComplements           []ExternalCapabilityRevision `json:"grce_complements"`
}

type ExternalRuntimeAttestation struct {
	SchemaVersion            string            `json:"schema_version"`
	GeneratedFromMaterializedGit bool          `json:"generated_from_materialized_git"`
	Repositories             map[string]string `json:"repositories"`
}

type ExternalAdapterRegistry struct {
	manifest           ExternalAdapterManifest
	byID               map[string]ExternalAdapterSpec
	expansionProviders []ExternalAdapterSpec
	revisions          map[string]string
	manifestPath       string
}

func NewExternalAdapterRegistry() (ExternalAdapterRegistry, error) {
	path := strings.TrimSpace(os.Getenv("SOUL_EXTERNAL_ADAPTER_MANIFEST"))
	root := repositoryRoot()
	if path == "" {
		path = filepath.Join(root, "integrations", "external-adapters.json")
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return ExternalAdapterRegistry{}, fmt.Errorf("external adapter manifest: %w", err)
	}
	var m ExternalAdapterManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return ExternalAdapterRegistry{}, fmt.Errorf("external adapter manifest JSON: %w", err)
	}
	if m.CanonicalControlPlane != "N07" || m.Runner.Path == "" {
		return ExternalAdapterRegistry{}, errors.New("external adapter manifest contract invalid")
	}
	by := make(map[string]ExternalAdapterSpec, len(m.Providers))
	for _, p := range m.Providers {
		if p.ID == "" || p.Owner == "" || p.Root == "" || len(p.Operations) == 0 {
			return ExternalAdapterRegistry{}, errors.New("external adapter provider incomplete")
		}
		if _, ok := by[p.ID]; ok {
			return ExternalAdapterRegistry{}, fmt.Errorf("duplicate external provider: %s", p.ID)
		}
		by[p.ID] = p
	}
	capabilityPath := filepath.Join(root, "integrations", "external-capabilities.json")
	capRaw, err := os.ReadFile(capabilityPath)
	if err != nil {
		return ExternalAdapterRegistry{}, fmt.Errorf("external capability registry: %w", err)
	}
	var capabilityRegistry ExternalCapabilityRegistry
	if err := json.Unmarshal(capRaw, &capabilityRegistry); err != nil {
		return ExternalAdapterRegistry{}, fmt.Errorf("external capability registry JSON: %w", err)
	}
	allRevisions := append([]ExternalCapabilityRevision{}, capabilityRegistry.Repositories...)
	allRevisions = append(allRevisions, capabilityRegistry.ComplementaryRepositories...)
	allRevisions = append(allRevisions, capabilityRegistry.ExpansionRepositories...)
	allRevisions = append(allRevisions, capabilityRegistry.GRCEComplements...)
	revisions := make(map[string]string, len(allRevisions))
	for _, item := range allRevisions {
		if item.ID == "" || !isGitRevision(item.Revision) {
			return ExternalAdapterRegistry{}, fmt.Errorf("external capability revision invalid: %s", item.ID)
		}
		if prior, exists := revisions[item.ID]; exists {
			if prior != item.Revision {
				return ExternalAdapterRegistry{}, fmt.Errorf("conflicting external capability revision: %s", item.ID)
			}
			continue // hora-graph-core is deliberately reused by active and expansion ledgers.
		}
		revisions[item.ID] = item.Revision
	}
	for id := range by {
		if _, ok := revisions[id]; !ok {
			return ExternalAdapterRegistry{}, fmt.Errorf("external capability revision missing: %s", id)
		}
	}

	canonicalOwners := map[string]bool{"N01": true, "N02": true, "N03": true, "N04": true, "N05": true, "N06": true, "N07": true}
	projected := make([]ExternalAdapterSpec, 0, len(m.ExpansionProviders))
	seenProjected := make(map[string]struct{}, len(m.ExpansionProviders))
	for _, p := range m.ExpansionProviders {
		if p.ID == "" || p.Owner == "" || p.Root == "" || len(p.Operations) == 0 {
			return ExternalAdapterRegistry{}, errors.New("external expansion provider incomplete")
		}
		if p.State != "PROJECTED" || p.IntegrationState == "" || p.RuntimeRequirement == "" {
			return ExternalAdapterRegistry{}, fmt.Errorf("external expansion provider must remain explicitly projected: %s", p.ID)
		}
		if !canonicalOwners[p.Owner] {
			return ExternalAdapterRegistry{}, fmt.Errorf("external expansion provider owner invalid: %s:%s", p.ID, p.Owner)
		}
		if _, exists := by[p.ID]; exists {
			return ExternalAdapterRegistry{}, fmt.Errorf("projected expansion duplicates active provider: %s", p.ID)
		}
		if _, exists := seenProjected[p.ID]; exists {
			return ExternalAdapterRegistry{}, fmt.Errorf("duplicate projected expansion provider: %s", p.ID)
		}
		seenProjected[p.ID] = struct{}{}
		if _, ok := revisions[p.ID]; !ok {
			return ExternalAdapterRegistry{}, fmt.Errorf("external expansion revision missing: %s", p.ID)
		}
		expectedRoot := filepath.Clean(filepath.Join("integrations", "external", p.ID))
		if filepath.Clean(p.Root) != expectedRoot {
			return ExternalAdapterRegistry{}, fmt.Errorf("external expansion root non-canonical: %s", p.ID)
		}
		targetDeclared := false
		for _, source := range append(append([]ExternalCapabilityRevision{}, capabilityRegistry.ExpansionRepositories...), capabilityRegistry.GRCEComplements...) {
			if source.ID != p.ID {
				continue
			}
			for _, target := range source.Targets {
				if target == p.Owner {
					targetDeclared = true
				}
			}
			if source.Target == p.Owner {
				targetDeclared = true
			}
		}
		if !targetDeclared {
			return ExternalAdapterRegistry{}, fmt.Errorf("external expansion owner is not a declared target: %s:%s", p.ID, p.Owner)
		}
		projected = append(projected, p)
	}
	return ExternalAdapterRegistry{manifest: m, byID: by, expansionProviders: projected, revisions: revisions, manifestPath: path}, nil
}
type externalSourceObservation struct {
	Present          bool
	ExpectedRevision string
	ObservedRevision string
	RevisionMatched  bool
	State            string
}

// inspectExternalSource treats a path as a source checkout only when it is the
// root of its own Git worktree, then compares HEAD with the recorded pin.
// os.Stat alone is insufficient: an uninitialized gitlink can appear as an
// empty directory and must never be reported as a present upstream repository.
func inspectExternalSource(root, expectedRevision string) externalSourceObservation {
	return inspectExternalSourceWithRepositoryRoot(root, expectedRevision, repositoryRoot())
}

// inspectExternalSourceWithRepositoryRoot validates source provenance in either
// a live checkout or the production image. Runtime images intentionally omit
// Git metadata, so only the build-generated attestation may stand in for HEAD;
// it is accepted only for the canonical provider path, an exact schema/pin,
// and a non-empty copied source tree.
func inspectExternalSourceWithRepositoryRoot(root, expectedRevision, repoRoot string) externalSourceObservation {
	observation := externalSourceObservation{
		ExpectedRevision: strings.ToLower(strings.TrimSpace(expectedRevision)),
		State:            "MISSING_WORKTREE",
	}
	absoluteRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return observation
	}
	resolvedRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return observation
	}

	topCommand := exec.Command("git", "-C", resolvedRoot, "rev-parse", "--show-toplevel")
	topBytes, topErr := topCommand.Output()
	if topErr == nil {
		resolvedTop, resolveErr := filepath.EvalSymlinks(strings.TrimSpace(string(topBytes)))
		// If Git resolves this path to a parent checkout, it is an uninitialized
		// gitlink. Do not use an image attestation to hide that condition.
		if resolveErr != nil || filepath.Clean(resolvedTop) != filepath.Clean(resolvedRoot) {
			return observation
		}
		headCommand := exec.Command("git", "-C", resolvedRoot, "rev-parse", "--verify", "HEAD")
		headBytes, err := headCommand.Output()
		if err != nil {
			return observation
		}
		observed := strings.ToLower(strings.TrimSpace(string(headBytes)))
		if !isGitRevision(observed) {
			return observation
		}
		observation.Present = true
		observation.ObservedRevision = observed
		observation.RevisionMatched = isGitRevision(observation.ExpectedRevision) && observed == observation.ExpectedRevision
		if observation.RevisionMatched {
			observation.State = "PIN_MATCH"
		} else {
			observation.State = "REVISION_MISMATCH"
		}
		return observation
	}

	// A production image deliberately excludes .git and Git may not even be
	// installed. Accept only the build attestation generated after Dockerfile
	// verified each worktree HEAD against its parent gitlink.
	rootAbs, err := filepath.Abs(filepath.Clean(repoRoot))
	if err != nil {
		return observation
	}
	expectedRoot := filepath.Join(rootAbs, "integrations", "external", filepath.Base(resolvedRoot))
	if filepath.Clean(expectedRoot) != filepath.Clean(resolvedRoot) {
		return observation
	}
	entries, err := os.ReadDir(resolvedRoot)
	if err != nil {
		return observation
	}
	hasContent := false
	for _, entry := range entries {
		if entry.Name() != ".git" {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return observation
	}

	attestationPath := strings.TrimSpace(os.Getenv("SOUL_EXTERNAL_RUNTIME_ATTESTATION"))
	if attestationPath == "" {
		attestationPath = filepath.Join(rootAbs, "config", "soul-external-runtime-attestation.json")
	} else if !filepath.IsAbs(attestationPath) {
		attestationPath = filepath.Join(rootAbs, attestationPath)
	}
	raw, err := os.ReadFile(attestationPath)
	if err != nil {
		return observation
	}
	var attestation ExternalRuntimeAttestation
	if err := json.Unmarshal(raw, &attestation); err != nil ||
		attestation.SchemaVersion != "soul.external.runtime-attestation.v1" ||
		!attestation.GeneratedFromMaterializedGit {
		return observation
	}
	observed := strings.ToLower(strings.TrimSpace(attestation.Repositories[filepath.Base(resolvedRoot)]))
	if !isGitRevision(observed) {
		return observation
	}
	observation.Present = true
	observation.ObservedRevision = observed
	observation.RevisionMatched = isGitRevision(observation.ExpectedRevision) && observed == observation.ExpectedRevision
	if observation.RevisionMatched {
		observation.State = "PIN_MATCH"
	} else {
		observation.State = "REVISION_MISMATCH"
	}
	return observation
}

func isGitRevision(revision string) bool {
	if len(revision) != 40 {
		return false
	}
	for _, char := range revision {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

func externalEnabled(provider string) bool {
	key := "SOUL_EXTERNAL_EXECUTE_" + strings.ToUpper(strings.ReplaceAll(provider, "-", "_"))
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}
func (r ExternalAdapterRegistry) python() string {
	if v := strings.TrimSpace(os.Getenv(r.manifest.Runner.InterpreterEnv)); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("SOUL_EXTERNAL_PYTHON")); v != "" {
		return v
	}
	return "python3"
}
func (r ExternalAdapterRegistry) operationAllowed(provider, mode, operation string) bool {
	if strings.EqualFold(strings.TrimSpace(mode), "probe") && operation == "probe" {
		return true
	}
	p, ok := r.byID[strings.ToLower(strings.TrimSpace(provider))]
	if !ok {
		return false
	}
	for _, allowed := range p.Operations {
		if allowed == operation {
			return true
		}
	}
	return false
}

func (r ExternalAdapterRegistry) validateMaterializedPin(ctx context.Context, p ExternalAdapterSpec) error {
	revision := r.revisions[p.ID]
	rootDir := repositoryRoot()
	root := p.Root
	if !filepath.IsAbs(root) {
		root = filepath.Join(rootDir, root)
	}
	root = filepath.Clean(root)
	canonical := filepath.Clean(filepath.Join(rootDir, "integrations", "external", p.ID))
	if root != canonical {
		return fmt.Errorf("EXTERNAL_PROVIDER_ROOT_NON_CANONICAL:%s", p.ID)
	}
	rel, err := filepath.Rel(rootDir, root)
	if err != nil || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
		return fmt.Errorf("EXTERNAL_PROVIDER_ROOT_INVALID:%s", p.ID)
	}
	gitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	gitDir := filepath.Join(rootDir, ".git")
	if _, statErr := os.Stat(gitDir); statErr == nil {
		treeCmd := exec.CommandContext(gitCtx, "git", "-C", rootDir, "ls-tree", "HEAD", "--", filepath.ToSlash(rel))
		treeOut, err := treeCmd.Output()
		if err != nil {
			if gitCtx.Err() != nil {
				return fmt.Errorf("EXTERNAL_GITLINK_CHECK_TIMEOUT:%s", p.ID)
			}
			return fmt.Errorf("EXTERNAL_GITLINK_CHECK_FAILED:%s", p.ID)
		}
		fields := strings.Fields(string(treeOut))
		if len(fields) < 3 || fields[0] != "160000" || fields[1] != "commit" || fields[2] != revision {
			return fmt.Errorf("EXTERNAL_GITLINK_REVISION_MISMATCH:%s:expected=%s", p.ID, revision)
		}
		// A nested provider path must resolve to its own Git worktree. Without
		// this check git -C may walk up to the N07 parent and return its HEAD.
		topCmd := exec.CommandContext(gitCtx, "git", "-C", root, "rev-parse", "--show-toplevel")
		topOut, err := topCmd.Output()
		if err != nil {
			if gitCtx.Err() != nil {
				return fmt.Errorf("EXTERNAL_SUBMODULE_ROOT_CHECK_TIMEOUT:%s", p.ID)
			}
			return fmt.Errorf("EXTERNAL_SUBMODULE_NOT_MATERIALIZED:%s", p.ID)
		}
		resolvedRoot, rootErr := filepath.EvalSymlinks(root)
		resolvedTop, topErr := filepath.EvalSymlinks(strings.TrimSpace(string(topOut)))
		if rootErr != nil || topErr != nil ||
			filepath.Clean(resolvedRoot) != filepath.Clean(resolvedTop) {
			return fmt.Errorf("EXTERNAL_SUBMODULE_NOT_MATERIALIZED:%s", p.ID)
		}
		headCmd := exec.CommandContext(gitCtx, "git", "-C", root, "rev-parse", "HEAD")
		headOut, err := headCmd.Output()
		if err != nil {
			if gitCtx.Err() != nil {
				return fmt.Errorf("EXTERNAL_SUBMODULE_HEAD_CHECK_TIMEOUT:%s", p.ID)
			}
			return fmt.Errorf("EXTERNAL_SUBMODULE_NOT_MATERIALIZED:%s", p.ID)
		}
		if strings.TrimSpace(string(headOut)) != revision {
			return fmt.Errorf("EXTERNAL_SUBMODULE_HEAD_MISMATCH:%s:expected=%s:actual=%s", p.ID, revision, strings.TrimSpace(string(headOut)))
		}
		if diffCmd := exec.CommandContext(gitCtx, "git", "-C", root, "diff", "--quiet"); diffCmd.Run() != nil {
			return fmt.Errorf("EXTERNAL_SUBMODULE_WORKTREE_DIRTY:%s", p.ID)
		}
		if cachedCmd := exec.CommandContext(gitCtx, "git", "-C", root, "diff", "--cached", "--quiet"); cachedCmd.Run() != nil {
			return fmt.Errorf("EXTERNAL_SUBMODULE_INDEX_DIRTY:%s", p.ID)
		}
		return nil
	}
	attestationPath := strings.TrimSpace(os.Getenv("SOUL_EXTERNAL_RUNTIME_ATTESTATION"))
	if attestationPath == "" {
		attestationPath = filepath.Join(rootDir, "config", "soul-external-runtime-attestation.json")
	} else if !filepath.IsAbs(attestationPath) {
		attestationPath = filepath.Join(rootDir, attestationPath)
	}
	raw, err := os.ReadFile(attestationPath)
	if err != nil {
		return fmt.Errorf("EXTERNAL_RUNTIME_ATTESTATION_UNAVAILABLE:%s", p.ID)
	}
	var attestation ExternalRuntimeAttestation
	if err := json.Unmarshal(raw, &attestation); err != nil {
		return fmt.Errorf("EXTERNAL_RUNTIME_ATTESTATION_INVALID:%s", p.ID)
	}
	if attestation.Repositories[p.ID] != revision {
		return fmt.Errorf("EXTERNAL_RUNTIME_ATTESTATION_REVISION_MISMATCH:%s:expected=%s", p.ID, revision)
	}
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("EXTERNAL_SUBMODULE_SOURCE_NOT_PRESENT:%s", p.ID)
	}
	return nil
}

func buildExternalAdapterRequest(provider, mode, operation, root string, metadata map[string]string) map[string]any {
	req := map[string]any{"metadata": metadata}
	for k, v := range metadata {
		req[k] = v
	}
	// These fields are transport authority, never caller-controlled metadata.
	req["provider"] = provider
	req["mode"] = mode
	req["operation"] = operation
	req["root"] = root
	return req
}

func (r ExternalAdapterRegistry) run(ctx context.Context, provider, mode, operation string, metadata map[string]string) (map[string]any, error) {
	p, ok := r.byID[strings.ToLower(strings.TrimSpace(provider))]
	if !ok {
		return nil, fmt.Errorf("EXTERNAL_PROVIDER_UNKNOWN:%s", provider)
	}
	if !r.operationAllowed(p.ID, mode, operation) {
		return nil, fmt.Errorf("EXTERNAL_OPERATION_NOT_REGISTERED:%s:%s", p.ID, operation)
	}
	if err := r.validateMaterializedPin(ctx, p); err != nil {
		return nil, err
	}
	root := p.Root
	if !filepath.IsAbs(root) {
		root = filepath.Join(repositoryRoot(), root)
	}
	root = filepath.Clean(root)
	externalRoot := filepath.Join(repositoryRoot(), "integrations", "external")
	if !strings.HasPrefix(root, externalRoot+string(os.PathSeparator)) {
		return nil, errors.New("EXTERNAL_PROVIDER_ROOT_INVALID")
	}
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("EXTERNAL_SOURCE_NOT_PRESENT:%s", provider)
	}
	runner := r.manifest.Runner.Path
	if !filepath.IsAbs(runner) {
		runner = filepath.Join(repositoryRoot(), runner)
	}
	runner = filepath.Clean(runner)
	if _, err := os.Stat(runner); err != nil {
		return nil, errors.New("EXTERNAL_ADAPTER_RUNNER_NOT_FOUND")
	}
	req := buildExternalAdapterRequest(provider, mode, operation, root, metadata)
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	max := r.manifest.RuntimeDefaults.MaxInputBytes
	if max <= 0 {
		max = 131072
	}
	if len(data) > max {
		return nil, errors.New("EXTERNAL_ADAPTER_INPUT_TOO_LARGE")
	}
	timeout := 60 * time.Second
	if raw := strings.TrimSpace(metadata["timeout_seconds"]); raw != "" {
		if n, e := strconv.Atoi(raw); e == nil && n > 0 && n <= 600 {
			timeout = time.Duration(n) * time.Second
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, r.python(), runner)
	cmd.Stdin = bytes.NewReader(data)
	var out, er bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &er
	if e := cmd.Run(); runCtx.Err() != nil {
		return nil, fmt.Errorf("EXTERNAL_ADAPTER_TIMEOUT:%s", provider)
	} else if e != nil {
		var result map[string]any
		if json.Unmarshal(out.Bytes(), &result) == nil && result["code"] != nil {
			return result, errors.New(fmt.Sprint(result["code"]))
		}
		return nil, fmt.Errorf("EXTERNAL_ADAPTER_PROCESS_FAILED:%s:%s", provider, strings.TrimSpace(er.String()))
	}
	var result map[string]any
	if e := json.Unmarshal(out.Bytes(), &result); e != nil {
		return nil, fmt.Errorf("EXTERNAL_ADAPTER_INVALID_OUTPUT:%s", provider)
	}
	return result, nil
}
func RegisterExternalAdapterOperations(e *Engine, r ExternalAdapterRegistry) error {
	if e == nil {
		return errors.New("engine is required")
	}
	if err := e.Register(ExternalFederationDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		providers := make([]map[string]any, 0, len(r.byID))
		for _, p := range r.byID {
			root := p.Root
			if !filepath.IsAbs(root) {
				root = filepath.Join(repositoryRoot(), root)
			}
			observation := inspectExternalSource(filepath.Clean(root), r.revisions[p.ID])
			providers = append(providers, map[string]any{
				"id": p.ID, "owner": p.Owner, "kind": p.Kind, "root": root,
				"revision": r.revisions[p.ID], "expectedRevision": observation.ExpectedRevision,
				"observedRevision": observation.ObservedRevision, "revisionMatched": observation.RevisionMatched,
				"sourceState": observation.State, "operations": p.Operations,
				"sourcePresent": observation.Present, "executionEnabled": externalEnabled(p.ID),
			})
		}
		expansionProviders := make([]map[string]any, 0, len(r.expansionProviders))
		for _, p := range r.expansionProviders {
			root := p.Root
			if !filepath.IsAbs(root) {
				root = filepath.Join(repositoryRoot(), root)
			}
			observation := inspectExternalSource(filepath.Clean(root), r.revisions[p.ID])
			expansionProviders = append(expansionProviders, map[string]any{
				"id": p.ID, "owner": p.Owner, "kind": p.Kind, "root": root,
				"revision": r.revisions[p.ID], "expectedRevision": observation.ExpectedRevision,
				"observedRevision": observation.ObservedRevision, "revisionMatched": observation.RevisionMatched,
				"sourceState": observation.State, "sourcePresent": observation.Present,
				"operations": p.Operations, "state": p.State, "integrationState": p.IntegrationState,
				"runtimeRequirement": p.RuntimeRequirement, "executionEnabled": false,
				"meshMetadataOperations": []string{"external." + p.ID + ".describe@1.0.0", "external." + p.ID + ".probe@1.0.0"},
				"runtimeRouteOperations": []string{},
				"executionProven": false,
			})
		}
		raw, _ := json.Marshal(map[string]any{
			"state": "PASS",
			"catalog_state": "DECLARED_WITH_PROJECTED_EXPANSIONS",
			"runtime_evidence": "UNVERIFIED",
			"control_plane": "N07",
			"provider_count": len(providers),
			"expansion_provider_count": len(expansionProviders),
			"declared_external_repository_count": len(r.revisions),
			"manifest": r.manifestPath,
			"providers": providers,
			"expansion_providers": expansionProviders,
		})
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-federation", Target: m.Source, Status: "ok", Metadata: map[string]string{"federation_json": string(raw)}}, nil
	}); err != nil {
		return err
	}
	// Canonical source metadata operations give every nucleus one stable
	// Mesh call shape for any external repository. Per-provider aliases remain
	// registered below for backwards compatibility.
	for _, operation := range []string{ExternalFederationSourceDescribeOperation, ExternalFederationSourceProbeOperation} {
		op := operation
		if err := e.Register(op, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			provider := strings.ToLower(strings.TrimSpace(m.Metadata["provider"]))
			if provider == "" {
				return externalAdapterFailure(m, "EXTERNAL_PROVIDER_REQUIRED")
			}
			result, err := externalFederationSourceMetadata(r, provider, op)
			if err != nil {
				return externalAdapterFailure(m, err.Error())
			}
			if op == ExternalFederationSourceProbeOperation {
				result = externalFederationSourceProbe(ctx, r, provider, m.Metadata, result)
			}
			return externalAdapterResult(m, result), nil
		}); err != nil {
			return err
		}
	}
	if err := e.Register(ExternalFederationExecuteOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		provider := strings.TrimSpace(m.Metadata["provider"])
		op := strings.TrimSpace(m.Metadata["external_operation"])
		if provider == "" || op == "" {
			return externalAdapterFailure(m, "EXTERNAL_PROVIDER_AND_OPERATION_REQUIRED")
		}
		if !externalEnabled(provider) {
			return externalAdapterFailure(m, "EXTERNAL_ADAPTER_DISABLED")
		}
		result, err := r.run(ctx, provider, "execute", op, m.Metadata)
		if err != nil {
			return externalAdapterFailure(m, err.Error())
		}
		return externalAdapterExecutionResult(m, result)
	}); err != nil {
		return err
	}
	// Expose source-state metadata for every projected expansion source through
	// the same Mesh control plane. These routes inspect provenance only; they do
	// not invoke external code and deliberately have no execute counterpart.
	for _, p := range r.expansionProviders {
		spec := p
		for _, routeKind := range []string{"describe", "probe"} {
			kind := routeKind
			opName := "external." + spec.ID + "." + kind + "@1.0.0"
			if err := e.Register(opName, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
				return externalProjectedSourceResult(m, r, spec, kind), nil
			}); err != nil {
				return err
			}
		}
	}
	for id := range r.byID {
		provider := id
		spec := r.byID[provider]
		registeredDescribe := false
		for _, registeredOperation := range spec.Operations {
			if registeredDescribe || !strings.HasSuffix(registeredOperation, ".describe") {
				continue
			}
			registeredDescribe = true
			describeOperation := "external." + provider + ".describe@1.0.0"
			describeProvider := provider
			describeSourceOperation := registeredOperation
			if err := e.Register(describeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
				result, err := r.run(ctx, describeProvider, "describe", describeSourceOperation, m.Metadata)
				if err != nil {
					return externalAdapterFailure(m, err.Error())
				}
				return externalAdapterResult(m, result), nil
			}); err != nil {
				return err
			}
		}
		if err := e.Register("external."+provider+".probe@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			result, err := r.run(ctx, provider, "probe", "probe", m.Metadata)
			if err != nil {
				return externalAdapterFailure(m, err.Error())
			}
			return externalAdapterResult(m, result), nil
		}); err != nil {
			return err
		}
		if err := e.Register("external."+provider+".execute@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			op := strings.TrimSpace(m.Metadata["external_operation"])
			if op == "" {
				return externalAdapterFailure(m, "EXTERNAL_OPERATION_REQUIRED")
			}
			if !externalEnabled(provider) {
				return externalAdapterFailure(m, "EXTERNAL_ADAPTER_DISABLED")
			}
			result, err := r.run(ctx, provider, "execute", op, m.Metadata)
			if err != nil {
				return externalAdapterFailure(m, err.Error())
			}
			return externalAdapterExecutionResult(m, result)
		}); err != nil {
			return err
		}
	}
	return nil
}
func externalFederationSourceMetadata(r ExternalAdapterRegistry, providerID, operation string) (map[string]any, error) {
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	var spec ExternalAdapterSpec
	projected := false
	if active, ok := r.byID[providerID]; ok {
		spec = active
	} else {
		found := false
		for _, candidate := range r.expansionProviders {
			if candidate.ID == providerID {
				spec = candidate
				projected = true
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("EXTERNAL_PROVIDER_UNKNOWN:%s", providerID)
		}
	}
	root := spec.Root
	if !filepath.IsAbs(root) {
		root = filepath.Join(repositoryRoot(), root)
	}
	root = filepath.Clean(root)
	observation := inspectExternalSource(root, r.revisions[providerID])
	state := "BLOCKED"
	if projected {
		if observation.Present && observation.RevisionMatched {
			state = "PROJECTED"
		}
	} else if observation.Present && observation.RevisionMatched {
		// Exact source provenance is measurable, but it is not proof of a
		// successful provider invocation or live external service.
		state = "UNMEASURABLE"
	}
	runtimeRoutes := []string{}
	executionEnabled := !projected && externalEnabled(providerID)
	if executionEnabled {
		runtimeRoutes = append(runtimeRoutes,
			ExternalFederationExecuteOperation,
			"external."+providerID+".execute@1.0.0",
		)
	}
	operations := []string{
		ExternalFederationSourceDescribeOperation,
		ExternalFederationSourceProbeOperation,
	}
	if projected {
		operations = append(operations,
			"external."+providerID+".describe@1.0.0",
			"external."+providerID+".probe@1.0.0",
		)
	}
	return map[string]any{
		"state": state,
		"provider": providerID,
		"owner": spec.Owner,
		"operation": operation,
		"root": root,
		"revision": r.revisions[providerID],
		"expectedRevision": observation.ExpectedRevision,
		"observedRevision": observation.ObservedRevision,
		"revisionMatched": observation.RevisionMatched,
		"sourceState": observation.State,
		"sourcePresent": observation.Present,
		"adapterState": func() string { if projected { return "PROJECTED" }; return "REGISTERED" }(),
		"registeredOperations": spec.Operations,
		"meshMetadataOperations": operations,
		"runtimeRouteOperations": runtimeRoutes,
		"executionEnabled": executionEnabled,
		"executionProven": false,
		"runtimeEvidence": "UNVERIFIED",
		"requiredEvidence": []string{
			"native-adapter-contract",
			"exact-pinned-source",
			"deterministic-contract-tests",
			"authenticated-runtime-e2e",
		},
	}, nil
}

// externalFederationSourceProbe reports a read-only health signal for a source.
 // Active adapters verify the exact pinned worktree and expose package importability.
 // Projected sources remain metadata-only and cannot gain an execution route here.
 func externalFederationSourceProbe(ctx context.Context, r ExternalAdapterRegistry, providerID string, metadata map[string]string, evidence map[string]any) map[string]any {
 	if _, active := r.byID[providerID]; active {
 		probe, err := r.run(ctx, providerID, "probe", "probe", metadata)
 		if err != nil {
 			evidence["sourceProbe"] = map[string]any{
 				"state": "BLOCKED",
 				"provider": providerID,
 				"mode": "probe",
 				"code": err.Error(),
 				"source_integrity_only": true,
 				"execution_proven": false,
 				"runtime_evidence": "UNVERIFIED",
 			}
 			evidence["sourceProbeState"] = "BLOCKED"
 			return evidence
 		}
 		evidence["sourceProbe"] = probe
 		state, _ := probe["state"].(string)
 		if state == "" {
 			state = "BLOCKED"
 		}
 		evidence["sourceProbeState"] = state
 		return evidence
 	}
 	state, _ := evidence["state"].(string)
 	evidence["sourceProbeState"] = state
 	evidence["sourceProbe"] = map[string]any{
 		"state": state,
 		"provider": providerID,
 		"mode": "source-integrity-only",
 		"source_present": evidence["sourcePresent"],
 		"revision_matched": evidence["revisionMatched"],
 		"source_integrity_only": true,
 		"execution_enabled": false,
 		"execution_proven": false,
 		"runtime_evidence": "UNVERIFIED",
 	}
 	return evidence
 }

func externalProjectedSourceResult(m protocol.Message, r ExternalAdapterRegistry, p ExternalAdapterSpec, operation string) protocol.Result {
	root := p.Root
	if !filepath.IsAbs(root) {
		root = filepath.Join(repositoryRoot(), root)
	}
	root = filepath.Clean(root)
	observation := inspectExternalSource(root, r.revisions[p.ID])
	state := "BLOCKED"
	if observation.Present && observation.RevisionMatched {
		state = "PROJECTED"
	}
	result := map[string]any{
		"state": state,
		"provider": p.ID,
		"owner": p.Owner,
		"operation": operation,
		"root": root,
		"revision": r.revisions[p.ID],
		"expectedRevision": observation.ExpectedRevision,
		"observedRevision": observation.ObservedRevision,
		"revisionMatched": observation.RevisionMatched,
		"sourceState": observation.State,
		"sourcePresent": observation.Present,
		"plannedOperations": p.Operations,
		"runtimeRouteOperations": []string{},
		"executionEnabled": false,
		"executionProven": false,
		"runtimeEvidence": "UNVERIFIED",
		"requiredEvidence": []string{
			"native-adapter-contract",
			"exact-pinned-source",
			"deterministic-contract-tests",
			"authenticated-runtime-e2e",
		},
	}
	return externalAdapterResult(m, result)
}

func externalAdapterFailure(m protocol.Message, code string) (protocol.Result, error) {
	err := errors.New(code)
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-federation", Target: m.Source, Status: "error", Error: code}, err
}
func repositoryRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, e := os.Stat(filepath.Join(cwd, "integrations", "external-adapters.json")); e == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return cwd
		}
		cwd = parent
	}
}

func externalAdapterResult(m protocol.Message, result map[string]any) protocol.Result {
	raw, _ := json.Marshal(result)
	state, _ := result["state"].(string)
	status := "error"
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "PASS", "REAL", "ACTIVE":
		status = "ok"
	case "DEGRADED", "PROJECTED", "UNMEASURABLE":
		status = "degraded"
	case "BLOCKED", "FAIL":
		status = "error"
	}
	provider, _ := result["provider"].(string)
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-federation", Target: m.Source, Status: status, Metadata: map[string]string{"provider": provider, "external_result_json": string(raw)}}
}

func externalAdapterExecutionResult(m protocol.Message, result map[string]any) (protocol.Result, error) {
	state, _ := result["state"].(string)
	if !strings.EqualFold(strings.TrimSpace(state), "PASS") {
		code, _ := result["code"].(string)
		if code == "" {
			code = "EXTERNAL_EXECUTION_NOT_PROVEN"
		}
		return externalAdapterFailure(m, code)
	}
	return externalAdapterResult(m, result), nil
}
