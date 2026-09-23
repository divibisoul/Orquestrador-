package octacore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type SchedulerConfig struct {
	MaxInflight       int
	TokenCapacity     int
	TokenRefillPerSec float64
	FailureThreshold  int
	CircuitCooldown   time.Duration
}

func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{MaxInflight: 8, TokenCapacity: 8, TokenRefillPerSec: 8, FailureThreshold: 3, CircuitCooldown: 30 * time.Second}
}

type ControlPublisher interface {
	Publish(ctx context.Context, event VagusEnvelope) error
	Status() string
}

type SlotHealth struct {
	Slot          SlotID `json:"slot"`
	Status        string `json:"status"`
	Inflight      int    `json:"inflight"`
	Failures      int    `json:"failures"`
	Circuit       string `json:"circuit"`
	LastLatencyMS int64  `json:"last_latency_ms"`
	LastError     string `json:"last_error,omitempty"`
}

type SchedulerHealth struct {
	Status             string       `json:"status"`
	ThrottleLevel      int          `json:"throttle_level"`
	Inflight           int          `json:"inflight"`
	QueueDepth         int          `json:"queue_depth"`
	ControlPlane       string       `json:"control_plane"`
	SuperGPUConnected  bool         `json:"supergpu_connected"`
	ParallelismEnabled bool         `json:"parallelism_enabled"`
	Slots              []SlotHealth `json:"slots"`
}

type slotRuntimeState struct {
	mu            sync.Mutex
	inflight      int
	failures      int
	circuit       string
	openUntil     time.Time
	halfOpenInUse bool
	lastLatency   time.Duration
	lastError     string
}

type tokenBucket struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	rate     float64
	last     time.Time
}

func newTokenBucket(capacity int, rate float64) *tokenBucket {
	if capacity < 1 {
		capacity = 1
	}
	if rate <= 0 {
		rate = float64(capacity)
	}
	now := time.Now()
	return &tokenBucket{capacity: float64(capacity), tokens: float64(capacity), rate: rate, last: now}
}

