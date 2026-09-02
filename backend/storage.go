package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultStorachaTimeout = 30 * time.Second

const (
	storageNotConfiguredCode = "STORAGE_NOT_CONFIGURED"
	missingUCANReason        = "missing_ucan"
	storageCureSuggestion    = "Set WEB3_STORAGE_TOKEN (UCAN) and STORACHA_SPACE to enable real uploads. Obtain an authorized UCAN/Space from Storacha."
)

var cidPattern = regexp.MustCompile(`^(?:b[a-z2-7][a-z0-9]{20,}|Qm[1-9A-HJ-NP-Za-km-z]{40,}|bafk[a-z0-9]{20,})$`)
var cidFinder = regexp.MustCompile(`(?:https?://[^/]+/ipfs/)?(b[a-z2-7][a-z0-9]{20,}|Qm[1-9A-HJ-NP-Za-km-z]{40,}|bafk[a-z0-9]{20,})`)

type Web3Storage struct {
	baseURL         string
	token           string
	client          *http.Client
	gateway         string
	mode            string
	storachaBin     string
	storachaSpace   string
	storachaDataDir string
	maxBytes        int64
}

func NewWeb3Storage(cfg Config) *Web3Storage {
	legacyURL := strings.TrimRight(cfg.Web3StorageURL, "/")
	legacyGateway := strings.TrimRight(cfg.IPFSGatewayURL, "/")
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("STORACHA_MODE")))
	if mode == "" {
		mode = "auto"
	}
	maxBytes := cfg.MaxUploadBytes
	if maxBytes <= 0 {
		maxBytes = 100 << 20
	}
	if raw := strings.TrimSpace(os.Getenv("N07_MAX_UPLOAD_BYTES")); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			maxBytes = parsed
		}
	}
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = defaultStorachaTimeout
	}
	if raw := strings.TrimSpace(os.Getenv("STORACHA_TIMEOUT")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	bin := strings.TrimSpace(os.Getenv("STORACHA_GUPPY_BIN"))
	if bin == "" {
		bin = "guppy"
	}
	space := strings.TrimSpace(os.Getenv("STORACHA_SPACE"))
	dataDir := strings.TrimSpace(os.Getenv("STORACHA_DATA_DIR"))
	token := strings.TrimSpace(cfg.Web3StorageToken)
	return &Web3Storage{
		baseURL:         legacyURL,
		token:           token,
		gateway:         legacyGateway,
		mode:            mode,
		storachaBin:     bin,
		storachaSpace:   space,
		storachaDataDir: dataDir,
		maxBytes:        maxBytes,
		client:          &http.Client{Timeout: timeout},
	}
}

func (s *Web3Storage) useStoracha() bool {
	modern := strings.TrimSpace(s.storachaSpace) != "" && strings.TrimSpace(s.storachaBin) != ""
	switch s.mode {
	case "storacha", "modern":
		return true
	case "legacy", "web3.storage":
		return false
	default:
		return modern
	}
}

func (s *Web3Storage) modernConfigured() bool {
	return strings.TrimSpace(s.storachaSpace) != "" && strings.TrimSpace(s.storachaBin) != "" && strings.TrimSpace(s.token) != ""
}

func (s *Web3Storage) legacyConfigured() bool {
	return strings.TrimSpace(s.baseURL) != "" && strings.TrimSpace(s.token) != ""
}

func (s *Web3Storage) Configured() bool {
	if s == nil {
		return false
	}
	if s.useStoracha() {
		return s.modernConfigured()
	}
	return s.legacyConfigured()
}

func (s *Web3Storage) Status() map[string]any {
	if s == nil {
		return map[string]any{"status": "degraded", "reason": missingUCANReason, "code": storageNotConfiguredCode, "cureSuggestion": storageCureSuggestion}
	}
	if s.Configured() {
		return map[string]any{"status": "ready", "reason": "configured", "code": "STORAGE_READY", "cureSuggestion": ""}
	}
	return map[string]any{"status": "degraded", "reason": missingUCANReason, "code": storageNotConfiguredCode, "cureSuggestion": storageCureSuggestion}
}

func storageNotConfiguredError() error {
	return errors.New(storageNotConfiguredCode)
}

func (s *Web3Storage) Upload(ctx context.Context, body io.Reader, filename string) (string, int64, error) {
	if body == nil {
		return "", 0, errors.New("upload body is required")
	}
	if ctx == nil {
		return "", 0, errors.New("context is nil")
	}
	if !s.Configured() {
		return "", 0, storageNotConfiguredError()
	}
	if s.useStoracha() {
		return s.uploadStoracha(ctx, body, filename)
	}
	return s.uploadLegacy(ctx, body, filename)
}

