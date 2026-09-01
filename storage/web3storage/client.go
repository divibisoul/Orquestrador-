package web3storage

import (
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

const DefaultAPIBaseURL = "https://api.web3.storage"

var ErrNotConfigured = errors.New("web3.storage is not configured")

type Config struct {
	APIBaseURL string
	Token      string
	Timeout    time.Duration
	MaxUpload  int64
}

func ConfigFromEnv() Config {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("WEB3_STORAGE_API_URL")), "/")
	if base == "" {
		base = DefaultAPIBaseURL
	}
	timeout := 60 * time.Second
	if raw := strings.TrimSpace(os.Getenv("WEB3_STORAGE_TIMEOUT")); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			timeout = d
		}
	}
	maxUpload := int64(100 << 20)
	if raw := strings.TrimSpace(os.Getenv("WEB3_STORAGE_MAX_UPLOAD_BYTES")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			maxUpload = n
		}
	}
	return Config{APIBaseURL: base, Token: strings.TrimSpace(os.Getenv("WEB3_STORAGE_TOKEN")), Timeout: timeout, MaxUpload: maxUpload}
}

func (c Config) validate() error {
	if strings.TrimSpace(c.APIBaseURL) == "" {
		return errors.New("web3.storage API base URL is empty")
	}
	if _, err := url.ParseRequestURI(c.APIBaseURL); err != nil {
		return fmt.Errorf("invalid web3.storage API base URL: %w", err)
	}
	if strings.TrimSpace(c.Token) == "" {
		return ErrNotConfigured
	}
	if c.Timeout <= 0 {
		return errors.New("web3.storage timeout must be positive")
	}
	if c.MaxUpload <= 0 {
		return errors.New("web3.storage max upload must be positive")
	}
	return nil
}

type Client struct {
	baseURL   string
	token     string
	http      *http.Client
	maxUpload int64
}

func New(cfg Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &Client{baseURL: strings.TrimRight(cfg.APIBaseURL, "/"), token: cfg.Token, http: &http.Client{Timeout: cfg.Timeout}, maxUpload: cfg.MaxUpload}, nil
}

func NewFromEnv() (*Client, error) { return New(ConfigFromEnv()) }

// Upload sends the raw file body to the documented POST /upload endpoint.
func (c *Client) Upload(ctx context.Context, name string, size int64, r io.Reader) (string, error) {
	if c == nil {
		return "", ErrNotConfigured
	}
	if size < 0 || size > c.maxUpload {
		return "", fmt.Errorf("web3.storage upload size %d exceeds limit %d", size, c.maxUpload)
	}
	if strings.TrimSpace(name) == "" {
		return "", errors.New("upload filename is required")
	}
	if r == nil {
		return "", errors.New("upload reader is required")
	}
	counted := &countingReader{r: io.LimitReader(r, size+1)}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/upload", counted)
	if err != nil {
		return "", err
	}
	req.ContentLength = size
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if counted.n != size {
		return "", fmt.Errorf("web3.storage upload source size mismatch: expected %d bytes, sent %d", size, counted.n)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", responseError(resp)
	}
	var out uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode upload response: %w", err)
	}
	if strings.TrimSpace(out.CID) == "" {
		return "", errors.New("web3.storage upload returned no CID")
	}
	return out.CID, nil
}

type countingReader struct {
	r io.Reader
	n int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.n += int64(n)
	return n, err
}

type uploadResponse struct {
	CID string `json:"cid"`
}
type responseErrorBody struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func responseError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	var e responseErrorBody
	if json.Unmarshal(data, &e) == nil {
		msg := strings.TrimSpace(e.Message)
		if msg == "" {
			msg = strings.TrimSpace(e.Error)
		}
		if msg != "" {
			return fmt.Errorf("web3.storage HTTP %d: %s", resp.StatusCode, msg)
		}
	}
	return fmt.Errorf("web3.storage HTTP %d", resp.StatusCode)
}

func (c *Client) Status(ctx context.Context, cid string) (map[string]any, error) {
	if c == nil {
		return nil, ErrNotConfigured
	}
	cid = strings.TrimSpace(cid)
	if cid == "" {
		return nil, errors.New("CID is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/status/"+url.PathEscape(cid), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, responseError(resp)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode status response: %w", err)
	}
	return out, nil
}

func (c *Client) List(ctx context.Context) ([]map[string]any, error) {
	if c == nil {
		return nil, ErrNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/user/uploads", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, responseError(resp)
	}
	var out []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode upload list: %w", err)
	}
	return out, nil
}

func (c *Client) Healthy(ctx context.Context) error { _, err := c.List(ctx); return err }

func GatewayURL(cid string) string {
	cid = strings.TrimSpace(cid)
	if cid == "" {
		return ""
	}
	return "https://dweb.link/ipfs/" + url.PathEscape(cid)
}