func (b *tokenBucket) take(ctx context.Context) error {
	for {
		now := time.Now()
		b.mu.Lock()
		elapsed := now.Sub(b.last).Seconds()
		if elapsed > 0 {
			b.tokens = minFloat(b.capacity, b.tokens+elapsed*b.rate)
			b.last = now
		}
		if b.tokens >= 1 {
			b.tokens--
			b.mu.Unlock()
			return nil
		}
		wait := time.Duration((1 - b.tokens) / b.rate * float64(time.Second))
		b.mu.Unlock()
		timer := time.NewTimer(maxDuration(wait, 5*time.Millisecond))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

type OctaCoreScheduler struct {
	cfg          SchedulerConfig
	compute      *supergpu.Runtime
	peers        *mesh.PeerClient
	sara         *backend.SARAProxy
	control      ControlPublisher
	slots        map[SlotID]OctaCoreSlot
	state        map[SlotID]*slotRuntimeState
	bucket       *tokenBucket
	throttle     atomic.Int32
	inflight     atomic.Int32
	queue        atomic.Int32
	halted       atomic.Bool
	kernelMu     sync.RWMutex
	localKernels map[SlotID]func(context.Context, OctaCoreJob) (map[string]any, error)
	computeMu    sync.RWMutex
}

func NewScheduler(cfg SchedulerConfig, control ControlPublisher, compute *supergpu.Runtime) (*OctaCoreScheduler, error) {
	if cfg.MaxInflight <= 0 {
		return nil, errors.New("MaxInflight must be positive")
	}
	if cfg.FailureThreshold <= 0 {
		return nil, errors.New("FailureThreshold must be positive")
	}
	if cfg.CircuitCooldown <= 0 {
		return nil, errors.New("CircuitCooldown must be positive")
	}
	peers, err := mesh.NewPeerClient(nil)
	if err != nil {
		return nil, fmt.Errorf("create Mesh peer client: %w", err)
	}
	s := &OctaCoreScheduler{cfg: cfg, compute: compute, peers: peers, sara: backend.NewSARAProxy(backend.DefaultConfig()), control: control, slots: defaultSlots(), state: make(map[SlotID]*slotRuntimeState, 8), bucket: newTokenBucket(cfg.TokenCapacity, cfg.TokenRefillPerSec), localKernels: make(map[SlotID]func(context.Context, OctaCoreJob) (map[string]any, error))}
	for _, slot := range []SlotID{G0, G1, G2, G3, G4, G5, G6, G7} {
		s.state[slot] = &slotRuntimeState{circuit: string(CircuitClosed)}
	}
	return s, nil
}

func defaultSlots() map[SlotID]OctaCoreSlot {
	return map[SlotID]OctaCoreSlot{
		G0: {Slot: G0, Nucleus: "SARA", Role: "regenerative compute kernel", Status: SlotImplemented, Capabilities: []string{"sara.cycle", "sara.audit", "sara.regenerate", "sara.state", "sara.trace"}, Execution: []Backend{BackendRemoteMesh}},
		G1: {Slot: G1, Nucleus: "N01", Role: "edge ingress / host gateway kernels", Status: SlotRepoPresentUnverified, Execution: []Backend{BackendRemoteMesh}},
		G2: {Slot: G2, Nucleus: "N02", Role: "conversation turn kernels", Status: SlotRepoPresentUnverified, Execution: []Backend{BackendRemoteMesh}},
		G3: {Slot: G3, Nucleus: "N03", Role: "perception / multimodal prep kernels", Status: SlotRepoPresentUnverified, Execution: []Backend{BackendRemoteMesh}},
		G4: {Slot: G4, Nucleus: "N04", Role: "tools / documents / research kernels", Status: SlotAdapterReady, Execution: []Backend{BackendRemoteMesh}},
		G5: {Slot: G5, Nucleus: "N05", Role: "dispatch kernels (existing runtime)", Status: SlotRepoPresentUnverified, Execution: []Backend{BackendRemoteMesh}},
		G6: {Slot: G6, Nucleus: "N06", Role: "cognition / session batching + SARA client", Status: SlotAdapterReady, Execution: []Backend{BackendRemoteMesh}},
		G7: {Slot: G7, Nucleus: "N07", Role: "SuperGPU scheduler + Mesh router + correlation", Status: SlotImplemented, Capabilities: []string{"octacore.submit", "octacore.batch", "octacore.health", "mesh.delegate", "mesh.supergpu.execute", "mesh.supergpu.parallel"}, Execution: []Backend{BackendInProcess, BackendRemoteMesh}},
	}
}

func (s *OctaCoreScheduler) Inventory() []OctaCoreSlot {
	out := make([]OctaCoreSlot, 0, len(s.slots))
	for _, slot := range s.slots {
		out = append(out, slot)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slot < out[j].Slot })
	return out
}

func (s *OctaCoreScheduler) RegisterLocalKernel(slot SlotID, handler func(context.Context, OctaCoreJob) (map[string]any, error)) error {
	if slot != G7 {
		return errors.New("local kernel registration is restricted to G7 until a concrete local kernel contract exists")
	}
	if handler == nil {
		return errors.New("local kernel handler is required")
	}
	s.kernelMu.Lock()
	s.localKernels[slot] = handler
	s.kernelMu.Unlock()
	return nil
}

func (s *OctaCoreScheduler) localKernel(slot SlotID) func(context.Context, OctaCoreJob) (map[string]any, error) {
	s.kernelMu.RLock()
	defer s.kernelMu.RUnlock()
	return s.localKernels[slot]
}

func (s *OctaCoreScheduler) AttachSuperGPU(runtime *supergpu.Runtime) error {
	if runtime == nil {
		return errors.New("SuperGPU runtime is required")
	}
	runtime.Discover()
	s.computeMu.Lock()
	s.compute = runtime
	s.computeMu.Unlock()
	return nil
}

func (s *OctaCoreScheduler) SuperGPUConnected() bool {
	s.computeMu.RLock()
	defer s.computeMu.RUnlock()
	return s.compute != nil
}

func (s *OctaCoreScheduler) SetThrottle(level int) error {
	if level < 0 || level > 3 {
		return errors.New("throttle level must be 0..3")
	}
	s.throttle.Store(int32(level))
	return nil
}
func (s *OctaCoreScheduler) Resume()      { s.halted.Store(false) }
func (s *OctaCoreScheduler) Halt()        { s.halted.Store(true) }
func (s *OctaCoreScheduler) Halted() bool { return s.halted.Load() }

func (s *OctaCoreScheduler) Execute(ctx context.Context, job OctaCoreJob) OctaCoreResult {
	queuedAt := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := job.Validate(); err != nil {
		return failedResult(job, "INVALID_JOB", err, 0, 0)
	}
	if s.halted.Load() {
		return failedResult(job, "OCTACORE_HALTED", errors.New("Octacore scheduler halted"), 0, 0)
	}
	slot, err := s.resolveSlot(job)
	if err != nil {
		return failedResult(job, "SLOT_RESOLUTION_FAILED", err, 0, 0)
	}
	if slot.Slot != G0 && len(job.BackendPrefs) == 0 {
		job.BackendPrefs = []Backend{BackendRemoteMesh}
	}
	deadline := time.Now().Add(time.Duration(job.TTLMS) * time.Millisecond)
	if existing, ok := ctx.Deadline(); !ok || deadline.Before(existing) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}
	s.queue.Add(1)
	defer s.queue.Add(-1)
	if err := s.bucket.take(ctx); err != nil {
		return failedResult(job, "TOKEN_BUCKET_EXHAUSTED", err, 0, time.Since(queuedAt).Milliseconds())
	}
	if err := s.acquireInflight(ctx, slot.Slot); err != nil {
		return failedResult(job, "THROTTLED", err, 0, time.Since(queuedAt).Milliseconds())
	}
	defer s.releaseInflight(slot.Slot)
	queueWait := time.Since(queuedAt).Milliseconds()
	submittedAt := time.Now()
	s.publish(ctx, newVagusEvent("gpu.submit", string(job.Source), string(slot.Slot), job.Priority, job.TTLMS, job.CorrelationID, map[string]any{"job": job}))
	output, backendUsed, execErr := s.executeSelected(ctx, job, slot)
	latency := time.Since(submittedAt).Milliseconds()
	result := OctaCoreResult{JobID: job.JobID, CorrelationID: job.CorrelationID, OK: execErr == nil, BackendUsed: backendUsed, Output: output, Metrics: OctaCoreMetrics{LatencyMS: latency, QueueWaitMS: queueWait}}
	if execErr != nil {
		result.Error = toOctaError(execErr)
		s.recordFailure(slot.Slot, latency, execErr.Error())
	} else {
		s.recordSuccess(slot.Slot, latency)
	}
	source := string(slot.Slot)
	if slot.Slot == G0 {
		source = "SARA"
	}
	s.publish(ctx, newVagusEvent("gpu.result", source, string(job.Source), job.Priority, job.TTLMS, job.CorrelationID, map[string]any{"result": result}))
	return result
}

