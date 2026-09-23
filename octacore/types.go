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
	BackendSARAHTTP   Backend = "SARA_HTTP"
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
	SlotImplemented           SlotStatus = "IMPLEMENTED"
	SlotAdapterReady          SlotStatus = "ADAPTER_READY"
	SlotRepoPresentUnverified SlotStatus = "REPO_PRESENT_RUNTIME_UNVERIFIED"
	SlotPendingKernelAdapter  SlotStatus = "PENDING_KERNEL_ADAPTER"
	SlotPendingInfrastructure SlotStatus = "PENDING_INFRASTRUCTURE"
)

type OctaCoreSlot struct {
	Slot         SlotID     `json:"slot"`
	Nucleus      string     `json:"nucleus"`
	Role         string     `json:"role"`
	Status       SlotStatus `json:"status"`
	Capabilities []string   `json:"capabilities,omitempty"`
	Execution    []Backend  `json:"execution,omitempty"`
}

type OctaCoreJob struct {
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

type GpuJob = OctaCoreJob

type OctaCoreError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type OctaCoreMetrics struct {
	LatencyMS   int64 `json:"latency_ms"`
	QueueWaitMS int64 `json:"queue_wait_ms"`
}

type OctaCoreResult struct {
	JobID         string          `json:"job_id"`
	CorrelationID string          `json:"correlation_id"`
	OK            bool            `json:"ok"`
	BackendUsed   string          `json:"backend_used"`
	Output        map[string]any  `json:"output,omitempty"`
	Error         *OctaCoreError  `json:"error,omitempty"`
	Metrics       OctaCoreMetrics `json:"metrics"`
}

type GpuResult = OctaCoreResult

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

func NewJobID() string { return fmt.Sprintf("octa-%d", time.Now().UnixNano()) }

func (j OctaCoreJob) Validate() error {
	if strings.TrimSpace(j.JobID) == "" {
		return errors.New("job_id is required")
	}
	if strings.TrimSpace(j.CorrelationID) == "" {
		return errors.New("correlation_id is required")
	}
	if !validSlot(j.Source) {
		return fmt.Errorf("source slot invalid: %s", j.Source)
	}
	if !validTarget(j.Target) {
		return fmt.Errorf("target invalid: %s", j.Target)
	}
	if _, ok := validJobKinds[j.Kind]; !ok {
		return fmt.Errorf("job kind invalid: %s", j.Kind)
	}
	if j.Priority < 0 || j.Priority > 100 {
		return errors.New("priority must be between 0 and 100")
	}
	if j.TTLMS <= 0 {
		return errors.New("ttl_ms must be positive")
	}
	if j.Kind != KindSARACycle && j.Kind != KindSARAudit && len(j.BackendPrefs) == 0 {
		return errors.New("backend_prefs required for non-regenerative job")
	}
	return nil
}

func (e VagusEnvelope) Validate() error {
	if e.VagusVersion != VagusVersion {
		return fmt.Errorf("unsupported vagus_version: %s", e.VagusVersion)
	}
	if strings.TrimSpace(e.MessageID) == "" || strings.TrimSpace(e.CorrelationID) == "" {
		return errors.New("message_id and correlation_id are required")
	}
	if strings.TrimSpace(e.Source) == "" || strings.TrimSpace(e.Target) == "" {
		return errors.New("source and target are required")
	}
	if strings.TrimSpace(e.Type) == "" {
		return errors.New("type is required")
	}
	if !validVagusType(e.Type) {
		return fmt.Errorf("unsupported vagus type: %s", e.Type)
	}
	if e.Priority < 0 || e.Priority > 100 {
		return errors.New("priority must be between 0 and 100")
	}
	if e.TTL <= 0 {
		return errors.New("ttl must be positive")
	}
	return nil
}

var validJobKinds = map[JobKind]struct{}{
	KindResearch: {}, KindPerceive: {}, KindTool: {}, KindSessionStep: {},
	KindSARAudit: {}, KindSARACycle: {}, KindDispatch: {}, KindCustom: {},
}

func validSlot(s SlotID) bool {
	switch s {
	case G0, G1, G2, G3, G4, G5, G6, G7:
		return true
	default:
		return false
	}
}

func validTarget(t string) bool {
	t = strings.TrimSpace(t)
	if t == "scheduler" {
		return true
	}
	return validSlot(SlotID(t))
}

func validVagusType(t string) bool {
	prefixes := []string{"health.", "capability.", "signal.", "sara.", "session.", "research."}
	if t == "gpu.submit" || t == "gpu.result" || t == "gpu.barrier" {
		return true
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}
