package octacore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type schedulerSlotState struct {
	mu          sync.Mutex
	inflight    int
	failures    int
	circuit     string
	retryAfter  time.Time
	halfOpenUse bool
	lastLatency time.Duration
	lastError   string
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
			b.tokens += elapsed * b.rate
			if b.tokens > b.capacity {
				b.tokens = b.capacity
			}
			b.last = now
		}
		if b.tokens >= 1 {
			b.tokens--
			b.mu.Unlock()
			return nil
		}
		wait := time.Duration(((1 - b.tokens) / b.rate) * float64(time.Second))
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

type Scheduler struct {
	cfg       Config
	compute   *supergpu.Runtime
	peers     *mesh.PeerClient
	sara      *backend.SARAProxy
	state     map[SlotID]*schedulerSlotState
	slots     map[SlotID]Slot
	bucket    *tokenBucket
	inflight  atomic.Int32
	queue     atomic.Int32
	throttle  atomic.Int32
	halted    atomic.Bool
	eventMu   sync.RWMutex
	publishFn func(context.Context, VagusEnvelope) error
}

func newScheduler(cfg Config, compute *supergpu.Runtime, peers *mesh.PeerClient, sara *backend.SARAProxy) *Scheduler {
	if cfg.MaxInflight <= 0 {
		cfg.MaxInflight = 8
	}
	if cfg.TokenCapacity <= 0 {
		cfg.TokenCapacity = cfg.MaxInflight
	}
	if cfg.TokenRefillPerSecond <= 0 {
		cfg.TokenRefillPerSecond = float64(cfg.TokenCapacity)
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.CircuitCooldown <= 0 {
		cfg.CircuitCooldown = 30 * time.Second
	}
	s := &Scheduler{
		cfg:     cfg,
		compute: compute,
		peers:   peers,
		sara:    sara,
		state:   make(map[SlotID]*schedulerSlotState, 8),
		slots:   defaultSlots(),
		bucket:  newTokenBucket(cfg.TokenCapacity, cfg.TokenRefillPerSecond),
	}
	for _, slot := range []SlotID{G0, G1, G2, G3, G4, G5, G6, G7} {
		s.state[slot] = &schedulerSlotState{circuit: "closed"}
	}
	return s
}

func defaultSlots() map[SlotID]Slot {
	return map[SlotID]Slot{
		G0: {Slot: G0, Nucleus: "SARA", Role: "regenerative compute kernel", Status: SlotImplemented, Capabilities: []string{"sara.cycle", "sara.audit", "sara.regenerate", "sara.state", "sara.trace"}, Execution: []Backend{BackendRemoteMesh}},
		G1: {Slot: G1, Nucleus: "N01", Role: "edge ingress / host gateway kernels", Status: SlotRepoPresent, Execution: []Backend{BackendRemoteMesh}},
		G2: {Slot: G2, Nucleus: "N02", Role: "conversation turn kernels", Status: SlotRepoPresent, Execution: []Backend{BackendRemoteMesh}},
		G3: {Slot: G3, Nucleus: "N03", Role: "perception / multimodal prep kernels", Status: SlotRepoPresent, Execution: []Backend{BackendRemoteMesh}},
		G4: {Slot: G4, Nucleus: "N04", Role: "tools / documents / research kernels", Status: SlotAdapterReady, Execution: []Backend{BackendRemoteMesh}},
		G5: {Slot: G5, Nucleus: "N05", Role: "dispatch kernels", Status: SlotRepoPresent, Execution: []Backend{BackendRemoteMesh}},
		G6: {Slot: G6, Nucleus: "N06", Role: "cognition / session batching + SARA client", Status: SlotAdapterReady, Execution: []Backend{BackendRemoteMesh}},
		G7: {Slot: G7, Nucleus: "N07", Role: "SuperGPU scheduler + Mesh router + correlation", Status: SlotImplemented, Capabilities: []string{"octacore.submit", "octacore.batch", "octacore.health", "supergpu.execute", "mesh.supergpu.parallel"}, Execution: []Backend{BackendInProcess}},
	}
}

func (s *Scheduler) inventory() []Slot {
	out := make([]Slot, 0, len(s.slots))
	for _, slot := range s.slots {
		out = append(out, slot)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slot < out[j].Slot })
	return out
}

