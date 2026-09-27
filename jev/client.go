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
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.typesafe.ai"
	DefaultModel = "jev-latest"
	SystemOnePath = "/v1/systemone"
)

type Client struct {
	BaseURL string
	APIKey string
	Model string
	HTTPClient *http.Client
}

type Request struct {
	State any `json:"state"`
	Model string `json:"model"`
	Questions map[string]map[string]any `json:"questions"`
}

type Response struct {
	Model string `json:"model"`
	Answers map[string]map[string]any `json:"answers"`
	Usage map[string]any `json:"usage,omitempty"`
}

func NewFromEnv() (*Client, error) {
	key := strings.TrimSpace(os.Getenv("JEV_API_KEY"))
	if key == "" { return nil, errors.New("JEV_API_KEY is not configured") }
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("JEV_API_BASE_URL")), "/")
	if base == "" { base = DefaultBaseURL }
	model := strings.TrimSpace(os.Getenv("JEV_MODEL"))
	if model == "" { model = DefaultModel }
	return &Client{BaseURL: base, APIKey: key, Model: model, HTTPClient: &http.Client{Timeout: 15*time.Second}}, nil
}

func (c *Client) SystemOne(ctx context.Context, state any, questions map[string]map[string]any) (Response, error) {
	if c == nil || strings.TrimSpace(c.APIKey) == "" { return Response{}, errors.New("JEV_API_KEY is not configured") }
	if ctx == nil { return Response{}, errors.New("context is nil") }
	if state == nil { return Response{}, errors.New("state is required") }
	if len(questions) == 0 { return Response{}, errors.New("questions are required") }
	model := strings.TrimSpace(c.Model); if model == "" { model = DefaultModel }
	body, err := json.Marshal(Request{State: state, Model: model, Questions: questions}); if err != nil { return Response{}, fmt.Errorf("encode Jev request: %w", err) }
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/"); if base == "" { return Response{}, errors.New("Jev base URL is not configured") }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+SystemOnePath, bytes.NewReader(body)); if err != nil { return Response{}, fmt.Errorf("create Jev request: %w", err) }
	req.Header.Set("Authorization", "Bearer "+c.APIKey); req.Header.Set("Content-Type", "application/json"); req.Header.Set("Accept", "application/json")
	client := c.HTTPClient; if client == nil { client = &http.Client{Timeout: 15*time.Second} }
	resp, err := client.Do(req); if err != nil { return Response{}, fmt.Errorf("Jev request failed: %w", err) }
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)); if err != nil { return Response{}, fmt.Errorf("read Jev response: %w", err) }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return Response{}, fmt.Errorf("Jev HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))) }
	var out Response; if err := json.Unmarshal(raw, &out); err != nil { return Response{}, fmt.Errorf("decode Jev response: %w", err) }
	if out.Answers == nil { return Response{}, errors.New("Jev response missing answers") }
	return out, nil
}
