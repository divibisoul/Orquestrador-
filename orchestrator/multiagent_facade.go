package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/divibisoul/Orquestrador-/protocol"
	"os"
	"strconv"
	"path/filepath"
	"strings"
)

const (
	MultiAgentDescribeOperation = "multiagent.crews.describe@1.0.0"
	MultiAgentExecuteOperation  = "multiagent.crews.execute@1.0.0"
)

type MultiAgentFacadeRequest struct {
	Provider  string   `json:"provider,omitempty"`
	Goal      string   `json:"goal"`
	Roles     []string `json:"roles,omitempty"`
	MaxRounds int      `json:"maxRounds,omitempty"`
}
type MultiAgentProviderEvidence struct {
	State              string `json:"state"`
	Code               string `json:"code,omitempty"`
	Provider           string `json:"provider"`
	Revision           string `json:"revision"`
	Root               string `json:"root"`
	Python             string `json:"python"`
	Enabled            bool   `json:"enabled"`
	SourcePresent      bool   `json:"sourcePresent"`
	CredentialsPresent bool   `json:"credentialsPresent"`
	Primary            bool   `json:"primary"`
}

func evBool(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes"
}
func providerEvidence(provider string) MultiAgentProviderEvidence {
	provider = strings.ToLower(strings.TrimSpace(provider))
	py := getEnv("SOUL_N07_MULTIAGENT_PYTHON", "python3")
	if provider == "" {
		provider = "crewai"
	}
	if provider != "crewai" && provider != "metagpt" {
		return MultiAgentProviderEvidence{State: "FAIL", Code: "MULTIAGENT_PROVIDER_UNSUPPORTED", Provider: provider, Python: py}
	}
	registry, err := NewExternalAdapterRegistry()
	if err != nil {
		return MultiAgentProviderEvidence{State: "FAIL", Code: "MULTIAGENT_EXTERNAL_REGISTRY_UNAVAILABLE", Provider: provider, Python: py}
	}
	spec, ok := registry.byID[provider]
	if !ok {
		return MultiAgentProviderEvidence{State: "FAIL", Code: "MULTIAGENT_PROVIDER_NOT_REGISTERED", Provider: provider, Python: py}
	}
	root := spec.Root
	if !filepath.IsAbs(root) {
		root = filepath.Join(repositoryRoot(), root)
	}
	root = filepath.Clean(root)
	rev := registry.revisions[provider]
	en := "SOUL_N07_CREWAI_ENABLED"
	if provider == "metagpt" {
		en = "SOUL_N07_METAGPT_ENABLED"
	}
	enabled := evBool(os.Getenv(en))
	source := fileExists(root) || fileExists(filepath.Join(root, "README.md"))
	creds := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""
	out := MultiAgentProviderEvidence{State: "DEGRADED", Provider: provider, Revision: rev, Root: root, Python: py, Enabled: enabled, SourcePresent: source, CredentialsPresent: creds, Primary: false}
	if !enabled {
		out.Code = "MULTIAGENT_ADAPTER_DISABLED"
		return out
	}
	if !source {
		out.Code = "MULTIAGENT_SOURCE_NOT_AVAILABLE"
		return out
	}
	if !creds {
		out.Code = "MULTIAGENT_LLM_CREDENTIALS_NOT_AVAILABLE"
		return out
	}
	out.Code = "MULTIAGENT_EXECUTION_NOT_YET_PROVEN"
	return out
}
func getEnv(n, f string) string {
	if v := strings.TrimSpace(os.Getenv(n)); v != "" {
		return v
	}
	return f
}
func fileExists(p string) bool { _, e := os.Stat(p); return e == nil }

func RegisterMultiAgentFacadeOperations(e *Engine) error {
	if e == nil {
		return errors.New("engine is nil")
	}
	if err := e.Register(MultiAgentDescribeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		raw, _ := json.Marshal(map[string]any{"native_owner": "N07", "provider_selection": "explicit", "providers": map[string]any{"crewai": providerEvidence("crewai"), "metagpt": providerEvidence("metagpt")}, "single_facade": true, "execution_boundary": "explicit provider adapter via canonical N07 control plane"})
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.multiagent", Target: m.Source, Status: "ok", Metadata: map[string]string{"facade_json": string(raw)}}, nil
	}); err != nil {
		return err
	}
	return e.Register(MultiAgentExecuteOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		q := multiAgentRequestFromMessage(m)
		if q.Goal == "" && q.Provider == "" {
			return multiAgentFail(m, "MULTIAGENT_REQUEST_INVALID")
		}
		p := strings.ToLower(strings.TrimSpace(q.Provider))
		if p == "" {
			p = "crewai"
		}
		registry, regErr := NewExternalAdapterRegistry()
		if regErr != nil {
			return multiAgentFail(m, "MULTIAGENT_EXTERNAL_REGISTRY_UNAVAILABLE")
		}
		ev := providerEvidence(p)
		if ev.State == "FAIL" || !ev.Enabled || !ev.SourcePresent || !ev.CredentialsPresent {
			return multiAgentFail(m, ev.Code)
		}
		if strings.TrimSpace(q.Goal) == "" {
			return multiAgentFail(m, "MULTIAGENT_GOAL_REQUIRED")
		}
		rounds := q.MaxRounds
		if rounds < 1 {
			rounds = 3
		}
		if rounds > 20 {
			rounds = 20
		}
		rolesJSON, _ := json.Marshal(q.Roles)
		metadata := map[string]string{"goal": q.Goal, "roles": string(rolesJSON), "maxRounds": strconv.Itoa(rounds)}
		result, runErr := registry.run(ctx, p, "execute", "team.execute", metadata)
		if runErr != nil {
			return multiAgentFail(m, runErr.Error())
		}
		if !strings.EqualFold(fmt.Sprint(result["state"]), "PASS") {
			code, _ := result["code"].(string)
			if code == "" {
				code = "MULTIAGENT_PROVIDER_NOT_PASS"
			}
			return multiAgentFail(m, code)
		}
		result["provider"] = p
		result["providerRevision"] = ev.Revision
		result["correlationId"] = m.CorrelationID
		raw, _ := json.Marshal(result)
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.multiagent", Target: m.Source, Status: "ok", Metadata: map[string]string{"result_json": string(raw)}}, nil
	})
}
func multiAgentRequestFromMessage(m protocol.Message) MultiAgentFacadeRequest {
	q := MultiAgentFacadeRequest{}
	if m.Metadata == nil {
		return q
	}
	q.Provider = strings.TrimSpace(m.Metadata["provider"])
	q.Goal = strings.TrimSpace(m.Metadata["goal"])
	if raw := strings.TrimSpace(m.Metadata["roles"]); raw != "" {
		_ = json.Unmarshal([]byte(raw), &q.Roles)
	}
	if raw := strings.TrimSpace(m.Metadata["maxRounds"]); raw != "" {
		var n int
		err := json.Unmarshal([]byte(raw), &n)
		if err == nil {
			q.MaxRounds = n
		}
	}
	return q
}

func multiAgentFail(m protocol.Message, code string) (protocol.Result, error) {
	e := errors.New(code)
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.multiagent", Target: m.Source, Status: "error", Error: code}, e
}