func (s *Scheduler) setPublisher(fn func(context.Context, VagusEnvelope) error) {
	s.eventMu.Lock()
	s.publishFn = fn
	s.eventMu.Unlock()
}

func (s *Scheduler) publish(ctx context.Context, event VagusEnvelope) {
	s.eventMu.RLock()
	fn := s.publishFn
	s.eventMu.RUnlock()
	if fn != nil {
		_ = fn(ctx, event)
	}
}

func (s *Scheduler) health() map[string]any {
	out := map[string]any{
		"processor": "Octacore",
		"status":    "READY",
		"slots":     s.inventory(),
		"inflight":  s.inflight.Load(),
		"queue_depth": s.queue.Load(),
		"throttle_level": s.throttle.Load(),
		"supergpu": s.superGPUHealth(),
	}
	if s.halted.Load() {
		out["status"] = "HALTED"
	}
	slotHealth := make([]map[string]any, 0, 8)
	for _, slot := range s.inventory() {
		st := s.state[slot.Slot]
		st.mu.Lock()
		slotHealth = append(slotHealth, map[string]any{
			"slot":           slot.Slot,
			"nucleus":        slot.Nucleus,
			"status":         slot.Status,
			"inflight":       st.inflight,
			"failures":       st.failures,
			"circuit":        st.circuit,
			"last_latency_ms": st.lastLatency.Milliseconds(),
			"last_error":     st.lastError,
		})
		st.mu.Unlock()
	}
	out["slot_health"] = slotHealth
	return out
}

func (s *Scheduler) superGPUHealth() map[string]any {
	if s.compute == nil {
		return map[string]any{"status": "UNAVAILABLE", "reason": "existing N07 SuperGPU runtime not connected"}
	}
	return s.compute.Health()
}

func (s *Scheduler) throttleLimit() int {
	base := s.cfg.MaxInflight
	switch int(s.throttle.Load()) {
	case 1:
		return maxInt(1, base/2)
	case 2:
		return maxInt(1, base/4)
	case 3:
		return 1
	default:
		return base
	}
}

func (s *Scheduler) setThrottle(level int) error {
	if level < 0 || level > 3 {
		return errors.New("throttle level must be 0..3")
	}
	s.throttle.Store(int32(level))
	s.publish(context.Background(), makeVagus("signal.throttle", "G7", "scheduler", 100, 1000, fmt.Sprintf("throttle-%d", level), map[string]any{"level": level}))
	return nil
}

func (s *Scheduler) halt() {
	s.halted.Store(true)
	s.publish(context.Background(), makeVagus("signal.halt", "G7", "scheduler", 100, 1000, "halt", map[string]any{}))
}

func (s *Scheduler) resume() {
	s.halted.Store(false)
	s.publish(context.Background(), makeVagus("signal.resume", "G7", "scheduler", 100, 1000, "resume", map[string]any{}))
}

func (s *Scheduler) inflightCount() int { return int(s.inflight.Load()) }