func (s *OctaCoreScheduler) ExecutePlan(ctx context.Context, jobs []OctaCoreJob) []OctaCoreResult {
	results := make([]OctaCoreResult, len(jobs))
	if len(jobs) == 0 {
		return results
	}
	pending := make(map[int]OctaCoreJob, len(jobs))
	for i, job := range jobs {
		pending[i] = job
	}
	blockedBarriers := make(map[string]string)
	for len(pending) > 0 {
		ready := make([]int, 0, len(pending))
		for i, job := range pending {
			if name := strings.TrimSpace(ptrString(job.Barrier)); name != "" {
				if reason, blocked := blockedBarriers[name]; blocked {
					results[i] = failedResult(job, "BARRIER_DEPENDENCY_FAILED", errors.New(reason), 0, 0)
					delete(pending, i)
					continue
				}
			}
			if barrierReady(i, job, pending) {
				ready = append(ready, i)
			}
		}
		if len(ready) == 0 {
			for i, job := range pending {
				results[i] = failedResult(job, "BARRIER_DEADLOCK", errors.New("barrier dependency has no executable frontier"), 0, 0)
			}
			break
		}
		sort.SliceStable(ready, func(i, j int) bool {
			if jobs[ready[i]].Priority != jobs[ready[j]].Priority {
				return jobs[ready[i]].Priority > jobs[ready[j]].Priority
			}
			return ready[i] < ready[j]
		})
		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, i := range ready {
			i := i
			job := pending[i]
			wg.Add(1)
			go func() { defer wg.Done(); result := s.Execute(ctx, job); mu.Lock(); results[i] = result; mu.Unlock() }()
		}
		wg.Wait()
		completedBarriers := map[string][]string{}
		for _, i := range ready {
			delete(pending, i)
			if name := strings.TrimSpace(ptrString(jobs[i].Barrier)); name != "" {
				completedBarriers[name] = append(completedBarriers[name], jobs[i].JobID)
				if !results[i].OK && blockedBarriers[name] == "" {
					reason := "barrier dependency failed"
					if results[i].Error != nil {
						reason = results[i].Error.Code + ":" + results[i].Error.Message
					}
					blockedBarriers[name] = reason
				}
			}
		}
		for name, ids := range completedBarriers {
			stillPending := false
			for _, job := range pending {
				if strings.TrimSpace(ptrString(job.Barrier)) == name {
					stillPending = true
					break
				}
			}
			if !stillPending {
				s.publish(ctx, newVagusEvent("gpu.barrier", "G7", "G6", 100, 1, barrierCorrelation(jobs, name), map[string]any{"barrier": name, "completed_job_ids": ids, "joined": true}))
			}
		}
	}
	return results
}

