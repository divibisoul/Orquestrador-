package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/divibisoul/Orquestrador-/protocol"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

type ExternalAdapterRegistry struct {
	manifest     ExternalAdapterManifest
	byID         map[string]ExternalAdapterSpec
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
	return ExternalAdapterRegistry{manifest: m, byID: by, manifestPath: path}, nil
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
func (r ExternalAdapterRegistry) run(ctx context.Context, provider, mode, operation string, metadata map[string]string) (map[string]any, error) {
	p, ok := r.byID[strings.ToLower(strings.TrimSpace(provider))]
	if !ok {
		return nil, fmt.Errorf("EXTERNAL_PROVIDER_UNKNOWN:%s", provider)
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
	req := map[string]any{"provider": provider, "mode": mode, "operation": operation, "root": root, "metadata": metadata}
	for k, v := range metadata {
		req[k] = v
	}
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
			_, err := os.Stat(filepath.Clean(p.Root))
			providers = append(providers, map[string]any{"id": p.ID, "owner": p.Owner, "kind": p.Kind, "root": p.Root, "operations": p.Operations, "sourcePresent": err == nil, "executionEnabled": externalEnabled(p.ID)})
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
		return externalAdapterResult(m, result), nil
	}); err != nil {
		return err
	}
	for id := range r.byID {
		provider := id
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
			return externalAdapterResult(m, result), nil
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
	status := "ok"
	if strings.EqualFold(state, "BLOCKED") || strings.EqualFold(state, "FAIL") {
		status = "error"
	}
	provider, _ := result["provider"].(string)
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-federation", Target: m.Source, Status: status, Metadata: map[string]string{"provider": provider, "external_result_json": string(raw)}}
}
