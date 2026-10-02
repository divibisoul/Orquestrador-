package agentarsenal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Proxy struct {
	baseURL string
	client  *http.Client
}

func NewFromEnv() *Proxy {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_AGENT_ARSENAL_URL")), "/")
	timeout := 25 * time.Second
	if raw := strings.TrimSpace(os.Getenv("SOUL_AGENT_ARSENAL_TIMEOUT")); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			timeout = time.Duration(seconds) * time.Second
		}
	}
	return &Proxy{baseURL: baseURL, client: &http.Client{Timeout: timeout}}
}

func (p *Proxy) Configured() bool { return p != nil && p.baseURL != "" }

func (p *Proxy) do(ctx context.Context, method, endpointPath string, query url.Values, body any) (map[string]any, error) {
	if p == nil || !p.Configured() {
		return nil, errors.New("SOUL_AGENT_ARSENAL_URL is not configured")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	endpoint := p.baseURL + endpointPath
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var detail map[string]any
		if json.Unmarshal(data, &detail) == nil && detail["error"] != nil {
			return nil, fmt.Errorf("agent arsenal gateway: %v", detail["error"])
		}
		return nil, fmt.Errorf("agent arsenal gateway: http %d", resp.StatusCode)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("agent arsenal gateway returned invalid JSON: %w", err)
	}
	return result, nil
}

func (p *Proxy) Catalog(ctx context.Context, source, kind string, limit, offset int) (map[string]any, error) {
	query := url.Values{}
	if strings.TrimSpace(source) != "" { query.Set("source", strings.TrimSpace(source)) }
	if strings.TrimSpace(kind) != "" { query.Set("kind", strings.TrimSpace(kind)) }
	if limit > 0 { query.Set("limit", strconv.Itoa(limit)) }
	if offset > 0 { query.Set("offset", strconv.Itoa(offset)) }
	return p.do(ctx, http.MethodGet, "/v1/catalog", query, nil)
}

func (p *Proxy) Resolve(ctx context.Context, source, artifactPath string) (map[string]any, error) {
	return p.do(ctx, http.MethodPost, "/v1/resolve", nil, map[string]string{
		"source": strings.TrimSpace(source),
		"path":   strings.TrimSpace(artifactPath),
	})
}

func (p *Proxy) Swarm(ctx context.Context, task, strategy, priority string) (map[string]any, error) {
	return p.do(ctx, http.MethodPost, "/v1/swarm", nil, map[string]string{
		"task": strings.TrimSpace(task), "strategy": strings.TrimSpace(strategy), "priority": strings.TrimSpace(priority),
	})
}

func (p *Proxy) Spawn(ctx context.Context, agentType, name string) (map[string]any, error) {
	return p.do(ctx, http.MethodPost, "/v1/agent", nil, map[string]string{
		"type": strings.TrimSpace(agentType), "name": strings.TrimSpace(name),
	})
}
