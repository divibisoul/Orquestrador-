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
  "time"

  "github.com/divibisoul/Orquestrador-/memory"
)

type SupabaseSemanticMemoryStore struct {
  baseURL string
  serviceKey string
  client *http.Client
}

func NewSemanticMemoryStoreFromEnv() *SupabaseSemanticMemoryStore {
  return &SupabaseSemanticMemoryStore{
    baseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/"),
    serviceKey: strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY")),
    client: &http.Client{},
  }
}

func (s *SupabaseSemanticMemoryStore) Configured() bool {
  return s != nil && s.baseURL != "" && s.serviceKey != ""
}

func (s *SupabaseSemanticMemoryStore) headers(req *http.Request) {
  req.Header.Set("apikey", s.serviceKey)
  req.Header.Set("Authorization", "Bearer "+s.serviceKey)
  req.Header.Set("Content-Type", "application/json")
}

func vectorLiteral(values []float64) (string, error) {
  if err := memory.ValidateEmbedding(values); err != nil {
    return "", err
  }
  parts := make([]string, len(values))
  for i, value := range values {
    parts[i] = strconv.FormatFloat(value, 'g', -1, 64)
  }
  return "["+strings.Join(parts, ",")+"]", nil
}

func (s *SupabaseSemanticMemoryStore) Record(ctx context.Context, item memory.Record) error {
  if !s.Configured() { return errors.New("Supabase semantic memory store is not configured") }
  if err := memory.ValidateRecord(item); err != nil { return err }
  embedding, err := vectorLiteral(item.Embedding)
  if err != nil { return err }
  createdAt := item.CreatedAt.UTC()
  if createdAt.IsZero() { createdAt = time.Now().UTC() }
  body, err := json.Marshal(map[string]any{
    "user_id": item.UserID,
    "session_id": item.SessionID,
    "resumo": item.Summary,
    "embedding": embedding,
    "tags": append([]string(nil), item.Tags...),
    "criado_em": createdAt,
  })
  if err != nil { return err }
  req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/rest/v1/memories", bytes.NewReader(body))
  if err != nil { return err }
  s.headers(req)
  req.Header.Set("Prefer", "return=minimal")
  resp, err := s.client.Do(req)
  if err != nil { return err }
  defer resp.Body.Close()
  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
    return fmt.Errorf("Supabase memory insert failed: %s", strings.TrimSpace(string(data)))
  }
  return nil
}

func (s *SupabaseSemanticMemoryStore) Search(ctx context.Context, userID string, embedding []float64, threshold float64, limit int) ([]memory.Match, error) {
  if !s.Configured() { return nil, errors.New("Supabase semantic memory store is not configured") }
  if err := memory.ValidateSearch(userID, embedding, threshold, limit); err != nil { return nil, err }
  queryEmbedding, err := vectorLiteral(embedding)
  if err != nil { return nil, err }
  body, err := json.Marshal(map[string]any{
    "query_embedding": queryEmbedding,
    "match_user_id": userID,
    "match_threshold": threshold,
    "match_count": limit,
  })
  if err != nil { return nil, err }
  req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/rest/v1/rpc/match_memories", bytes.NewReader(body))
  if err != nil { return nil, err }
  s.headers(req)
  resp, err := s.client.Do(req)
  if err != nil { return nil, err }
  defer resp.Body.Close()
  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
    return nil, fmt.Errorf("Supabase memory search failed: %s", strings.TrimSpace(string(data)))
  }
  var matches []memory.Match
  if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&matches); err != nil {
    return nil, fmt.Errorf("Supabase memory search decode failed: %w", err)
  }
  return matches, nil
}

var _ memory.Store = (*SupabaseSemanticMemoryStore)(nil)

var _ = url.PathEscape