func (s *Scheduler) execute(ctx context.Context, job Job) Result {
	start := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := job.Validate(); err != nil {
		return failed(job, "INVALID_OCTACORE_JOB", err, 0, 0)
	}
	if s.halted.Load() {
		return failed(job, "OCTACORE_HALTED", errors.New("Octacore is halted"), 0, 0)
	}
	if s.expired(start, job.TTLMS) {
		return failed(job, "TTL_EXPIRED", errors.New("job ttl expired before admission"), 0, 0)
	}

	slot, err := s.resolve(job)
	if err != nil {
		return failed(job, "SLOT_RESOLUTION_FAILED", err, 0, 0)
	}

	deadline := start.Add(time.Duration(job.TTLMS) * time.Millisecond)
	if parent, ok := ctx.Deadline(); !ok || deadline.Before(parent) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	s.queue.Add(1)
	defer s.queue.Add(-1)

	if err := s.bucket.take(ctx); err != nil {
		return failed(job, "ADMISSION_BACKPRESSURE", err, 0, time.Since(start).Milliseconds())
	}
	if err := s.acquire(ctx, slot.Slot); err != nil {
		return failed(job, "ADMISSION_THROTTLED", err, 0, time.Since(start).Milliseconds())
	}
	defer s.release(slot.Slot)

	queueWait := time.Since(start).Milliseconds()
	s.publish(ctx, makeVagus("gpu.submit", string(job.Source), string(slot.Slot), job.Priority, job.TTLMS, job.CorrelationID, map[string]any{"job": job}))
	output, backendUsed, execErr := s.run(ctx, job, slot)
	latency := time.Since(start).Milliseconds()

	result := Result{
		JobID:         job.JobID,
		CorrelationID: job.CorrelationID,
		OK:            execErr == nil,
		BackendUsed:   backendUsed,
		Output:        output,
		Metrics:       Metrics{LatencyMS: latency, QueueWaitMS: queueWait},
	}
	if execErr != nil {
		result.Error = &Error{Code: classifyError(execErr), Message: execErr.Error()}
		s.recordFailure(slot.Slot, latency, execErr.Error())
	} else {
		s.recordSuccess(slot.Slot, latency)
	}
	s.publish(ctx, makeVagus("gpu.result", string(slot.Slot), string(job.Source), job.Priority, job.TTLMS, job.CorrelationID, map[string]any{"result": result}))
	return result
}