func (s *Web3Storage) uploadLegacy(ctx context.Context, body io.Reader, filename string) (string, int64, error) {
	buf := new(bytes.Buffer)
	n, err := copyLimited(buf, body, s.maxBytes)
	if err != nil {
		return "", n, err
	}
	if n == 0 {
		return "", 0, errors.New("empty upload")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/upload", bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/octet-stream")
	if strings.TrimSpace(filename) != "" {
		req.Header.Set("X-Name", filename)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", n, fmt.Errorf("storage upload failed: %s", strings.TrimSpace(string(data)))
	}
	var result struct {
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(data, &result); err != nil || !validCID(result.CID) {
		return "", n, errors.New("storage upload response did not contain a valid cid")
	}
	return strings.TrimSpace(result.CID), n, nil
}

func (s *Web3Storage) uploadStoracha(ctx context.Context, body io.Reader, filename string) (string, int64, error) {
	tmpDir, err := os.MkdirTemp("", "n07-storacha-upload-")
	if err != nil {
		return "", 0, err
	}
	defer os.RemoveAll(tmpDir)

	name := filepath.Base(strings.TrimSpace(filename))
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = "upload.bin"
	}
	path := filepath.Join(tmpDir, name)
	tmp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", 0, fmt.Errorf("create upload staging file: %w", err)
	}
	written, copyErr := copyLimited(tmp, body, s.maxBytes)
	closeErr := tmp.Close()
	if copyErr != nil {
		return "", written, copyErr
	}
	if closeErr != nil {
		return "", written, closeErr
	}
	if written == 0 {
		return "", 0, errors.New("empty upload")
	}

	addArgs := []string{"upload", "source", "add", s.storachaSpace, path}
	uploadArgs := []string{"upload", s.storachaSpace}
	if s.storachaDataDir != "" {
		addArgs = append([]string{"--data-dir", s.storachaDataDir}, addArgs...)
		uploadArgs = append([]string{"--data-dir", s.storachaDataDir}, uploadArgs...)
	}
	addOutput, err := runCommand(ctx, s.storachaBin, addArgs...)
	if err != nil {
		return "", written, fmt.Errorf("Storacha source registration failed: %w", err)
	}
	uploadOutput, err := runCommand(ctx, s.storachaBin, uploadArgs...)
	if err != nil {
		return "", written, fmt.Errorf("Storacha upload failed: %w", err)
	}
	cid := firstCID(uploadOutput)
	if cid == "" {
		cid = firstCID(addOutput)
	}
	if !validCID(cid) {
		return "", written, errors.New("Storacha upload completed without a valid root cid")
	}
	return cid, written, nil
}

func (s *Web3Storage) StatusForCID(ctx context.Context, cid string) (map[string]any, error) {
	cid = strings.TrimSpace(cid)
	if !validCID(cid) {
		return nil, errors.New("invalid cid")
	}
	if !s.Configured() {
		return nil, storageNotConfiguredError()
	}
	if s.useStoracha() {
		return s.statusStoracha(ctx, cid)
	}
	return s.statusLegacy(ctx, cid)
}

func (s *Web3Storage) statusLegacy(ctx context.Context, cid string) (map[string]any, error) {
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

func (s *Web3Storage) statusStoracha(ctx context.Context, cid string) (map[string]any, error) {
	url := s.modernObjectURL(cid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Storacha gateway status failed: %s", resp.Status)
	}
	return map[string]any{"cid": cid, "available": true, "status": "reachable", "http_status": resp.StatusCode, "gateway": url}, nil
}

func (s *Web3Storage) ObjectURL(cid string) string {
	cid = strings.TrimSpace(cid)
	if !validCID(cid) || !s.Configured() {
		return ""
	}
	if s.useStoracha() {
		return s.modernObjectURL(cid)
	}
	if s.gateway == "" {
		return ""
	}
	return s.gateway + "/" + cid
}

func (s *Web3Storage) modernObjectURL(cid string) string {
	gateway := strings.TrimRight(strings.TrimSpace(os.Getenv("STORACHA_GATEWAY_URL")), "/")
	if gateway == "" {
		gateway = "https://storacha.link/ipfs"
	}
	return gateway + "/" + cid
}

func validCID(cid string) bool {
	cid = strings.TrimSpace(cid)
	return cid != "" && len(cid) <= 256 && cidPattern.MatchString(cid)
}

func firstCID(value string) string {
	match := cidFinder.FindStringSubmatch(value)
	if len(match) != 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func copyLimited(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	if maxBytes <= 0 {
		maxBytes = 100 << 20
	}
	limited := io.LimitReader(src, maxBytes+1)
	n, err := io.Copy(dst, limited)
	if err != nil {
		return n, err
	}
	if n > maxBytes {
		return n, fmt.Errorf("upload exceeds maximum size of %d bytes", maxBytes)
	}
	return n, nil
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			detail = err.Error()
		}
		return "", fmt.Errorf("%s: %s", name, detail)
	}
	return string(output), nil
}
