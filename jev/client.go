package jev

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
)

const (
	DefaultBaseURL = "https://api.typesafe.ai"
	DefaultModel   = "jev-latest"
	SystemOnePath = "/v1/systemone"
)

type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
	MaxRetries int
	Backoff    time.Duration
}

type Request struct {
	State     any                          `json:"state"`
	Model     string                       `json:"model"`
	Questions map[string]map[string]any    `json:"questions"`
}

type Response struct {
	Model   string                      `json:"model"`
	Answers map[string]map[string]any    `json:"answers"`
	Usage   map[string]any               `json:"usage,omitempty"`
}

func NewFromEnv() (*Client, error) {
	key := strings.TrimSpace(os.Getenv("JEV_API_KEY"))
	if key == "" {
		return nil, errors.New("JEV_API_KEY is not configured")
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("JEV_API_BASE_URL")), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	model := strings.TrimSpace(os.Getenv("JEV_MODEL"))
	if model == "" {
		model = DefaultModel
	}
	retries := envInt("JEV_MAX_RETRIES", 2)
	backoff := envDuration("JEV_RETRY_BACKOFF", 1*time.Second)
	return &Client{
		BaseURL: base,
		APIKey: key,
		Model: model,
		HTTPClient: &http.Client{Timeout: 20 * time.Second},
		MaxRetries: retries,
		Backoff: backoff,
	}, nil
}

func (c *Client) SystemOne(ctx context.Context, state any, questions map[string]map[string]any) (Response, error) {
	if c == nil || strings.TrimSpace(c.APIKey) == "" {
		return Response{}, errors.New("JEV_API_KEY is not configured")
	}
	if ctx == nil {
		return Response{}, errors.New("context is nil")
	}
	if state == nil {
		return Response{}, errors.New("state is required")
	}
	if len(questions) == 0 {
		return Response{}, errors.New("questions are required")
	}

	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = DefaultModel
	}
	body, err := json.Marshal(Request{State: state, Model: model, Questions: questions})
	if err != nil {
		return Response{}, fmt.Errorf("encode Jev request: %w", err)
	}

	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		return Response{}, errors.New("Jev base URL is not configured")
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	maxRetries := c.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	backoff := c.Backoff
	if backoff <= 0 {
		backoff = time.Second
	}

	for attempt := 0; ; attempt++ {
		out, status, retryAfter, err := c.doSystemOne(ctx, client, base, body)
		if err == nil {
			return out, nil
		}
		if attempt >= maxRetries || !retryableStatus(status) {
			return Response{}, err
		}

		delay := backoff * time.Duration(1<<attempt)
		if retryAfter > 0 && retryAfter > delay {
			delay = retryAfter
		}
		select {
		case <-ctx.Done():
			return Response{}, ctx.Err()
		case <-time.After(delay + retryJitter()):
		}
	}
}

func (c *Client) doSystemOne(ctx context.Context, client *http.Client, base string, body []byte) (Response, int, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+SystemOnePath, bytes.NewReader(body))
	if err != nil {
		return Response{}, 0, 0, fmt.Errorf("create Jev request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, 0, 0, fmt.Errorf("Jev request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, resp.StatusCode, 0, fmt.Errorf("read Jev response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, resp.StatusCode, retryAfterDuration(resp.Header.Get("Retry-After")), fmt.Errorf("Jev HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out Response
	if err := json.Unmarshal(raw, &out); err != nil {
		return Response{}, resp.StatusCode, 0, fmt.Errorf("decode Jev response: %w", err)
	}
	if strings.TrimSpace(out.Model) == "" {
		return Response{}, resp.StatusCode, 0, errors.New("Jev response missing model")
	}
	if out.Answers == nil {
		return Response{}, resp.StatusCode, 0, errors.New("Jev response missing answers")
	}
	return out, resp.StatusCode, 0, nil
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == 529
}

func retryAfterDuration(value string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func retryJitter() time.Duration {
	n := time.Now().UnixNano() % int64(250*time.Millisecond)
	if n < 0 {
		n = -n
	}
	return time.Duration(n)
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
