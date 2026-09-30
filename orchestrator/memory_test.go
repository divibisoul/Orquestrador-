package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/memory"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type fakeSemanticStore struct {
	records []memory.Record
	matches []memory.Match
}

func (s *fakeSemanticStore) Record(_ context.Context, item memory.Record) error {
	s.records = append(s.records, item)
	return nil
}

func (s *fakeSemanticStore) Search(_ context.Context, userID string, _ []float64, _ float64, _ int) ([]memory.Match, error) {
	out := make([]memory.Match, 0, len(s.matches))
	for _, item := range s.matches {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func TestMemoryCapabilitiesUseCanonicalOrchestratorBoundary(t *testing.T) {
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}

	store := &fakeSemanticStore{
		matches: []memory.Match{{
			ID: "m1", UserID: "user-1", SessionID: "session-1",
			Summary: "prior context", Similarity: 0.91, CreatedAt: time.Now().UTC(),
		}},
	}
	if err := e.SetMemoryStore(store); err != nil {
		t.Fatal(err)
	}

	vector := make([]float64, memory.EmbeddingDimensions)
	for i := range vector {
		vector[i] = float64(i%7) / 7
	}

	record := protocol.NewMessage("N05", "N07", "command", "memory.record@1.0.0", vector)
	record.Metadata["memory_user_id"] = "user-1"
	record.Metadata["memory_session_id"] = "session-1"
	record.Metadata["memory_summary"] = "remember this"
	record.Metadata["memory_tags_json"] = "[\"learning\",\"gemini\"]"
	recordResult, err := e.Submit(context.Background(), record)
	if err != nil {
		t.Fatal(err)
	}
	if recordResult.Status != "ok" || len(store.records) != 1 {
		t.Fatalf("memory record did not commit: result=%+v records=%d", recordResult, len(store.records))
	}
	if len(store.records[0].Embedding) != memory.EmbeddingDimensions {
		t.Fatalf("unexpected stored embedding dimensions: %d", len(store.records[0].Embedding))
	}

	search := protocol.NewMessage("N05", "N07", "command", "memory.search@1.0.0", vector)
	search.Metadata["memory_user_id"] = "user-1"
	search.Metadata["memory_similarity_threshold"] = "0.75"
	search.Metadata["memory_match_count"] = "5"
	searchResult, err := e.Submit(context.Background(), search)
	if err != nil {
		t.Fatal(err)
	}
	if searchResult.Status != "ok" {
		t.Fatalf("memory search failed: %+v", searchResult)
	}
	if searchResult.Metadata["match_count"] != "1" {
		t.Fatalf("expected one semantic match, metadata=%v", searchResult.Metadata)
	}
}
