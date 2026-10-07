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
	"os"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/learning"
)

// SupabaseLearningStore is optional durable storage for learning experiences.
// It reads the same server-only Supabase credentials already used by N07.
type SupabaseLearningStore struct {
	baseURL    string
	serviceKey string
	table      string
	client     *http.Client
}

func NewLearningStoreFromEnv() *SupabaseLearningStore {
	return &SupabaseLearningStore{
		baseURL:    strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/"),
		serviceKey: strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY")),
		table:      envString("SUPABASE_LEARNING_TABLE", "n07_learning_experiences"),
		client:     &http.Client{},
	}
}

func (s *SupabaseLearningStore) Configured() bool {
	return s != nil && s.baseURL != "" && s.serviceKey != ""
}

func (s *SupabaseLearningStore) RecordLearning(ctx context.Context, exp learning.PersistedExperience) error {
	if !s.Configured() {
		return errors.New("Supabase learning store is not configured")
	}
	row := map[string]any{
		"trace_id": exp.TraceID, "correlation_id": exp.CorrelationID,
		"source": exp.Source, "target": exp.Target, "capability": exp.Capability,
		"event_type": exp.EventType, "outcome": exp.Outcome,
		"reward": exp.Reward, "confidence": exp.Confidence,
		"input": exp.Input, "target_vector": exp.TargetVector,
		"provenance": exp.Provenance, "metadata": exp.Metadata,
		"created_at": exp.Timestamp.UTC(),
	}
	return s.insert(ctx, row)
}

func (s *SupabaseLearningStore) insert(ctx context.Context, row map[string]any) error {
	body, err := json.Marshal(row)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.baseURL+"/rest/v1/"+url.PathEscape(s.table),
		bytes.NewReader(body))
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
		return fmt.Errorf("Supabase learning insert failed: %s", strings.TrimSpace(string(data)))
	}
	return nil
}

func (s *SupabaseLearningStore) LoadLearning(ctx context.Context, limit int) ([]learning.PersistedExperience, error) {
	if !s.Configured() {
		return nil, errors.New("Supabase learning store is not configured")
	}
	if limit <= 0 || limit > 10000 {
		limit = 5000
	}
	endpoint := s.baseURL + "/rest/v1/" + url.PathEscape(s.table) +
		"?select=id,trace_id,correlation_id,source,target,capability,event_type,outcome,reward,confidence,input,target_vector,provenance,metadata,created_at&order=created_at.asc&limit=" + strconv.Itoa(limit)
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
