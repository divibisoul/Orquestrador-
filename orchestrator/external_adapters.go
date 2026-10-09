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
	ID         string   `json:"id"`
	Owner      string   `json:"owner"`
	Kind       string   `json:"kind"`
	Root       string   `json:"root"`
	Operations []string `json:"operations"`
	Package    *string  `json:"package"`
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
	Providers []ExternalAdapterSpec `json:"providers"`
}

const (
	ExternalFederationDescribeOperation = "external.federation.describe@1.0.0"
	ExternalFederationExecuteOperation  = "external.federation.execute@1.0.0"
)

type ExternalCapabilityRevision struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}

type ExternalCapabilityRegistry struct {
	Repositories              []ExternalCapabilityRevision `json:"repositories"`
	ComplementaryRepositories []ExternalCapabilityRevision `json:"complementary_repositories"`
}

type ExternalRuntimeAttestation struct {
	SchemaVersion string            `json:"schema_version"`
	Repositories  map[string]string `json:"repositories"`
}

type ExternalAdapterRegistry struct {
	manifest     ExternalAdapterManifest
	byID         map[string]ExternalAdapterSpec
	revisions    map[string]string
	manifestPath string
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
	allRevisions := append(append([]ExternalCapabilityRevision{}, capabilityRegistry.Repositories...), capabilityRegistry.ComplementaryRepositories...)
	revisions := make(map[string]string, len(allRevisions))
	for _, item := range allRevisions {
		if item.ID == "" || len(item.Revision) != 40 {
			return ExternalAdapterRegistry{}, fmt.Errorf("external capability revision invalid: %s", item.ID)
		}
		if _, exists := revisions[item.ID]; exists {
			return ExternalAdapterRegistry{}, fmt.Errorf("duplicate external capability revision: %s", item.ID)
		}
		revisions[item.ID] = item.Revision
	}
	for id := range by {
		if _, ok := revisions[id]; !ok {
			return ExternalAdapterRegistry{}, fmt.Errorf("external capability revision missing: %s", id)
		}
	}
	return ExternalAdapterRegistry{manifest: m, byID: by, revisions: revisions, manifestPath: path}, nil
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

// validateRegisteredPin validates the immutable source pin from the parent repository
// without fetching or executing the public source. It is intended only for
// read-only probe/describe operations; execute mode still requires a clean,
// materialized worktree at the exact pinned revision.
func (r ExternalAdapterRegistry) validateRegisteredPin(ctx context.Context, p ExternalAdapterSpec) error {
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
	if _, err := os.Stat(filepath.Join(rootDir, ".git")); err != nil {
		return fmt.Errorf("EXTERNAL_GITLINK_CHECK_UNAVAILABLE:%s", p.ID)
	}
	treeCmd := exec.CommandContext(gitCtx, "git", "-C", rootDir, "ls-tree", "HEAD", "--", filepath.ToSlash(rel))
	treeOut, err := treeCmd.Output()
	if err != nil {
		if gitCtx.Err() != nil {
			return fmt.Errorf("EXTERNAL_GITLINK_CHECK_TIMEOUT:%s", p.ID)
		}
		return fmt.Errorf("EXTERNAL_GITLINK_CHECK_FAILED:%s", p.ID)
	}
	fields := strings.Fields(string(treeOut))
	if len(fields) < 3 || fields[0] != "160000" || fields[1] != "commit" {
		return fmt.Errorf("EXTERNAL_GITLINK_NOT_REGISTERED:%s", p.ID)
	}
	if !isGitRevision(revision) || !strings.EqualFold(fields[2], revision) {
		return fmt.Errorf("EXTERNAL_GITLINK_REVISION_MISMATCH:%s:expected=%s:actual=%s", p.ID, revision, fields[2])
	}
	return nil
}

func sourceWorktreeIsGit(root string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--show-toplevel")
	return cmd.Run() == nil
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
	readOnlyMetadataOperation := strings.EqualFold(strings.TrimSpace(mode), "probe") ||
		strings.EqualFold(strings.TrimSpace(mode), "describe")
	if readOnlyMetadataOperation && !sourceWorktreeIsGit(filepath.Join(repositoryRoot(), p.Root)) {
		if err := r.validateRegisteredPin(ctx, p); err != nil {
			return nil, err
		}
	} else {
		if err := r.validateMaterializedPin(ctx, p); err != nil {
			return nil, err
		}
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
	if _, err := os.Stat(root); err != nil && !readOnlyMetadataOperation {
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
			_, err := os.Stat(filepath.Clean(root))
			providers = append(providers, map[string]any{"id": p.ID, "owner": p.Owner, "kind": p.Kind, "root": root, "revision": r.revisions[p.ID], "operations": p.Operations, "sourcePresent": err == nil, "executionEnabled": externalEnabled(p.ID)})
		}
		raw, _ := json.Marshal(map[string]any{"state": "PASS", "control_plane": "N07", "provider_count": len(providers), "manifest": r.manifestPath, "providers": providers})
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-federation", Target: m.Source, Status: "ok", Metadata: map[string]string{"federation_json": string(raw)}}, nil
	}); err != nil {
		return err
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
