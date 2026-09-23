package octacore

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

type SARAControlPublisher struct {
    baseURL string
    token string
    client *http.Client
}

func NewSARAControlPublisher(baseURL, token string, timeout time.Duration) *SARAControlPublisher {
    if timeout <= 0 { timeout = 15 * time.Second }
    return &SARAControlPublisher{baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"), token: strings.TrimSpace(token), client: &http.Client{Timeout: timeout}}
}

func (p *SARAControlPublisher) Status() string {
    if p == nil || p.baseURL == "" || p.token == "" { return "UNCONFIGURED" }
    return "CONFIGURED"
}

func (p *SARAControlPublisher) Publish(ctx context.Context, event VagusEnvelope) error {
    if p == nil || p.Status() == "UNCONFIGURED" { return errors.New("VAGUS_CONTROL_UNCONFIGURED") }
    if ctx == nil { return errors.New("context is nil") }
    if err := event.Validate(); err != nil { return err }
    body, err := json.Marshal(event)
    if err != nil { return fmt.Errorf("encode Vagus envelope: %w", err) }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/vagus", bytes.NewReader(body))
    if err != nil { return err }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+p.token)
    req.Header.Set("X-Correlation-ID", event.CorrelationID)
    resp, err := p.client.Do(req)
    if err != nil { return fmt.Errorf("Vagus control transport: %w", err) }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 { return fmt.Errorf("Vagus control HTTP %d", resp.StatusCode) }
    return nil
}