func (s *Scheduler) executePlan(ctx context.Context, jobs []Job) []Result {
	results := make([]Result, len(jobs))
	pending := make(map[int]Job, len(jobs))
	for i, job := range jobs {
		pending[i] = job
	}
	if len(jobs) == 0 {
		return results
	}

	for len(pending) > 0 {
		ready := make([]int, 0, len(pending))
		for i, job := range pending {
			if barrierReady(i, job, pending) {
				ready = append(ready, i)
			}
		}
		if len(ready) == 0 {
			for i, job := range pending {
				results[i] = failed(job, "BARRIER_DEADLOCK", errors.New("no executable frontier remains"), 0, 0)
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
		for _, index := range ready {
			index := index
			wg.Add(1)
			go func() {
				defer wg.Done()
				results[index] = s.execute(ctx, jobs[index])
			}()
		}
		wg.Wait()

		for _, index := range ready {
			delete(pending, index)
		}
		barrierSignals := barriersForReady(jobs, results, ready)
		for name, signal := range barrierSignals {
			s.publish(ctx, signal)
			_ = name
		}
	}
	return results
}

func barrierReady(index int, job Job, pending map[int]Job) bool {
	if job.Barrier == nil || strings.TrimSpace(*job.Barrier) == "" {
		return true
	}
	barrier := strings.TrimSpace(*job.Barrier)
	consumerGroup := ""
	if job.ParallelGroup != nil {
		consumerGroup = strings.TrimSpace(*job.ParallelGroup)
	}
	for i, other := range pending {
		if i == index || other.Barrier == nil || strings.TrimSpace(*other.Barrier) != barrier {
			continue
		}
		otherGroup := ""
		if other.ParallelGroup != nil {
			otherGroup = strings.TrimSpace(*other.ParallelGroup)
		}
		if otherGroup == consumerGroup {
			continue
		}
		return false
	}
	return true
}

func barriersForReady(jobs []Job, results []Result, ready []int) map[string]VagusEnvelope {
	out := map[string]VagusEnvelope{}
	for _, index := range ready {
		job := jobs[index]
		if job.Barrier == nil || strings.TrimSpace(*job.Barrier) == "" {
			continue
		}
		name := strings.TrimSpace(*job.Barrier)
		corr := job.CorrelationID
		ok := results[index].OK
		out[name] = makeVagus("gpu.barrier", "G7", string(job.Source), job.Priority, job.TTLMS, corr, map[string]any{
			"barrier": name,
			"job_id":  job.JobID,
			"ok":      ok,
		})
	}
	return out
}

func (s *Scheduler) resolve(job Job) (Slot, error) {
	if job.Kind == KindSARAudit || job.Kind == KindSARACycle {
		return s.slots[G0], nil
	}
	if strings.TrimSpace(job.Target) != "" && job.Target != "scheduler" {
		slot, ok := s.slots[SlotID(job.Target)]
		if !ok {
			return Slot{}, fmt.Errorf("unknown target slot %s", job.Target)
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
	default:
		return Slot{}, errors.New("scheduler routing requires explicit target for this job kind")
	}
}

func (s *Scheduler) acquire(ctx context.Context, slot SlotID) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		st := s.state[slot]
		probe := false
		st.mu.Lock()
		now := time.Now()
		if st.circuit == "open" {
			if now.Before(st.retryAfter) {
				st.mu.Unlock()
				return errors.New("slot circuit open")
			}
			if st.halfOpenUse {
				st.mu.Unlock()
				return errors.New("slot half-open probe busy")
			}
			st.circuit = "half-open"
			st.halfOpenUse = true
			probe = true
		}
		st.mu.Unlock()

		limit := s.throttleLimit()
		if slot == G0 {
			limit = 1
		}
		admitted := false
		for {
			current := s.inflight.Load()
			if int(current) >= limit {
				break
			}
			if s.inflight.CompareAndSwap(current, current+1) {
				admitted = true
				break
			}
		}
		if admitted {
			st.mu.Lock()
			st.inflight++
			st.mu.Unlock()
			return nil
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			if probe {
				st.mu.Lock()
				if st.circuit == "half-open" {
					st.circuit = "open"
					st.halfOpenUse = false
					st.retryAfter = time.Now().Add(s.cfg.CircuitCooldown)
				}
				st.mu.Unlock()
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (s *Scheduler) release(slot SlotID) {
	s.inflight.Add(-1)
	st := s.state[slot]
	st.mu.Lock()
	st.inflight--
	if st.circuit == "half-open" {
		st.halfOpenUse = false
	}
	st.mu.Unlock()
}

func (s *Scheduler) recordSuccess(slot SlotID, latency int64) {
	st := s.state[slot]
	st.mu.Lock()
	st.failures = 0
	st.circuit = "closed"
	st.halfOpenUse = false
	st.retryAfter = time.Time{}
	st.lastLatency = time.Duration(latency) * time.Millisecond
	st.lastError = ""
	st.mu.Unlock()
}

func (s *Scheduler) recordFailure(slot SlotID, latency int64, message string) {
	st := s.state[slot]
	st.mu.Lock()
	st.failures++
	st.lastLatency = time.Duration(latency) * time.Millisecond
	st.lastError = message
	st.halfOpenUse = false
	if st.failures >= s.cfg.FailureThreshold {
		st.circuit = "open"
		st.retryAfter = time.Now().Add(s.cfg.CircuitCooldown)
	}
	st.mu.Unlock()
}

func (s *Scheduler) run(ctx context.Context, job Job, slot Slot) (map[string]any, string, error) {
	var lastErr error
	for _, preferred := range job.BackendPrefs {
		switch preferred {
		case BackendRemoteMesh:
			output, backendUsed, err := s.runRemote(ctx, job, slot)
			if err == nil {
				return output, backendUsed, nil
			}
			lastErr = err
		case BackendInProcess:
			if slot.Slot != G7 {
				lastErr = errors.New("IN_PROCESS_NOT_SUPPORTED_FOR_SLOT")
				continue
			}
			if s.compute == nil {
				lastErr = errors.New("SUPERGPU_UNAVAILABLE")
				continue
			}
			operation, _ := job.Payload["operation"].(string)
			values, err := numericSlice(job.Payload["values"])
			if err != nil {
				lastErr = err
				continue
			}
			if strings.TrimSpace(operation) == "" {
				lastErr = errors.New("IN_PROCESS_OPERATION_REQUIRED")
				continue
			}
			device, err := s.compute.Select("")
			if err != nil {
				lastErr = err
				continue
			}
			if err := s.compute.Reserve(device.ID, job.JobID); err != nil {
				lastErr = err
				continue
			}
			out, err := s.compute.Execute(ctx, device, operation, values)
			_ = s.compute.Release(device.ID, job.JobID)
			if err != nil {
				lastErr = err
				continue
			}
			return map[string]any{"operation": operation, "values": out, "device": device.ID}, string(BackendInProcess), nil
		case BackendWASM, BackendWebGPU:
			lastErr = fmt.Errorf("%s_BACKEND_UNAVAILABLE", preferred)
		default:
			lastErr = fmt.Errorf("BACKEND_UNSUPPORTED:%s", preferred)
		}
		if ctx.Err() != nil {
			return nil, string(preferred), ctx.Err()
		}
	}
	if lastErr != nil {
		return nil, "", lastErr
	}
	return nil, "", errors.New("BACKEND_PREFERENCE_EMPTY")
}

func (s *Scheduler) runRemote(ctx context.Context, job Job, slot Slot) (map[string]any, string, error) {
	if s.peers == nil {
		return nil, string(BackendRemoteMesh), errors.New("MESH_UNAVAILABLE")
	}
	capability, _ := job.Payload["capability"].(string)
	if strings.TrimSpace(capability) == "" {
		return nil, string(BackendRemoteMesh), errors.New("REMOTE_CAPABILITY_REQUIRED")
	}
	payload := map[string]any{}
	for k, v := range job.Payload {
		payload[k] = v
	}
	delete(payload, "capability")
	payload["octacore"] = map[string]any{
		"job_id": job.JobID,
		"correlation_id": job.CorrelationID,
		"source": string(job.Source),
		"target": string(slot.Slot),
		"kind": string(job.Kind),
	}
	if slot.Slot == G0 {
		return nil, string(BackendRemoteMesh), errors.New("G0_MUST_USE_SARA_AUTHORITY")
	}
	result, err := s.peers.CallWithCorrelation(ctx, slot.Nucleus, capability, payload, job.CorrelationID)
	if err != nil {
		return nil, string(BackendRemoteMesh), err
	}
	if result == nil {
		return nil, string(BackendRemoteMesh), errors.New("MESH_EMPTY_RESULT")
	}
	if payload, ok := result["payload"].(map[string]any); ok {
		return payload, string(BackendRemoteMesh), nil
	}
	return result, string(BackendRemoteMesh), nil
}

func (s *Scheduler) executeG0(ctx context.Context, job Job) (map[string]any, string, error) {
	if s.sara == nil || !s.sara.Configured() {
		return nil, "SARA_HTTP", errors.New("SARA_UNAVAILABLE")
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
		contextPayload, _ := job.Payload["context"].(map[string]any)
		out, err := s.sara.CycleWithContext(ctx, input, cycleID, job.CorrelationID, contextPayload)
		return out, "SARA_HTTP", err
	default:
		return nil, "SARA_HTTP", errors.New("G0_UNSUPPORTED_OPERATION")
	}
}

func (s *Scheduler) expired(now time.Time, ttl int64) bool {
	return ttl <= 0
}

func numericSlice(value any) ([]float64, error) {
	raw, ok := value.([]any)
	if ok {
		out := make([]float64, len(raw))
		for i, item := range raw {
			n, ok := item.(float64)
			if !ok {
				return nil, fmt.Errorf("values[%d] must be numeric", i)
			}
			out[i] = n
		}
		return out, nil
	}
	if typed, ok := value.([]float64); ok {
		return append([]float64(nil), typed...), nil
	}
	return nil, errors.New("payload.values must be numeric array")
}

func failed(job Job, code string, err error, latency, queueWait int64) Result {
	if err == nil {
		err = errors.New(code)
	}
	return Result{
		JobID: job.JobID,
		CorrelationID: job.CorrelationID,
		OK: false,
		Error: &Error{Code: code, Message: err.Error()},
		Metrics: Metrics{LatencyMS: latency, QueueWaitMS: queueWait},
	}
}

func classifyError(err error) string {
	msg := strings.TrimSpace(err.Error())
	switch {
	case strings.HasPrefix(msg, "SARA_"):
		return strings.SplitN(msg, ":", 2)[0]
	case strings.Contains(msg, "UNAVAILABLE"):
		return "DEPENDENCY_UNAVAILABLE"
	default:
		return "EXECUTION_ERROR"
	}
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

func makeVagus(typ, source, target string, priority int, ttl int64, correlation string, payload map[string]any) VagusEnvelope {
	return VagusEnvelope{
		VagusVersion: VagusVersion,
		MessageID: NewJobID(),
		CorrelationID: correlation,
		Source: source,
		Target: target,
		Priority: priority,
		TTL: ttl,
		Type: typ,
		Payload: payload,
	}
}