func barrierReady(_ int, job OctaCoreJob, pending map[int]OctaCoreJob) bool {
	name := strings.TrimSpace(ptrString(job.Barrier))
	if name == "" {
		return true
	}
	if isBarrierConsumer(job) {
		for _, other := range pending {
			if strings.TrimSpace(ptrString(other.Barrier)) != name || isBarrierConsumer(other) {
				continue
			}
			return false
		}
		return true
	}
	// Non-regenerative jobs are producers for a named join. They can execute
	// concurrently; G0 is the consumer/authority for regenerative work.
	return true
}

func isBarrierConsumer(job OctaCoreJob) bool {
	return job.Target == string(G0) || job.Kind == KindSARAudit || job.Kind == KindSARACycle
}

func (s *OctaCoreScheduler) resolveSlot(job OctaCoreJob) (OctaCoreSlot, error) {
	if job.Kind == KindSARAudit || job.Kind == KindSARACycle {
		return s.slots[G0], nil
	}
	if job.Target != "scheduler" {
		slot := s.slots[SlotID(job.Target)]
		if slot.Slot == "" {
			return OctaCoreSlot{}, fmt.Errorf("unknown slot: %s", job.Target)
		}
		return slot, nil
	}
	switch job.Kind {
	case KindResearch, KindTool:
		return s.slots[G4], nil
	case KindPerceive:
		return s.slots[G3], nil
	case KindSessionStep:
		return s.slots[G6], nil
	case KindDispatch:
		return s.slots[G7], nil
	case KindCustom:
		return OctaCoreSlot{}, errors.New("custom job requires explicit target")
	default:
		return OctaCoreSlot{}, fmt.Errorf("no scheduler route for kind: %s", job.Kind)
	}
}

