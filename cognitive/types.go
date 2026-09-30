package cognitive

import (
	"context"
	"time"
)

type PeerDescriptor struct {
	Nucleus string
}

type MeshPeer interface {
	ConfiguredPeers() []PeerDescriptor
	Discover(context.Context, string) (map[string]any, error)
	CallBestDynamic(context.Context, string, map[string]any, string) (map[string]any, string, error)
}

type PolicyAuditor interface {
	Configured() bool
	Audit(context.Context, string, string) (map[string]any, error)
}

type RunStore interface {
	Configured() bool
	RecordRun(context.Context, map[string]any) error
}

type Config struct {
	Enabled              bool
	WorkingMemoryItems   int
	WorkingMemoryTTL     time.Duration
	GoalTTL              time.Duration
	RequireSaraForPolicy bool
}

func DefaultConfig() Config {
	return Config{
		Enabled:              true,
		WorkingMemoryItems:   64,
		WorkingMemoryTTL:     15 * time.Minute,
		GoalTTL:              10 * time.Minute,
		RequireSaraForPolicy: true,
	}
}

type Goal struct {
	ID              string         `json:"goal_id"`
	Objective       string         `json:"objective"`
	SuccessCriteria []string       `json:"success_criteria"`
	Capabilities    []string       `json:"capabilities"`
	Input           map[string]any `json:"input"`
	Risk            float64        `json:"risk"`
	Cost            float64        `json:"cost"`
	Urgency         float64        `json:"urgency"`
	Impact          float64        `json:"impact"`
	Irreversible    bool           `json:"irreversible"`
	CorrelationID   string         `json:"correlation_id"`
	CreatedAt       time.Time      `json:"created_at"`
	ExpiresAt       time.Time      `json:"expires_at"`
}

type Step struct {
	ID            string         `json:"step_id"`
	GoalID        string         `json:"goal_id"`
	Capability    string         `json:"capability"`
	Target        string         `json:"target,omitempty"`
	Payload       map[string]any `json:"payload"`
	ParallelGroup string         `json:"parallel_group,omitempty"`
	CorrelationID string         `json:"correlation_id"`
}

type Observation struct {
	GoalID        string    `json:"goal_id"`
	StepID        string    `json:"step_id"`
	Capability    string    `json:"capability"`
	OK            bool      `json:"ok"`
	Peer          string    `json:"peer,omitempty"`
	LatencyMS     int64     `json:"latency_ms"`
	CorrelationID string    `json:"correlation_id"`
	Error         string    `json:"error,omitempty"`
	At            time.Time `json:"at"`
}
