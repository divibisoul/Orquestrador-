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
)

type Web3Storage struct {
	baseURL  string
	token    string
	client   *http.Client
	gateway  string
	maxBytes int64
}

func NewWeb3Storage(cfg Config) *Web3Storage {
	maxBytes := cfg.MaxUploadBytes
	if maxBytes <= 0 {
		maxBytes = 100 << 20
	}
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Web3Storage{
		baseURL:  strings.TrimRight(cfg.Web3StorageURL, "/"),
		token:    cfg.Web3StorageToken,
		gateway:  strings.TrimRight(cfg.IPFSGatewayURL, "/"),
		client:   &http.Client{Timeout: timeout},
		maxBytes: maxBytes,
	}
}

func (s *Web3Storage) Configured() bool {
	return strings.TrimSpace(s.baseURL) != "" && strings.TrimSpace(s.token) != ""
}

func (s *Web3Storage) Upload(ctx context.Context, body io.Reader, filename string) (string, int64, error) {
	if !s.Configured() {
		return "", 0, errors.New("Web3 Storage credentials are not configured")
	}
	if body == nil {
		return "", 0, errors.New("upload body is required")
	}
	if ctx == nil {
		return "", 0, errors.New("context is nil")
	}
	if s.maxBytes <= 0 {
		return "", 0, errors.New("invalid upload size limit")
	}

	buf := new(bytes.Buffer)
	n, err := io.CopyN(buf, body, s.maxBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", 0, err
	}
	if n == 0 {
		return "", 0, errors.New("empty upload")
	}
	if n > s.maxBytes {
		return "", n, fmt.Errorf("upload exceeds maximum size of %d bytes", s.maxBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/upload", bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", n, err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", n))
	if strings.TrimSpace(filename) != "" {
		req.Header.Set("X-Name", filename)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", n, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", n, fmt.Errorf("storage upload failed: %s", strings.TrimSpace(string(data)))
	}
	var result struct {
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(data, &result); err != nil || strings.TrimSpace(result.CID) == "" {
		return "", n, errors.New("storage upload response did not contain cid")
	}
	return strings.TrimSpace(result.CID), n, nil
}

func (s *Web3Storage) Status(ctx context.Context, cid string) (map[string]any, error) {
	if !s.Configured() {
		return nil, errors.New("Web3 Storage credentials are not configured")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	cid = strings.TrimSpace(cid)
	if cid == "" {
		return nil, errors.New("cid is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/status/"+cid, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("storage status failed: %s", strings.TrimSpace(string(data)))
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Web3Storage) ObjectURL(cid string) string {
	cid = strings.TrimSpace(cid)
	if cid == "" || strings.TrimSpace(s.gateway) == "" {
		return ""
	}
	return s.gateway + "/" + cid
}
