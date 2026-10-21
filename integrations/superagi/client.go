package superagi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	CreatePath string
	RunPath    string
}

type Config struct {
	BaseURL string
	APIKey  string
}

type CreateAgentRequest struct {
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Goals           []string       `json:"goal"`
	Instruction     []string       `json:"instruction,omitempty"`
	Tools           []int          `json:"tools"`
	Model           string         `json:"model"`
	AgentWorkflow   string         `json:"agent_workflow,omitempty"`
	PermissionType  string         `json:"permission_type,omitempty"`
	MaxIterations   int            `json:"max_iterations,omitempty"`
	Knowledge       *int           `json:"knowledge,omitempty"`
	LTMDB           string         `json:"LTM_DB,omitempty"`
}

type AgentResponse struct {
	AgentID     json.RawMessage `json:"agent_id"`
	ExecutionID json.RawMessage `json:"execution_id"`
	Raw         map[string]any  `json:"-"`
}

type RunAgentRequest struct {
	AgentID any            `json:"agent_id"`
	Input   map[string]any `json:"input,omitempty"`
}

func New(cfg Config) (*Client, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		return nil, errors.New("SUPERAGI_BLOCKED_INFRASTRUCTURE: base URL is not configured")
	}
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return nil, errors.New("SUPERAGI_BLOCKED_INFRASTRUCTURE: API key is not configured")
	}
	return &Client{
		BaseURL: base,
		APIKey: key,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		CreatePath: "/agents/create",
		RunPath: "/agents/run",
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload any) (map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("superagi API returned HTTP %d", resp.StatusCode)
	}
	return out, nil
}

func (c *Client) CreateAgent(ctx context.Context, req CreateAgentRequest) (map[string]any, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("agent name is required")
	}
	if len(req.Goals) == 0 {
		return nil, errors.New("at least one agent goal is required")
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, errors.New("agent model is required")
	}
	return c.doJSON(ctx, http.MethodPost, c.CreatePath, req)
}

func (c *Client) StartRun(ctx context.Context, req RunAgentRequest) (map[string]any, error) {
	if req.AgentID == nil {
		return nil, errors.New("agent_id is required")
	}
	return c.doJSON(ctx, http.MethodPost, c.RunPath, req)
}
