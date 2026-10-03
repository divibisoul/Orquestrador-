package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

const (
	ExternalProviderStatusOperation = "external.provider.status@1.0.0"
	ExternalProviderInvokeOperation = "external.provider.invoke@1.0.0"
)

const (
	externalVLLMRevision     = "7dfe3338d5f15dd3ccb233326aa65fde46e427ad"
	externalSGLangRevision   = "65f759144d192671af5301568113e38686999871"
	externalSwarmClawRevision = "ed38ba5329c20e48c03b4a4028f4a76a1a75e2d1"
	externalLangfuseRevision = "f75c661dbe8c6b85523c81486b39e8403ac2c141"
)

type externalProviderRuntime struct {
	ID        string
	Revision  string
	BaseURL   string
	Token     string
	Mode      string
	LiveProbe string
}

func externalProviderRuntimeConfig(provider string) externalProviderRuntime {
	p := strings.ToLower(strings.TrimSpace(provider))
	switch p {
	case "vllm":
		return externalProviderRuntime{ID: p, Revision: externalVLLMRevision, BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_VLLM_URL")), "/"), Token: strings.TrimSpace(os.Getenv("SOUL_VLLM_TOKEN")), Mode: "openai-compatible-http", LiveProbe: "/v1/models"}
	case "sglang":
		return externalProviderRuntime{ID: p, Revision: externalSGLangRevision, BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_SGLANG_URL")), "/"), Token: strings.TrimSpace(os.Getenv("SOUL_SGLANG_TOKEN")), Mode: "openai-compatible-http", LiveProbe: "/v1/models"}
	case "swarmclaw":
		return externalProviderRuntime{ID: p, Revision: externalSwarmClawRevision, BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_SWARMCLAW_URL")), "/"), Token: strings.TrimSpace(os.Getenv("SOUL_SWARMCLAW_TOKEN")), Mode: "a2a-jsonrpc-http", LiveProbe: "/.well-known/agent-card.json"}
	case "langfuse":
		return externalProviderRuntime{ID: p, Revision: externalLangfuseRevision, BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_LANGFUSE_URL")), "/"), Token: strings.TrimSpace(os.Getenv("SOUL_LANGFUSE_TOKEN")), Mode: "health-only-http", LiveProbe: "/api/public/health"}
	default:
		return externalProviderRuntime{ID: p}
	}
}

func externalHTTP(ctx context.Context, cfg externalProviderRuntime, method, path string, body any) (int, []byte, error) {
	if ctx == nil {
		return 0, nil, errors.New("context is nil")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return 0, nil, errors.New("provider endpoint is not configured")
	}
	rawURL := cfg.BaseURL
	if strings.HasSuffix(rawURL, "/v1") && strings.HasPrefix(path, "/v1/") {
		rawURL += strings.TrimPrefix(path, "/v1")
	} else {
		rawURL += path
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, data, nil
}

func externalProviderStatus(ctx context.Context, provider string) (map[string]any, error) {
	cfg := externalProviderRuntimeConfig(provider)
	if cfg.ID == "" {
		return nil, errors.New("EXTERNAL_PROVIDER_UNSUPPORTED")
	}
	out := map[string]any{
		"provider":   cfg.ID,
		"revision":   cfg.Revision,
		"mode":       cfg.Mode,
		"configured": cfg.BaseURL != "",
		"live":       false,
	}
	if cfg.BaseURL == "" {
		out["state"] = "BLOCKED"
		out["code"] = "EXTERNAL_PROVIDER_ENDPOINT_NOT_CONFIGURED"
		return out, nil
	}
	status, body, err := externalHTTP(ctx, cfg, http.MethodGet, cfg.LiveProbe, nil)
	if err != nil {
		out["state"] = "BLOCKED"
		out["code"] = "EXTERNAL_PROVIDER_UNREACHABLE"
		out["detail"] = err.Error()
		return out, nil
	}
	out["http_status"] = status
	if status < 200 || status >= 300 {
		out["state"] = "BLOCKED"
		out["code"] = "EXTERNAL_PROVIDER_HEALTH_FAILED"
		out["detail"] = strings.TrimSpace(string(body))
		return out, nil
	}
	out["state"] = "LIVE"
	out["live"] = true
	if cfg.ID == "vllm" || cfg.ID == "sglang" {
		var payload map[string]any
		if json.Unmarshal(body, &payload) == nil {
			out["model_inventory"] = payload["data"]
		}
	}
	return out, nil
}

func externalOpenAIInvoke(ctx context.Context, cfg externalProviderRuntime, model, prompt string, messages []map[string]any, temperature string) (map[string]any, error) {
	if len(messages) == 0 {
		if strings.TrimSpace(prompt) == "" {
			return nil, errors.New("prompt is required")
		}
		messages = []map[string]any{{"role": "user", "content": prompt}}
	}
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("model is required")
	}
	request := map[string]any{
		"model":    model,
		"messages": messages,
	}
	if strings.TrimSpace(temperature) != "" {
		v, err := strconv.ParseFloat(strings.TrimSpace(temperature), 64)
		if err != nil {
			return nil, errors.New("temperature must be numeric")
		}
		request["temperature"] = v
	}
	status, body, err := externalHTTP(ctx, cfg, http.MethodPost, "/v1/chat/completions", request)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("provider returned HTTP %d: %s", status, strings.TrimSpace(string(body)))
	}
	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("provider returned invalid JSON: %w", err)
	}
	choices, ok := response["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil, errors.New("provider returned no choices")
	}
	response["provider"] = cfg.ID
	response["providerRevision"] = cfg.Revision
	return response, nil
}

func externalSwarmClawInvoke(ctx context.Context, cfg externalProviderRuntime, goal, agentID string, timeout time.Duration) (map[string]any, error) {
	if strings.TrimSpace(goal) == "" {
		return nil, errors.New("goal is required")
	}
	if strings.TrimSpace(agentID) == "" {
		return nil, errors.New("SOUL_SWARMCLAW_AGENT_ID is required")
	}
	rpcID := protocol.NewTraceID()
	requester := protocol.NewTraceID()
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      rpcID,
		"method":  "executeTask",
		"params": map[string]any{
			"taskId":   rpcID,
			"taskName": "SOUL N07 delegated task",
			"message":  goal,
			"agentId":  agentID,
		},
	}
	status, raw, err := externalHTTP(ctx, cfg, http.MethodPost, "/api/a2a", body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("SwarmClaw A2A returned HTTP %d: %s", status, strings.TrimSpace(string(raw)))
	}
	var rpc map[string]any
	if err := json.Unmarshal(raw, &rpc); err != nil {
		return nil, fmt.Errorf("SwarmClaw returned invalid JSON: %w", err)
	}
	if rpc["error"] != nil {
		return nil, fmt.Errorf("SwarmClaw A2A error: %v", rpc["error"])
	}
	result, ok := rpc["result"].(map[string]any)
	if !ok {
		return nil, errors.New("SwarmClaw A2A response missing result")
	}
	taskID, _ := result["taskId"].(string)
	if taskID == "" {
		return nil, errors.New("SwarmClaw A2A response missing taskId")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pollBody := map[string]any{
			"jsonrpc": "2.0",
			"id":      protocol.NewTraceID(),
			"method":  "getStatus",
			"params":  map[string]any{"taskId": taskID},
		}
		pollStatus, pollRaw, pollErr := externalHTTP(ctx, cfg, http.MethodPost, "/api/a2a", pollBody)
		if pollErr != nil {
			return nil, pollErr
		}
		if pollStatus < 200 || pollStatus >= 300 {
			return nil, fmt.Errorf("SwarmClaw status returned HTTP %d: %s", pollStatus, strings.TrimSpace(string(pollRaw)))
		}
		var pollRPC map[string]any
		if err := json.Unmarshal(pollRaw, &pollRPC); err != nil {
			return nil, fmt.Errorf("SwarmClaw status invalid JSON: %w", err)
		}
		if pollRPC["error"] != nil {
			return nil, fmt.Errorf("SwarmClaw status error: %v", pollRPC["error"])
		}
		if pollResult, ok := pollRPC["result"].(map[string]any); ok {
			state, _ := pollResult["status"].(string)
			switch state {
			case "completed":
				pollResult["provider"] = cfg.ID
				pollResult["providerRevision"] = cfg.Revision
				pollResult["requesterId"] = requester
				return pollResult, nil
			case "failed", "cancelled":
				return nil, fmt.Errorf("SwarmClaw task %s: %s", taskID, state)
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(750 * time.Millisecond):
		}
	}
	return nil, errors.New("SWARMCLAW_TASK_TIMEOUT")
}

func RegisterExternalProviderRuntimeOperations(e *Engine) error {
	if e == nil {
		return errors.New("engine is nil")
	}
	if err := e.Register(ExternalProviderStatusOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		provider := strings.TrimSpace(m.Metadata["provider"])
		status, err := externalProviderStatus(ctx, provider)
		if err != nil {
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-runtime", Target: m.Source, Status: "error", Error: err.Error()}, err
		}
		raw, _ := json.Marshal(status)
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-runtime", Target: m.Source, Status: "ok", Metadata: map[string]string{
			"provider_status_json": string(raw),
			"evidence_state":       fmt.Sprint(status["state"]),
		}}, nil
	}); err != nil {
		return err
	}

	return e.Register(ExternalProviderInvokeOperation, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		provider := strings.ToLower(strings.TrimSpace(m.Metadata["provider"]))
		if provider == "" {
			return externalProviderFailure(m, "EXTERNAL_PROVIDER_REQUIRED")
		}
		switch provider {
		case "vllm", "sglang":
			cfg := externalProviderRuntimeConfig(provider)
			if cfg.BaseURL == "" {
				return externalProviderFailure(m, "EXTERNAL_PROVIDER_ENDPOINT_NOT_CONFIGURED")
			}
			messages := []map[string]any{}
			if raw := strings.TrimSpace(m.Metadata["messages_json"]); raw != "" {
				if err := json.Unmarshal([]byte(raw), &messages); err != nil {
					return externalProviderFailure(m, "MESSAGES_JSON_INVALID")
				}
			}
			model := strings.TrimSpace(m.Metadata["model"])
			if model == "" {
				if provider == "vllm" {
					model = strings.TrimSpace(os.Getenv("SOUL_VLLM_MODEL"))
				} else {
					model = strings.TrimSpace(os.Getenv("SOUL_SGLANG_MODEL"))
				}
			}
			response, err := externalOpenAIInvoke(ctx, cfg, model, strings.TrimSpace(m.Metadata["prompt"]), messages, m.Metadata["temperature"])
			if err != nil {
				return externalProviderFailure(m, "EXTERNAL_PROVIDER_INVOKE_FAILED:"+err.Error())
			}
			raw, _ := json.Marshal(response)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-runtime", Target: m.Source, Status: "ok", Metadata: map[string]string{
				"provider":       provider,
				"providerRevision": cfg.Revision,
				"response_json": string(raw),
				"evidence_state": "LIVE",
			}}, nil

		case "swarmclaw":
			cfg := externalProviderRuntimeConfig(provider)
			agentID := strings.TrimSpace(m.Metadata["agent_id"])
			if agentID == "" {
				agentID = strings.TrimSpace(os.Getenv("SOUL_SWARMCLAW_AGENT_ID"))
			}
			timeout := 2 * time.Minute
			if raw := strings.TrimSpace(m.Metadata["timeout_seconds"]); raw != "" {
				if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
					timeout = time.Duration(seconds) * time.Second
				}
			}
			result, err := externalSwarmClawInvoke(ctx, cfg, strings.TrimSpace(m.Metadata["goal"]), agentID, timeout)
			if err != nil {
				return externalProviderFailure(m, "SWARMCLAW_INVOKE_FAILED:"+err.Error())
			}
			raw, _ := json.Marshal(result)
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-runtime", Target: m.Source, Status: "ok", Metadata: map[string]string{
				"provider": provider,
				"providerRevision": cfg.Revision,
				"result_json": string(raw),
				"evidence_state": "LIVE",
			}}, nil
		default:
			return externalProviderFailure(m, "EXTERNAL_PROVIDER_ADAPTER_NOT_IMPLEMENTED:"+provider)
		}
	})
}

func externalProviderFailure(m protocol.Message, code string) (protocol.Result, error) {
	err := errors.New(code)
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.external-runtime", Target: m.Source, Status: "error", Error: code}, err
}
