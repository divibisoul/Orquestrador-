package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type SupabaseStore struct {
	baseURL       string
	serviceKey    string
	runsTable     string
	artifactTable string
	client        *http.Client
}

func NewSupabaseStore(cfg Config) *SupabaseStore {
	return &SupabaseStore{
		baseURL:       strings.TrimRight(cfg.SupabaseURL, "/"),
		serviceKey:    cfg.SupabaseServiceKey,
		runsTable:     cfg.SupabaseRunsTable,
		artifactTable: cfg.SupabaseArtifactsTable,
		client:        &http.Client{},
	}
}

func (s *SupabaseStore) Configured() bool {
	return strings.TrimSpace(s.baseURL) != "" && strings.TrimSpace(s.serviceKey) != ""
}

func (s *SupabaseStore) insert(ctx context.Context, table string, row map[string]any) error {
	if !s.Configured() {
		return errors.New("Supabase server credentials are not configured")
	}
	body, err := json.Marshal(row)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/rest/v1/"+url.PathEscape(table), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("Supabase insert failed: %s", strings.TrimSpace(string(data)))
	}
	return nil
}

func (s *SupabaseStore) RecordRun(ctx context.Context, row map[string]any) error {
	if row == nil {
		return errors.New("run row is required")
	}
	return s.insert(ctx, s.runsTable, row)
}

func (s *SupabaseStore) RecordArtifact(ctx context.Context, row map[string]any) error {
	if row == nil {
		return errors.New("artifact row is required")
	}
	return s.insert(ctx, s.artifactTable, row)
}
