package backend

import (
	"bytes"
	"github.com/divibisoul/Orquestrador-/learning"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SupabaseStore struct {
	baseURL       string
	serviceKey    string
	runsTable     string
	artifactTable  string
	learningTable  string
	client         *http.Client
}

func NewSupabaseStore(cfg Config) *SupabaseStore {
	return &SupabaseStore{
		baseURL:       strings.TrimRight(cfg.SupabaseURL, "/"),
		serviceKey:    cfg.SupabaseServiceKey,
		runsTable:     cfg.SupabaseRunsTable,
		artifactTable: cfg.SupabaseArtifactsTable,
		learningTable: cfg.SupabaseLearningTable,
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

func (s *SupabaseStore) RecordLearning(ctx context.Context, exp learning.PersistedExperience) error {
	row := map[string]any{
		"trace_id": exp.TraceID,
		"correlation_id": exp.CorrelationID,
		"source": exp.Source,
		"target": exp.Target,
		"capability": exp.Capability,
		"event_type": exp.EventType,
		"outcome": exp.Outcome,
		"reward": exp.Reward,
		"confidence": exp.Confidence,
		"input": exp.Input,
		"target_vector": exp.TargetVector,
		"provenance": exp.Provenance,
		"created_at": exp.Timestamp.UTC(),
		"metadata": exp.Metadata,
	}
	return s.insert(ctx, s.learningTable, row)
}

func (s *SupabaseStore) LoadLearning(ctx context.Context, limit int) ([]learning.PersistedExperience, error) {
	if !s.Configured() {
		return nil, errors.New("Supabase server credentials are not configured")
	}
	if limit <= 0 || limit > 10000 {
		limit = 5000
	}
	endpoint := s.baseURL + "/rest/v1/" + url.PathEscape(s.learningTable) +
		"?select=*&order=created_at.asc&limit=" + strconv.Itoa(limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("Supabase learning load failed: %s", strings.TrimSpace(string(data)))
	}
	var rows []learning.PersistedExperience
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&rows); err != nil {
		return nil, fmt.Errorf("Supabase learning decode failed: %w", err)
	}
	return rows, nil
}
