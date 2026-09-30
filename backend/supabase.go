package backend

import (
	"strconv"
	"bytes"
	"context"
	"github.com/divibisoul/Orquestrador-/learning"
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


// RecordLearning persists a validated learning observation using the existing Supabase service-role channel.
func (s *SupabaseStore) RecordLearning(ctx context.Context, item learning.PersistedExperience) error {
	if strings.TrimSpace(s.learningTable) == "" { return errors.New("Supabase learning table is not configured") }
	row := map[string]any{"id":item.ID,"trace_id":item.TraceID,"correlation_id":item.CorrelationID,"source":item.Source,"target":item.Target,"capability":item.Capability,"event_type":item.EventType,"outcome":item.Outcome,"reward":item.Reward,"confidence":item.Confidence,"input":item.Input,"target_vector":item.TargetVector,"provenance":item.Provenance,"metadata":item.Metadata,"created_at":item.Timestamp.UTC()}
	return s.insert(ctx,s.learningTable,row)
}
func (s *SupabaseStore) LoadLearning(ctx context.Context, limit int) ([]learning.PersistedExperience,error) {
	if !s.Configured() { return nil, errors.New("Supabase server credentials are not configured") }
	if limit < 1 || limit > 10000 { limit = 5000 }
	u := s.baseURL + "/rest/v1/" + url.PathEscape(s.learningTable) + "?select=*&order=created_at.asc&limit=" + strconv.Itoa(limit)
	req,err := http.NewRequestWithContext(ctx,http.MethodGet,u,nil); if err != nil { return nil,err }
	req.Header.Set("apikey",s.serviceKey); req.Header.Set("Authorization","Bearer "+s.serviceKey)
	resp,err:=s.client.Do(req); if err != nil{return nil,err}; defer resp.Body.Close()
	if resp.StatusCode<200 || resp.StatusCode>=300 { data,_:=io.ReadAll(io.LimitReader(resp.Body,1<<20)); return nil,fmt.Errorf("Supabase learning load failed: %s",strings.TrimSpace(string(data))) }
	var rows []learning.PersistedExperience
	if err:=json.NewDecoder(io.LimitReader(resp.Body,10<<20)).Decode(&rows); err != nil{return nil,err}
	return rows,nil
}
