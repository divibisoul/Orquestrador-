package octacore

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Backend string

const (
	BackendInProcess  Backend = "IN_PROCESS"
	BackendWASM       Backend = "WEBASSEMBLY"
	BackendWebGPU     Backend = "WEBGPU"
	BackendRemoteMesh Backend = "REMOTE_MESH"
)

type JobKind string

const (
	KindResearch    JobKind = "research"
	KindPerceive    JobKind = "perceive"
	KindTool        JobKind = "tool"
	KindSessionStep JobKind = "session_step"
	KindSARAudit    JobKind = "sara_audit"
	KindSARACycle   JobKind = "sara_cycle"
	KindDispatch    JobKind = "dispatch"
	KindCustom      JobKind = "custom"
)

const UnifiedCoreLaneCount = 8

type LaneID string

const (
	LaneOrchestration LaneID = "L0"
	LaneRouting       LaneID = "L1"
	LaneNeuralSignal  LaneID = "L2"
	LaneDecision      LaneID = "L3"
	LaneOrbital       LaneID = "L4"
	LaneSimulation    LaneID = "L5"
	LaneResourceLease LaneID = "L6"
	LaneExecution     LaneID = "L7"
)

type ExecutiveLane struct {
	ID          LaneID `json:"id"`
	Component   string `json:"component"`
	Role        string `json:"role"`
	Transport   string `json:"transport"`
	Required    bool   `json:"required"`
}

type SlotID string

const (
	G0 SlotID = "G0"
	G1 SlotID = "G1"
	G2 SlotID = "G2"
	G3 SlotID = "G3"
	G4 SlotID = "G4"
	G5 SlotID = "G5"
	G6 SlotID = "G6"
	G7 SlotID = "G7"
)

type SlotStatus string

const (
	SlotImplemented    SlotStatus = "IMPLEMENTED"
	SlotAdapterReady   SlotStatus = "ADAPTER_READY"
	SlotRepoPresent    SlotStatus = "REPO_PRESENT_RUNTIME_UNVERIFIED"
	SlotPendingRepo    SlotStatus = "PENDING_REPO"
	SlotPendingAdapter SlotStatus = "PENDING_KERNEL_ADAPTER"
)

type Slot struct {
	Slot         SlotID     `json:"slot"`
	Nucleus      string     `json:"nucleus"`
	Role         string     `json:"role"`
	Status       SlotStatus `json:"status"`
	Capabilities []string   `json:"capabilities,omitempty"`
	Execution    []Backend  `json:"execution,omitempty"`
}

type Job struct {
	JobID         string         `json:"job_id"`
	CorrelationID string         `json:"correlation_id"`
	Kind          JobKind        `json:"kind"`
	Source        SlotID         `json:"source"`
	Target        string         `json:"target"`
	BackendPrefs  []Backend      `json:"backend_prefs"`
	ParallelGroup *string        `json:"parallel_group,omitempty"`
	Barrier       *string        `json:"barrier,omitempty"`
	Payload       map[string]any `json:"payload"`
	Priority      int            `json:"priority"`
	TTLMS         int64          `json:"ttl_ms"`
}

type Result struct {
	JobID         string         `json:"job_id"`
	CorrelationID string         `json:"correlation_id"`
	OK            bool           `json:"ok"`
	BackendUsed   string         `json:"backend_used"`
	Output        map[string]any `json:"output,omitempty"`
	Error         *Error         `json:"error,omitempty"`
	Metrics       Metrics        `json:"metrics"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Metrics struct {
	LatencyMS   int64 `json:"latency_ms"`
	QueueWaitMS int64 `json:"queue_wait_ms"`
}

type VagusEnvelope struct {
	VagusVersion  string         `json:"vagus_version"`
	MessageID     string         `json:"message_id"`
	CorrelationID string         `json:"correlation_id"`
	Source        string         `json:"source"`
	Target        string         `json:"target"`
	Priority      int            `json:"priority"`
	TTL           int64          `json:"ttl"`
	Type          string         `json:"type"`
	Payload       map[string]any `json:"payload"`
}

const VagusVersion = "1.0"

func (j Job) Validate() error {
	if strings.TrimSpace(j.JobID) == "" || strings.TrimSpace(j.CorrelationID) == "" {
		return errors.New("job_id and correlation_id are required")
	}
	if j.Source < G0 || j.Source > G7 {
		return fmt.Errorf("invalid source slot: %s", j.Source)
	}
	if strings.TrimSpace(j.Target) == "" {
		return errors.New("target is required")
	}
	switch j.Kind {
	case KindResearch, KindPerceive, KindTool, KindSessionStep, KindSARAudit, KindSARACycle, KindDispatch, KindCustom:
	default:
		return fmt.Errorf("invalid job kind: %s", j.Kind)
	}
	if j.Priority < 0 || j.Priority > 100 {
		return errors.New("priority must be 0..100")
	}
	if j.TTLMS <= 0 {
		return errors.New("ttl_ms must be positive")
	}
	if len(j.BackendPrefs) == 0 && j.Kind != KindSARAudit && j.Kind != KindSARACycle {
		return errors.New("backend_prefs are required for non-SARA jobs")
	}
	if j.Payload == nil && j.Kind != KindSARAudit && j.Kind != KindSARACycle {
		return errors.New("payload is required")
	}
	return nil
}

func (e VagusEnvelope) Validate() error {
	if e.VagusVersion != VagusVersion || strings.TrimSpace(e.MessageID) == "" || strings.TrimSpace(e.CorrelationID) == "" {
		return errors.New("invalid Vagus envelope identity")
	}
	if strings.TrimSpace(e.Source) == "" || strings.TrimSpace(e.Target) == "" || strings.TrimSpace(e.Type) == "" {
		return errors.New("invalid Vagus envelope routing fields")
	}
	if e.Priority < 0 || e.Priority > 100 || e.TTL <= 0 {
		return errors.New("invalid Vagus priority/ttl")
	}
	if !(e.Type == "gpu.submit" || e.Type == "gpu.result" || e.Type == "gpu.barrier" ||
		strings.HasPrefix(e.Type, "health.") || strings.HasPrefix(e.Type, "capability.") ||
		strings.HasPrefix(e.Type, "signal.") || strings.HasPrefix(e.Type, "sara.") ||
		strings.HasPrefix(e.Type, "session.") || strings.HasPrefix(e.Type, "research.")) {
		return fmt.Errorf("unsupported Vagus type: %s", e.Type)
	}
	return nil
}

func NewJobID() string { return fmt.Sprintf("octa-%d", time.Now().UnixNano()) }