func (s *OctaCoreScheduler) executeSelected(ctx context.Context, job OctaCoreJob, slot OctaCoreSlot) (map[string]any, string, error) {
	if slot.Slot == G0 {
		return s.executeG0(ctx, job)
	}
	for _, preferred := range job.BackendPrefs {
		switch preferred {
		case BackendInProcess:
			if slot.Slot == G7 {
				if job.Kind == KindDispatch {
					child, ok := parseNestedJob(job.Payload)
					if !ok {
						return nil, string(BackendInProcess), errors.New("dispatch job requires payload.job")
					}
					childResult := s.Execute(ctx, child)
					if !childResult.OK {
						return nil, string(BackendInProcess), fmt.Errorf("child job failed: %s", childResult.Error.Code)
					}
					return map[string]any{"result": childResult}, "IN_PROCESS_SCHEDULER", nil
				}
				if handler := s.localKernel(slot.Slot); handler != nil {
					output, err := handler(ctx, job)
					return output, string(BackendInProcess), err
				}
				output, used, err := s.executeG7Compute(ctx, job)
				if err == nil {
					return output, used, nil
				}
				return nil, string(BackendInProcess), err
			}
		case BackendRemoteMesh:
			if slot.Slot != G7 {
				return s.executeRemote(ctx, job, slot)
			}
		case BackendWASM, BackendWebGPU:
			continue
		case "SARA_HTTP":
			return nil, "SARA_HTTP", errors.New("SARA_HTTP_ONLY_FOR_G0")
		default:
			return nil, "", fmt.Errorf("unsupported backend: %s", preferred)
		}
	}
	return nil, "", errors.New("BACKEND_UNAVAILABLE")
}

func (s *OctaCoreScheduler) executeG0(ctx context.Context, job OctaCoreJob) (map[string]any, string, error) {
	if s.sara == nil || !s.sara.Configured() {
		return nil, "SARA_HTTP", errors.New("SARA_UNAVAILABLE:SARA service is not configured")
	}
	input, _ := job.Payload["input"].(string)
	if strings.TrimSpace(input) == "" {
		return nil, "SARA_HTTP", errors.New("SARA_INPUT_REQUIRED")
	}
	switch job.Kind {
	case KindSARAudit:
		out, err := s.sara.Audit(ctx, input, job.CorrelationID)
		return out, "SARA_HTTP", err
	case KindSARACycle:
		cycleID, _ := job.Payload["cycle_id"].(string)
		out, err := s.sara.Cycle(ctx, input, cycleID, job.CorrelationID)
		return out, "SARA_HTTP", err
	}
	operation, _ := job.Payload["operation"].(string)
	switch strings.TrimSpace(operation) {
	case "sara.audit":
		out, err := s.sara.Audit(ctx, input, job.CorrelationID)
		return out, "SARA_HTTP", err
	case "sara.cycle":
		cycleID, _ := job.Payload["cycle_id"].(string)
		out, err := s.sara.Cycle(ctx, input, cycleID, job.CorrelationID)
		return out, "SARA_HTTP", err
	case "sara.regenerate":
		out, err := s.sara.Regenerate(ctx, input, job.CorrelationID)
		return out, "SARA_HTTP", err
	case "sara.state":
		out, err := s.sara.State(ctx, job.CorrelationID)
		return out, "SARA_HTTP", err
	case "sara.trace":
		cycleID, _ := job.Payload["cycle_id"].(string)
		out, err := s.sara.Trace(ctx, cycleID, job.CorrelationID)
		return out, "SARA_HTTP", err
	default:
		return nil, "SARA_HTTP", fmt.Errorf("SARA_OPERATION_UNSUPPORTED:%s", strings.TrimSpace(operation))
	}
}

func (s *OctaCoreScheduler) executeRemote(ctx context.Context, job OctaCoreJob, slot OctaCoreSlot) (map[string]any, string, error) {
	capability, _ := job.Payload["capability"].(string)
	capability = strings.TrimSpace(capability)
	if capability == "" {
		return nil, string(BackendRemoteMesh), errors.New("REMOTE_CAPABILITY_REQUIRED")
	}
	callCapability := capability
	preflightCapability := capability
	if slot.Slot == G4 || slot.Slot == G6 {
		preflightCapability = "octacore.execute"
	}
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	supported, probeErr := s.peers.SupportsCapability(probeCtx, slot.Nucleus, preflightCapability)
	cancel()
	if probeErr != nil {
		return nil, string(BackendRemoteMesh), fmt.Errorf("REMOTE_CAPABILITY_DISCOVERY_FAILED:%w", probeErr)
	}
	if !supported {
		return nil, string(BackendRemoteMesh), fmt.Errorf("REMOTE_CAPABILITY_UNAVAILABLE:%s:%s", slot.Nucleus, preflightCapability)
	}
	payload := cloneMap(job.Payload)
	delete(payload, "capability")
	payload["octacore"] = map[string]any{"job_id": job.JobID, "correlation_id": job.CorrelationID, "source": job.Source, "target": slot.Slot, "kind": job.Kind}
	if slot.Slot == G4 || slot.Slot == G6 {
		callCapability = "octacore.execute"
		payload = map[string]any{"capability": capability, "payload": cloneMap(job.Payload), "job_id": job.JobID}
	}
	out, err := s.peers.CallWithCorrelation(ctx, slot.Nucleus, callCapability, payload, job.CorrelationID)
	return out, string(BackendRemoteMesh), err
}

