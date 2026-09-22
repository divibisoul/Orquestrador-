package backend

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

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

// SARAProxy is the N07 -> SARA service boundary.
// It does not reproduce SARA internals; it transports requests/results.
type SARAProxy struct {
	BaseURL string
	Token   string
	Client  *http.Client
	Timeout time.Duration
}

func NewSARAProxy(cfg Config) *SARAProxy {
	timeout := cfg.SARARequestTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &SARAProxy{
		BaseURL: strings.TrimRight(strings.TrimSpace(cfg.SARAServiceURL), "/"),
		Token: cfg.SARAServiceToken,
		Client: &http.Client{Timeout: timeout},
		Timeout: timeout,
	}
}

func (p *SARAProxy) Configured() bool {
	return p != nil && p.BaseURL != "" && p.Token != ""
}

func (p *SARAProxy) request(ctx context.Context, method, path string, body any, out *map[string]any) error {
	if p == nil || !p.Configured() {
		return errors.New("SARA service is not configured")
	}
	if ctx == nil {
		return errors.New("context is nil")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode SARA request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.BaseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("SARA transport: %w", err)
	}
	defer resp.Body.Close()

	if resp.ContentLength > 2<<20 {
		return errors.New("SARA response exceeds 2 MiB")
	}
	limited := io.LimitReader(resp.Body, 2<<20)
	var payload map[string]any
	if err := json.NewDecoder(limited).Decode(&payload); err != nil {
		return fmt.Errorf("decode SARA response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if errValue, ok := payload["error"]; ok {
			return fmt.Errorf("SARA HTTP %d: %v", resp.StatusCode, errValue)
		}
		return fmt.Errorf("SARA HTTP %d", resp.StatusCode)
	}
	if out != nil {
		*out = payload
	}
	return nil
}

func (p *SARAProxy) Cycle(ctx context.Context, input, cycleID string) (map[string]any, error) {
	body := map[string]any{"input": input}
	if strings.TrimSpace(cycleID) != "" {
		body["cycle_id"] = strings.TrimSpace(cycleID)
	}
	var out map[string]any
	err := p.request(ctx, http.MethodPost, "/v1/cycle", body, &out)
	return out, err
}

func (p *SARAProxy) Audit(ctx context.Context, input string) (map[string]any, error) {
	var out map[string]any
	err := p.request(ctx, http.MethodPost, "/v1/audit", map[string]any{"input": input}, &out)
	return out, err
}

func (p *SARAProxy) Regenerate(ctx context.Context, input string) (map[string]any, error) {
	var out map[string]any
	err := p.request(ctx, http.MethodPost, "/v1/regenerate", map[string]any{"input": input}, &out)
	return out, err
}

func (p *SARAProxy) State(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := p.request(ctx, http.MethodGet, "/v1/state", nil, &out)
	return out, err
}

func (p *SARAProxy) Capabilities(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := p.request(ctx, http.MethodGet, "/v1/capabilities", nil, &out)
	return out, err
}

func (p *SARAProxy) Trace(ctx context.Context, cycleID string) (map[string]any, error) {
	cycleID = strings.TrimSpace(cycleID)
	if cycleID == "" {
		return nil, errors.New("cycle id is required")
	}
	var out map[string]any
	err := p.request(ctx, http.MethodGet, "/v1/trace/"+cycleID, nil, &out)
	return out, err
}

func RegisterSARAOperations(e *orchestrator.Engine, proxy *SARAProxy) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if proxy == nil {
		return errors.New("SARA proxy is required")
	}

	registrations := []struct {
		name string
		call func(context.Context, protocol.Message) (protocol.Result, error)
	}{
		{"sara.cycle@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			input := strings.TrimSpace(message.Metadata["sara_input"])
			cycleID := strings.TrimSpace(message.Metadata["sara_cycle_id"])
			if input == "" {
				return protocol.Result{}, errors.New("metadata.sara_input is required")
			}
			out, err := proxy.Cycle(ctx, input, cycleID)
			return saraResult(message, out, err)
		}},
		{"sara.audit@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			input := strings.TrimSpace(message.Metadata["sara_input"])
			if input == "" {
				return protocol.Result{}, errors.New("metadata.sara_input is required")
			}
			out, err := proxy.Audit(ctx, input)
			return saraResult(message, out, err)
		}},
		{"sara.regenerate@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			input := strings.TrimSpace(message.Metadata["sara_input"])
			if input == "" {
				return protocol.Result{}, errors.New("metadata.sara_input is required")
			}
			out, err := proxy.Regenerate(ctx, input)
			return saraResult(message, out, err)
		}},
		{"sara.state@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			out, err := proxy.State(ctx)
			return saraResult(message, out, err)
		}},
		{"sara.capabilities@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			out, err := proxy.Capabilities(ctx)
			return saraResult(message, out, err)
		}},
		{"sara.trace@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
			cycleID := strings.TrimSpace(message.Metadata["sara_cycle_id"])
			out, err := proxy.Trace(ctx, cycleID)
			return saraResult(message, out, err)
		}},
	}

	for _, registration := range registrations {
		if err := e.Register(registration.name, registration.call); err != nil {
			return err
		}
	}
	return nil
}

func saraResult(message protocol.Message, payload map[string]any, err error) (protocol.Result, error) {
	metadata := map[string]string{
		"sara_transport": "HTTP",
		"sara_result_json": "{}",
	}
	if payload != nil {
		raw, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return protocol.Result{}, marshalErr
		}
		metadata["sara_result_json"] = string(raw)
	}
	result := protocol.Result{
		TraceID: message.TraceID,
		CorrelationID: message.CorrelationID,
		Source: "N07.sara",
		Target: message.Source,
		Status: "ok",
		Metadata: metadata,
	}
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
	}
	return result, err
}
