package memory

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
)

const EmbeddingDimensions = 768

type Record struct {
	ID        string
	UserID    string
	SessionID string
	Summary   string
	Embedding []float64
	Tags      []string
	CreatedAt time.Time
}

type Match struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	SessionID  string `json:"session_id"`
	Summary    string `json:"summary"`
	Tags       []string `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	Similarity float64 `json:"similarity"`
}

type Store interface {
	Record(context.Context, Record) error
	Search(context.Context, string, []float64, float64, int) ([]Match, error)
}

func ValidateEmbedding(values []float64) error {
	if len(values) != EmbeddingDimensions {
		return errors.New("memory embedding must contain exactly 768 dimensions")
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("memory embedding contains non-finite value")
		}
	}
	return nil
}

func ValidateRecord(item Record) error {
	if strings.TrimSpace(item.UserID) == "" {
		return errors.New("memory user_id is required")
	}
	if strings.TrimSpace(item.SessionID) == "" {
		return errors.New("memory session_id is required")
	}
	if strings.TrimSpace(item.Summary) == "" {
		return errors.New("memory summary is required")
	}
	return ValidateEmbedding(item.Embedding)
}

func ValidateSearch(userID string, embedding []float64, threshold float64, limit int) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("memory user_id is required")
	}
	if err := ValidateEmbedding(embedding); err != nil {
		return err
	}
	if math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold < 0 || threshold > 1 {
		return errors.New("memory similarity threshold must be within [0,1]")
	}
	if limit < 1 || limit > 50 {
		return errors.New("memory search limit must be within [1,50]")
	}
	return nil
}
