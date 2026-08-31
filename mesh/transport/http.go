package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	Token      string
}

type Envelope struct {
	EventID           string                 `json:"event_id"`
	EventType         string                 `json:"event_type"`
	TraceID           string                 `json:"trace_id"`
	Source            string                 `json:"source"`
	Target            string                 `json:"target"`
	Timestamp         int64                  `json:"timestamp"`
	CompatibleSystems []string               `json:"compatible_systems,omitempty"`
	Payload           map[string]interface{} `json:"payload,omitempty"`
}

func (c Client) Do(ctx context.Context, endpoint string, env Envelope) (Envelope, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return Envelope{}, errors.New("endpoint required")
	}
	body, err := json.Marshal(env)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal envelope: %w", err)
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Envelope{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Soul-Event-Type", env.EventType)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Envelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Envelope{}, fmt.Errorf("remote nucleus returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	var out Envelope
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return Envelope{}, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}
