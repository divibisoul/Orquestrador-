package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/divibisoul/Orquestrador-/protocol"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	MultiAgentDescribeOperation = "multiagent.crews.describe@1.0.0"
	MultiAgentExecuteOperation  = "multiagent.crews.execute@1.0.0"
)
const crewaiRevision = "1133f16cab274b9863b36fdacca7766b30e549fd"
const metagptRevision = "11cdf466d042aece04fc6cfd13b28e1a70341b1f"

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
	var root, rev, en string
	switch provider {
	case "", "crewai":
		provider = "crewai"
		root = getEnv("SOUL_N07_CREWAI_ROOT", "integrations/external/crewai")
		rev = crewaiRevision
		en = "SOUL_N07_CREWAI_ENABLED"
	case "metagpt":
		root = getEnv("SOUL_N07_METAGPT_ROOT", "integrations/external/metagpt")
		rev = metagptRevision
		en = "SOUL_N07_METAGPT_ENABLED"
	default:
		return MultiAgentProviderEvidence{State: "FAIL", Code: "MULTIAGENT_PROVIDER_UNSUPPORTED", Provider: provider, Python: py}
	}
	enabled := evBool(os.Getenv(en))
	root = filepath.Clean(root)
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
		if p == "metagpt" {
			return multiAgentFail(m, "MULTIAGENT_NATIVE_OWNER_N06")
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
		b, _ := json.Marshal(map[string]any{"goal": q.Goal, "roles": q.Roles, "maxRounds": rounds})
		script := "scripts/crewai_runner.py"
		if p == "metagpt" {
			script = "scripts/metagpt_runner.py"
		}
		cctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(cctx, getEnv("SOUL_N07_MULTIAGENT_PYTHON", "python3"), script)
		cmd.Stdin = strings.NewReader(string(b))
		out, err := cmd.Output()
		if err != nil {
			code := "MULTIAGENT_PROCESS_FAILED"
			if errors.Is(cctx.Err(), context.DeadlineExceeded) {
				code = "MULTIAGENT_TIMEOUT"
			}
			return multiAgentFail(m, code)
		}
		var result map[string]any
		if err := json.Unmarshal(out, &result); err != nil {
			return multiAgentFail(m, "MULTIAGENT_INVALID_RUNNER_OUTPUT")
		}
		if result["state"] != "PASS" {
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
