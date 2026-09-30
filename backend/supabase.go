package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"io"
	"strconv"
	"net/http"
	"net/url"
	"strings"

	"github.com/divibisoul/Orquestrador-/learning"
	"github.com/divibisoul/Orquestrador-/memory"
)

type SupabaseStore struct {
	baseURL       string
	serviceKey    string
	runsTable     string
	artifactTable string
	learningTable string
	client        *http.Client
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


func (s *SupabaseStore) RecordLearning(ctx context.Context, item learning.PersistedExperience) error {
	if strings.TrimSpace(s.learningTable) == "" {
		return errors.New("Supabase learning table is not configured")
	}
	row := map[string]any{
		"id": item.ID, "trace_id": item.TraceID, "correlation_id": item.CorrelationID,
		"source": item.Source, "target": item.Target, "capability": item.Capability,
		"event_type": item.EventType, "outcome": item.Outcome,
		"reward": item.Reward, "confidence": item.Confidence,
		"input": item.Input, "target_vector": item.TargetVector,
		"provenance": item.Provenance, "metadata": item.Metadata, "created_at": item.Timestamp.UTC(),
	}
	return s.insert(ctx, s.learningTable, row)
}

func (s *SupabaseStore) LoadLearning(ctx context.Context, limit int) ([]learning.PersistedExperience, error) {
	if strings.TrimSpace(s.learningTable) == "" {
		return nil, errors.New("Supabase learning table is not configured")
	}
	if limit < 1 { limit = 1 }
	if limit > 50000 { limit = 50000 }
	u := s.baseURL + "/rest/v1/" + url.PathEscape(s.learningTable) + "?select=*&order=created_at.asc&limit=" + strconv.Itoa(limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil { return nil, err }
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	resp, err := s.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("Supabase learning load failed: %s", strings.TrimSpace(string(data)))
	}
	var rows []learning.PersistedExperience
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&rows); err != nil { return nil, err }
	return rows, nil
}

func embeddingLiteral(values []float64) (string, error) {
	if err := memory.ValidateEmbedding(values); err != nil { return "", err }
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.FormatFloat(v, 'g', -1, 64)
	}
	return "[" + strings.Join(parts, ",") + "]", nil
}

func (s *SupabaseStore) Record(ctx context.Context, item memory.Record) error {
	if err := memory.ValidateRecord(item); err != nil { return err }
	embedding, err := embeddingLiteral(item.Embedding)
	if err != nil { return err }
	row := map[string]any{
		"id": item.ID, "user_id": item.UserID, "session_id": item.SessionID,
		"resumo": item.Summary, "embedding": embedding, "tags": item.Tags, "criado_em": item.CreatedAt.UTC(),
	}
	return s.insert(ctx, "memories", row)
}

func (s *SupabaseStore) Search(ctx context.Context, userID string, embedding []float64, threshold float64, limit int) ([]memory.Match, error) {
	if !s.Configured() { return nil, errors.New("Supabase server credentials are not configured") }
	if err := memory.ValidateSearch(userID, embedding, threshold, limit); err != nil { return nil, err }
	literal, err := embeddingLiteral(embedding)
	if err != nil { return nil, err }
	body, err := json.Marshal(map[string]any{
		"query_embedding": literal, "match_user_id": userID,
		"match_threshold": threshold, "match_count": limit,
	})
	if err != nil { return nil, err }
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/rest/v1/rpc/match_memories", bytes.NewReader(body))
	if err != nil { return nil, err }
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("Supabase memory search failed: %s", strings.TrimSpace(string(data)))
	}
	var matches []memory.Match
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&matches); err != nil { return nil, err }
	return matches, nil
}