func (s *OctaCoreScheduler) executeG7Compute(ctx context.Context, job OctaCoreJob) (map[string]any, string, error) {
	s.computeMu.RLock()
	compute := s.compute
	s.computeMu.RUnlock()
	if compute == nil {
		return nil, string(BackendInProcess), errors.New("SUPERGPU_NOT_CONNECTED")
	}
	operation, _ := job.Payload["operation"].(string)
	operation = strings.TrimSpace(operation)
	values, err := numberArray(job.Payload["values"])
	if err != nil {
		return nil, string(BackendInProcess), err
	}
	if operation == "" {
		return nil, string(BackendInProcess), errors.New("IN_PROCESS_OPERATION_REQUIRED")
	}
	backendName, err := compute.Select("")
	if err != nil {
		return nil, string(BackendInProcess), err
	}
	if err := compute.Reserve(backendName.ID, job.JobID); err != nil {
		return nil, string(BackendInProcess), err
	}
	defer compute.Release(backendName.ID, job.JobID)
	result, err := compute.Execute(ctx, backendName, operation, values)
	if err != nil {
		return nil, string(BackendInProcess), err
	}
	return map[string]any{"operation": operation, "values": result, "backend": backendName.ID}, string(BackendInProcess), nil
}

func (s *OctaCoreScheduler) acquireInflight(ctx context.Context, slot SlotID) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !s.slotCircuitAllows(slot) {
			return fmt.Errorf("slot circuit open: %s", slot)
		}
		limit := s.effectiveInflightLimit()
		if slot == G0 {
			limit = 1
		}
		current := s.inflight.Load()
		if int(current) < limit && s.inflight.CompareAndSwap(current, current+1) {
			st := s.state[slot]
			st.mu.Lock()
			st.inflight++
			st.mu.Unlock()
			return nil
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
func (s *OctaCoreScheduler) releaseInflight(slot SlotID) {
	s.inflight.Add(-1)
	st := s.state[slot]
	st.mu.Lock()
	st.inflight--
	if st.halfOpenInUse {
		st.halfOpenInUse = false
	}
	st.mu.Unlock()
}
func (s *OctaCoreScheduler) effectiveInflightLimit() int {
	base := s.cfg.MaxInflight
	switch int(s.throttle.Load()) {
	case 1:
		base = maxInt(1, base/2)
	case 2:
		base = maxInt(1, base/4)
	case 3:
		base = 1
	}
	return base
}
func (s *OctaCoreScheduler) slotCircuitAllows(slot SlotID) bool {
	st := s.state[slot]
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.circuit != string(CircuitOpen) {
		return true
	}
	if time.Now().Before(st.openUntil) {
		return false
	}
	if st.halfOpenInUse {
		return false
	}
	st.circuit = string(CircuitHalfOpen)
	st.halfOpenInUse = true
	return true
}
func (s *OctaCoreScheduler) recordSuccess(slot SlotID, latencyMS int64) {
	st := s.state[slot]
	st.mu.Lock()
	st.failures = 0
	st.circuit = string(CircuitClosed)
	st.openUntil = time.Time{}
	st.lastLatency = time.Duration(latencyMS) * time.Millisecond
	st.lastError = ""
	st.mu.Unlock()
}
func (s *OctaCoreScheduler) recordFailure(slot SlotID, latencyMS int64, message string) {
	st := s.state[slot]
	st.mu.Lock()
	st.failures++
	st.lastLatency = time.Duration(latencyMS) * time.Millisecond
	st.lastError = message
	if st.failures >= s.cfg.FailureThreshold {
		st.circuit = string(CircuitOpen)
		st.openUntil = time.Now().Add(s.cfg.CircuitCooldown)
	}
	st.mu.Unlock()
}

func (s *OctaCoreScheduler) Health() SchedulerHealth {
	s.computeMu.RLock()
	computeConnected := s.compute != nil
	s.computeMu.RUnlock()
	out := SchedulerHealth{Status: "READY", ThrottleLevel: int(s.throttle.Load()), Inflight: int(s.inflight.Load()), QueueDepth: int(s.queue.Load()), ControlPlane: "UNCONFIGURED", ParallelismEnabled: true, SuperGPUConnected: computeConnected}
	if s.halted.Load() {
		out.Status = "HALTED"
	}
	if s.control != nil {
		out.ControlPlane = s.control.Status()
	}
	for _, slot := range s.Inventory() {
		st := s.state[slot.Slot]
		st.mu.Lock()
		out.Slots = append(out.Slots, SlotHealth{Slot: slot.Slot, Status: string(slot.Status), Inflight: st.inflight, Failures: st.failures, Circuit: st.circuit, LastLatencyMS: st.lastLatency.Milliseconds(), LastError: st.lastError})
		st.mu.Unlock()
	}
	return out
}
func (s *OctaCoreScheduler) publish(ctx context.Context, event VagusEnvelope) {
	if s.control != nil {
		_ = s.control.Publish(ctx, event)
	}
}

func newVagusEvent(typ, source, target string, priority int, ttl int64, correlation string, payload map[string]any) VagusEnvelope {
	return VagusEnvelope{VagusVersion: VagusVersion, MessageID: NewJobID(), CorrelationID: correlation, Source: source, Target: target, Priority: priority, TTL: ttl, Type: typ, Payload: payload}
}
func failedResult(job OctaCoreJob, code string, err error, latencyMS, queueWaitMS int64) OctaCoreResult {
	if err == nil {
		err = errors.New(code)
	}
	return OctaCoreResult{JobID: job.JobID, CorrelationID: job.CorrelationID, OK: false, Error: &OctaCoreError{Code: code, Message: err.Error()}, Metrics: OctaCoreMetrics{LatencyMS: latencyMS, QueueWaitMS: queueWaitMS}}
}
func toOctaError(err error) *OctaCoreError {
	if err == nil {
		return nil
	}
	msg := err.Error()
	code := "EXECUTION_ERROR"
	if strings.Contains(msg, "SARA_") {
		code = stringBefore(msg, ":")
	}
	return &OctaCoreError{Code: code, Message: msg}
}
func stringBefore(value, sep string) string {
	if i := strings.Index(value, sep); i > 0 {
		return value[:i]
	}
	return value
}
func parseNestedJob(payload map[string]any) (OctaCoreJob, bool) {
	raw, ok := payload["job"]
	if !ok {
		return OctaCoreJob{}, false
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return OctaCoreJob{}, false
	}
	var job OctaCoreJob
	if err = json.Unmarshal(b, &job); err != nil {
		return OctaCoreJob{}, false
	}
	return job, job.Validate() == nil
}
func numberArray(value any) ([]float64, error) {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...), nil
	case []any:
		out := make([]float64, len(v))
		for i, item := range v {
			n, ok := item.(float64)
			if !ok {
				if str, yes := item.(string); yes {
					parsed, err := strconv.ParseFloat(str, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid numeric value at index %d", i)
					}
					n = parsed
				} else {
					return nil, fmt.Errorf("invalid numeric value at index %d", i)
				}
			}
			out[i] = n
		}
		return out, nil
	default:
		return nil, errors.New("values must be numeric array")
	}
}
func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func barrierCorrelation(jobs []OctaCoreJob, name string) string {
	for _, job := range jobs {
		if strings.TrimSpace(ptrString(job.Barrier)) == name && strings.TrimSpace(job.CorrelationID) != "" {
			return job.CorrelationID
		}
	}
	return NewJobID()
}
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half-open"
)
