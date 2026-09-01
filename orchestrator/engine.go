package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/observability"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type Handler func(context.Context, protocol.Message) (protocol.Result, error)

type OperationRegistration struct {
	Name      string
	Version   string
	Handler   Handler
	Timeout   time.Duration
	RateLimit int
	Schema    json.RawMessage
	Metadata  map[string]string
}

type traceState struct {
	cancel  context.CancelFunc
	stage   string
	device  string
	started time.Time
}

type rateState struct {
	window time.Time
	count  int
}

type routeCacheEntry struct {
	handler   Handler
	expires   time.Time
	operation string
	version   string
}

type Engine struct {
	mu               sync.RWMutex
	handlers         map[string]OperationRegistration
	active           map[string]traceState
	routes           map[string]routeCacheEntry
	rates            map[string]rateState
	neural           *neural.Network
	cortex           *prefrontal.Cortex
	compute          *supergpu.Runtime
	running          atomic.Bool
	sequence         atomic.Uint64
	failures         atomic.Uint64
	metrics          *observability.Metrics
	logger           *observability.Logger
	rateLimit        int
	routeTTL         time.Duration
	failureThreshold uint64
	breakerUntil     time.Time
	breakerMu        sync.Mutex
}

var semverRx = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?$`)

func splitOperation(operation string) (string, string, error) {
	operation = strings.TrimSpace(operation)
	if operation == "" {
		return "", "", errors.New("operation is required")
	}
	name, version := operation, ""
	if index := strings.LastIndex(operation, "@"); index > 0 {
		name = strings.TrimSpace(operation[:index])
		version = strings.TrimSpace(operation[index+1:])
	}
	if name == "" {
		return "", "", errors.New("operation name is required")
	}
	if version != "" && !semverRx.MatchString(version) {
		return "", "", errors.New("operation version must semantic version")
	}
	return name, version, nil
}

func semverGreater(a, b string) bool {
	parse := func(value string) [3]int {
		base := strings.SplitN(value, "-", 2)[0]
		parts := strings.SplitN(base, ".", 3)
		var out [3]int
		for i := range out {
			if i >= len(parts) {
				continue
			}
			for _, ch := range parts[i] {
				if ch < '0' || ch > '9' {
					break
				}
				out[i] = out[i]*10 + int(ch-'0')
			}
		}
		return out
	}
	aa, bb := parse(a), parse(b)
	for i := range aa {
		if aa[i] != bb[i] {
			return aa[i] > bb[i]
		}
	}
	return a > b
}

func New(n *neural.Network, c *prefrontal.Cortex, g *supergpu.Runtime) (*Engine, error) {
	if n == nil || c == nil || g == nil {
		return nil, errors.New("all nucleus services are required")
	}
	e := &Engine{
		handlers:         make(map[string]OperationRegistration),
		active:           make(map[string]traceState),
		routes:           make(map[string]routeCacheEntry),
		rates:            make(map[string]rateState),
		neural:           n,
		cortex:           c,
		compute:          g,
		metrics:          &observability.Metrics{},
		logger:           observability.NewLogger("N07.orchestrator"),
		rateLimit:        60,
		routeTTL:         30 * time.Second,
		failureThreshold: 5,
	}
	e.running.Store(true)
	if err := e.registerBuiltins(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) Register(operation string, handler Handler) error {
	name, version, err := splitOperation(operation)
	if err != nil {
		return err
	}
	if handler == nil {
		return errors.New("handler is required")
	}
	if version == "" {
		version = "1.0.0"
	}
	key := name + "@" + version
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.handlers[key]; exists {
		return errors.New("operation/version already registered")
	}
	e.handlers[key] = OperationRegistration{
		Name:      name,
		Version:   version,
		Handler:   handler,
		Timeout:   15 * time.Second,
		RateLimit: e.rateLimit,
		Metadata:  map[string]string{"capability": name},
	}
	for cacheKey := range e.routes {
		if cacheKey == key || strings.HasPrefix(cacheKey, name+"@") {
			delete(e.routes, cacheKey)
		}
	}
	return nil
}

func (e *Engine) Operations() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	ops := make([]string, 0, len(e.handlers))
	for key, registration := range e.handlers {
		if registration.Handler != nil {
			ops = append(ops, key)
		}
	}
	sort.Strings(ops)
	return ops
}

func (e *Engine) resolveRegistrationLocked(name, version string) (OperationRegistration, bool) {
	if version != "" {
		registration, ok := e.handlers[name+"@"+version]
		return registration, ok && registration.Handler != nil
	}
	var best OperationRegistration
	found := false
	for _, candidate := range e.handlers {
		if candidate.Name != name || candidate.Handler == nil {
			continue
		}
		if !found || semverGreater(candidate.Version, best.Version) {
			best = candidate
			found = true
		}
	}
	return best, found
}

func (e *Engine) Route(message protocol.Message) (Handler, error) {
	if err := message.Validate(); err != nil {
		return nil, err
	}
	name, version, err := splitOperation(message.Operation)
	if err != nil {
		return nil, err
	}
	cacheKey := name + "@" + version
	now := time.Now()
	e.mu.RLock()
	cached, ok := e.routes[cacheKey]
	e.mu.RUnlock()
	if ok && now.Before(cached.expires) && cached.handler != nil {
		return cached.handler, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if cached, ok := e.routes[cacheKey]; ok && now.Before(cached.expires) && cached.handler != nil {
		return cached.handler, nil
	}
	registration, ok := e.resolveRegistrationLocked(name, version)
	if !ok {
		return nil, errors.New("no route for operation: " + message.Operation)
	}
	e.routes[cacheKey] = routeCacheEntry{
		handler:   registration.Handler,
		expires:   now.Add(e.routeTTL),
		operation: registration.Name,
		version:   registration.Version,
	}
	return registration.Handler, nil
}

func (e *Engine) registration(operation string) (OperationRegistration, error) {
	name, version, err := splitOperation(operation)
	if err != nil {
		return OperationRegistration{}, err
	}
	e.mu.RLock()
	registration, ok := e.resolveRegistrationLocked(name, version)
	e.mu.RUnlock()
	if !ok || registration.Handler == nil {
		return OperationRegistration{}, errors.New("registered operation unavailable")
	}
	return registration, nil
}

func (e *Engine) allow(source string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	state := e.rates[strings.TrimSpace(source)]
	if state.window.IsZero() || now.Sub(state.window) >= time.Second {
		state = rateState{window: now}
	}
	state.count++
	e.rates[strings.TrimSpace(source)] = state
	return state.count <= e.rateLimit
}

func (e *Engine) breakerOpen() bool {
	e.breakerMu.Lock()
	defer e.breakerMu.Unlock()
	return time.Now().Before(e.breakerUntil)
}

func (e *Engine) recordFailure() {
	count := e.failures.Add(1)
	if count < e.failureThreshold {
		return
	}
	e.breakerMu.Lock()
	e.breakerUntil = time.Now().Add(10 * time.Second)
	e.breakerMu.Unlock()
}

func (e *Engine) recordSuccess() { e.failures.Store(0) }

func (e *Engine) Submit(ctx context.Context, message protocol.Message) (protocol.Result, error) {
	start := time.Now()
	e.sequence.Add(1)
	e.metrics.Requests.Add(1)
	e.metrics.InFlight.Add(1)
	defer e.metrics.InFlight.Add(-1)

	reject := func(state string, err error) (protocol.Result, error) {
		e.metrics.Observe(start, err, state == "cancelled")
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07", Target: message.Source, Status: state, Error: err.Error()}, err
	}

	if !e.running.Load() {
		return reject("rejected", errors.New("orchestrator stopped"))
	}
	if ctx == nil {
		return reject("rejected", errors.New("context is nil"))
	}
	if e.breakerOpen() {
		return reject("busy", errors.New("circuit breaker open"))
	}
	if err := message.Validate(); err != nil {
		return reject("rejected", err)
	}
	if !e.allow(message.Source) {
		return reject("rate_limited", errors.New("rate limit exceeded"))
	}

	handler, err := e.Route(message)
	if err != nil {
		e.recordFailure()
		return reject("rejected", err)
	}
	registration, err := e.registration(message.Operation)
	if err != nil {
		e.recordFailure()
		return reject("rejected", err)
	}

	timeout := registration.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if !message.Deadline.IsZero() {
		if deadline, ok := runCtx.Deadline(); !ok || message.Deadline.Before(deadline) {
			var deadlineCancel context.CancelFunc
			runCtx, deadlineCancel = context.WithDeadline(runCtx, message.Deadline)
			defer deadlineCancel()
		}
	}

	runCtx, span := observability.Start(runCtx, message.Operation, message.TraceID, len(message.Payload))
	e.mu.Lock()
	if _, exists := e.active[message.TraceID]; exists {
		e.mu.Unlock()
		err := errors.New("duplicate active trace id")
		observability.End(span, err)
		return reject("rejected", err)
	}
	e.active[message.TraceID] = traceState{cancel: cancel, stage: "routed", started: start}
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		delete(e.active, message.TraceID)
		e.mu.Unlock()
	}()

	result, err := handler(runCtx, message)
	result.TraceID = message.TraceID
	result.CorrelationID = message.CorrelationID
	if result.Source == "" {
		result.Source = "N07"
	}
	if result.Target == "" {
		result.Target = message.Source
	}
	cancelled := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
	observability.End(span, err)
	e.metrics.Observe(start, err, cancelled)
	if err != nil {
		e.recordFailure()
		if cancelled {
			result.Status = "cancelled"
		} else if result.Status == "" {
			result.Status = "error"
		}
		return result, err
	}
	e.recordSuccess()
	if result.Status == "" {
		result.Status = "ok"
	}
	return result, nil
}

func (e *Engine) Execute(ctx context.Context, operation string, payload []float64, metadata map[string]string) (protocol.Result, error) {
	message := protocol.NewMessage("N07", "N07", "command", operation, payload)
	message.Metadata = metadata
	if metadata != nil {
		if schema := strings.TrimSpace(metadata["schema"]); schema != "" {
			if err := validateSchema(schema, payload); err != nil {
				return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07", Target: "N07", Status: "rejected", Error: err.Error()}, err
			}
		}
	}
	return e.Submit(ctx, message)
}

func (e *Engine) Cancel(traceID string) error {
	if strings.TrimSpace(traceID) == "" {
		return errors.New("trace id is required")
	}
	e.mu.RLock()
	state, ok := e.active[traceID]
	e.mu.RUnlock()
	if !ok {
		return errors.New("trace id is not active")
	}
	state.cancel()
	return nil
}

func (e *Engine) Status() string {
	if e.running.Load() {
		return "ready"
	}
	return "stopped"
}

func (e *Engine) Health() map[string]any {
	e.mu.RLock()
	handlers, active := len(e.handlers), len(e.active)
	e.mu.RUnlock()
	circuit := "closed"
	if e.breakerOpen() {
		circuit = "open"
	}
	return map[string]any{
		"nucleus":       "N07",
		"status":        e.Status(),
		"handlers":      handlers,
		"active_traces": active,
		"metrics":       e.metrics.Snapshot(),
		"circuit":       map[string]any{"state": circuit},
		"neural":        e.neural.Health(),
		"prefrontal":    e.cortex.Health(),
		"compute":       e.compute.Health(),
	}
}

func (e *Engine) Stats() map[string]any {
	health := e.Health()
	return map[string]any{
		"sequence":      e.sequence.Load(),
		"status":        e.Status(),
		"active_traces": health["active_traces"],
		"circuit":       health["circuit"],
		"metrics":       health["metrics"],
		"timestamp":     time.Now().UTC(),
	}
}

func (e *Engine) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	e.running.Store(false)
	e.mu.Lock()
	for id, state := range e.active {
		state.cancel()
		delete(e.active, id)
	}
	e.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return e.compute.Shutdown()
}

func (e *Engine) registerBuiltins() error {
	if err := e.Register("neural.forward@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		e.setStage(message.TraceID, "neural", "")
		values, err := e.neural.Forward(ctx, message.Payload)
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.neural", Target: message.Source, Status: status(err), Payload: values, Error: errorText(err)}, err
	}); err != nil {
		return err
	}

	if err := e.Register("neural.learn@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		e.setStage(message.TraceID, "neural.learn", "")
		select {
		case <-ctx.Done():
			return protocol.Result{}, ctx.Err()
		default:
		}
		half := len(message.Payload) / 2
		if half == 0 || half*2 != len(message.Payload) {
			return protocol.Result{}, errors.New("learn payload must contain input and target halves")
		}
		err := e.neural.Learn(message.Payload[:half], message.Payload[half:])
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.neural", Target: message.Source, Status: status(err), Error: errorText(err)}, err
	}); err != nil {
		return err
	}

	if err := e.Register("compute.execute@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		device, err := e.compute.Select(message.Metadata["device"])
		if err != nil {
			return protocol.Result{}, err
		}
		if err := e.compute.Reserve(device.ID, message.TraceID); err != nil {
			return protocol.Result{}, err
		}
		defer e.compute.Release(device.ID, message.TraceID)
		e.setStage(message.TraceID, "compute", device.ID)
		op := strings.TrimSpace(message.Metadata["operation"])
		if op == "" {
			return protocol.Result{}, errors.New("metadata.operation is required")
		}
		values, err := e.compute.Execute(ctx, device, op, message.Payload)
		return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Source: "N07.compute", Target: message.Source, Status: status(err), Payload: values, Error: errorText(err)}, err
	}); err != nil {
		return err
	}

	if err := e.Register("cognitive.execute@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		e.setStage(message.TraceID, "neural", "")
		encoded, err := e.neural.Forward(ctx, message.Payload)
		if err != nil {
			return protocol.Result{}, err
		}
		energy := 0.0
		for _, value := range encoded {
			energy += math.Abs(value)
		}
		utility := 0.0
		if len(encoded) > 0 {
			utility = energy / float64(len(encoded))
		}
		candidate := prefrontal.Candidate{ID: message.TraceID, Utility: utility, Cost: 0.05, Risk: 0.02}
		selected, err := e.cortex.Select([]prefrontal.Candidate{candidate})
		if err != nil {
			return protocol.Result{}, err
		}
		if _, err := e.cortex.Commit(selected, "neural-output-approved"); err != nil {
			return protocol.Result{}, err
		}
		device, err := e.compute.Select(message.Metadata["device"])
		if err != nil {
			return protocol.Result{}, err
		}
		if err := e.compute.Reserve(device.ID, message.TraceID); err != nil {
			return protocol.Result{}, err
		}
		defer e.compute.Release(device.ID, message.TraceID)
		e.setStage(message.TraceID, "compute", device.ID)
		op := strings.TrimSpace(message.Metadata["operation"])
		if op == "" {
			op = "identity"
		}
		output, err := e.compute.Execute(ctx, device, op, encoded)
		return protocol.Result{
			TraceID:       message.TraceID,
			CorrelationID: message.CorrelationID,
			Source:        "N07.pipeline",
			Target:        message.Source,
			Status:        status(err),
			Payload:       output,
			Metadata:      map[string]string{"decision": "approved", "device": device.ID},
			Error:         errorText(err),
		}, err
	}); err != nil {
		return err
	}
	return nil
}

func (e *Engine) setStage(traceID, stage, device string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	state, ok := e.active[traceID]
	if !ok {
		return
	}
	state.stage = stage
	state.device = device
	e.active[traceID] = state
}

func validateSchema(schema string, payload []float64) error {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return nil
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(schema), &raw); err != nil {
		return errors.New("invalid schema JSON")
	}
	if value, ok := raw["type"].(string); ok && value != "array" {
		return errors.New("payload schema type must be array")
	}
	if maximum, ok := raw["maxItems"].(float64); ok && float64(len(payload)) > maximum {
		return errors.New("payload exceeds schema maxItems")
	}
	if minimum, ok := raw["minItems"].(float64); ok && float64(len(payload)) < minimum {
		return errors.New("payload below schema minItems")
	}
	return nil
}

func status(err error) string {
	if err != nil {
		return "error"
	}
	return "ok"
}

func errorText(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
