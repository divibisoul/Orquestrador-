package resilience

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
)

type Severity uint8

const (
	SeverityCritical Severity = 1
	SeverityHigh     Severity = 2
	SeverityMedium   Severity = 3
	SeverityLow      Severity = 4
)

type State string

const (
	StateHealthy       State = "HEALTHY"
	StateDegraded      State = "DEGRADED"
	StateFailed        State = "FAILED"
	StateNotConfigured State = "NOT_CONFIGURED"
)

type Layer string

const (
	LayerReactive   Layer = "REACTIVE"
	LayerCorrective Layer = "CORRECTIVE"
	LayerGenerative Layer = "GENERATIVE"
)

type Fault struct {
	ID            string
	Signature     string
	Operation     string
	Severity      Severity
	Layer         Layer
	State         State
	Attempt       int
	Error         string
	Stack         string
	OccurredAt    time.Time
	RemediationID string
}

type FaultStore struct {
	mu     sync.Mutex
	faults []Fault
}

func (s *FaultStore) Record(fault Fault) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.faults = append(s.faults, fault)
}

func (s *FaultStore) Snapshot() []Fault {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Fault(nil), s.faults...)
}

func (s *FaultStore) Count(signature string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, fault := range s.faults {
		if fault.Signature == signature {
			count++
		}
	}
	return count
}

type Metrics struct {
	mu     sync.Mutex
	counts map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{counts: make(map[string]int64)}
}

func metricKey(signature, state string) string {
	return sanitize(signature) + ",state=" + sanitize(state)
}

func sanitize(value string) string {
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == ':' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func (m *Metrics) Inc(signature string, state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[metricKey(signature, string(state))]++
}

func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]int64, len(m.counts))
	for key, value := range m.counts {
		out[key] = value
	}
	return out
}

func (m *Metrics) Prometheus() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.counts))
	for key := range m.counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# TYPE soul_resilience_faults_total counter\n")
	for _, key := range keys {
		parts := strings.SplitN(key, ",state=", 2)
		signature := parts[0]
		state := "unknown"
		if len(parts) == 2 {
			state = parts[1]
		}
		fmt.Fprintf(&b, "soul_resilience_faults_total{signature=\"%s\",state=\"%s\"} %d\n", signature, state, m.counts[key])
	}
	return b.String()
}

type BreakerState string

const (
	BreakerClosed   BreakerState = "CLOSED"
	BreakerOpen     BreakerState = "OPEN"
	BreakerHalfOpen BreakerState = "HALF_OPEN"
)

type CircuitBreaker struct {
	mu        sync.Mutex
	state     BreakerState
	failures  int
	threshold int
	cooldown  time.Duration
	openedAt  time.Time
	now       func() time.Time
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold < 1 {
		threshold = 1
	}
	if cooldown <= 0 {
		cooldown = time.Second
	}
	return &CircuitBreaker{state: BreakerClosed, threshold: threshold, cooldown: cooldown, now: time.Now}
}

func (b *CircuitBreaker) State() BreakerState {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refreshLocked()
	return b.state
}

func (b *CircuitBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refreshLocked()
	return b.state != BreakerOpen
}

func (b *CircuitBreaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = BreakerClosed
	b.failures = 0
	b.openedAt = time.Time{}
}

func (b *CircuitBreaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refreshLocked()
	b.failures++
	if b.failures >= b.threshold {
		b.state = BreakerOpen
		b.openedAt = b.now()
	}
}

func (b *CircuitBreaker) refreshLocked() {
	if b.state == BreakerOpen && !b.openedAt.IsZero() && b.now().Sub(b.openedAt) >= b.cooldown {
		b.state = BreakerHalfOpen
	}
}

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Sleep       func(context.Context, time.Duration) error
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond, MaxDelay: 100 * time.Millisecond, Sleep: sleep}
}

func sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryDelay(policy RetryPolicy, retryIndex int) time.Duration {
	if retryIndex <= 0 || policy.BaseDelay <= 0 {
		return 0
	}
	delay := policy.BaseDelay
	for i := 1; i < retryIndex; i++ {
		if policy.MaxDelay > 0 && delay >= policy.MaxDelay/2 {
			return policy.MaxDelay
		}
		delay *= 2
	}
	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		return policy.MaxDelay
	}
	return delay
}

func ExecuteWithRetry(ctx context.Context, policy RetryPolicy, operation func(context.Context) error) (int, error) {
	if ctx == nil {
		return 0, errors.New("resilience: context is nil")
	}
	if operation == nil {
		return 0, errors.New("resilience: operation is nil")
	}
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}
	if policy.Sleep == nil {
		policy.Sleep = sleep
	}
	var lastErr error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return attempt - 1, err
		}
		lastErr = operation(ctx)
		if lastErr == nil {
			return attempt, nil
		}
		if errors.Is(lastErr, context.Canceled) || errors.Is(lastErr, context.DeadlineExceeded) {
			return attempt, lastErr
		}
		if attempt < policy.MaxAttempts {
			if err := policy.Sleep(ctx, retryDelay(policy, attempt)); err != nil {
				return attempt, err
			}
		}
	}
	return policy.MaxAttempts, lastErr
}

type Runbook interface {
	ID() string
	Execute(context.Context, Fault) error
}

type RemediationPolicy struct {
	Armed                    bool
	MinimumOccurrences       int
	MaximumAutomaticSeverity Severity
}

type Config struct {
	Retry       RetryPolicy
	Breaker     *CircuitBreaker
	Remediation RemediationPolicy
	Runbooks    []Runbook
	Store       *FaultStore
	Metrics     *Metrics
}

type ExecutionReport struct {
	State          State
	Attempts       int
	FallbackUsed   bool
	CircuitState   BreakerState
	FaultSignature string
	FaultCount     int
	RemediationID  string
}

type Engine struct {
	cfg Config
}

func NewEngine(cfg Config) *Engine {
	if cfg.Retry.MaxAttempts <= 0 {
		cfg.Retry = DefaultRetryPolicy()
	}
	if cfg.Breaker == nil {
		cfg.Breaker = NewCircuitBreaker(3, time.Second)
	}
	if cfg.Store == nil {
		cfg.Store = &FaultStore{}
	}
	if cfg.Metrics == nil {
		cfg.Metrics = NewMetrics()
	}
	return &Engine{cfg: cfg}
}

func (e *Engine) ExecuteCompute(ctx context.Context, op string, severity Severity, primary func(context.Context) ([]float64, error), fallback func(context.Context) ([]float64, error)) ([]float64, ExecutionReport, error) {
	if primary == nil {
		return nil, ExecutionReport{State: StateFailed}, errors.New("resilience: primary operation is nil")
	}
	if !e.cfg.Breaker.Allow() {
		report := ExecutionReport{State: StateDegraded, CircuitState: e.cfg.Breaker.State(), FaultSignature: "circuit.open"}
		e.recordFault(op, severity, LayerReactive, report.FaultSignature, errors.New("circuit breaker open"), 0)
		return nil, report, errors.New("resilience: circuit breaker open")
	}

	var output []float64
	attempts, primaryErr := ExecuteWithRetry(ctx, e.cfg.Retry, func(callCtx context.Context) error {
		value, err := primary(callCtx)
		if err != nil {
			return err
		}
		output = value
		return nil
	})
	if primaryErr == nil {
		e.cfg.Breaker.Success()
		return output, ExecutionReport{State: StateHealthy, Attempts: attempts, CircuitState: e.cfg.Breaker.State()}, nil
	}

	e.cfg.Breaker.Failure()
	signature := "compute.primary_failure"
	if !e.cfg.Breaker.Allow() {
		signature = "compute.primary_failure.circuit_opened"
	}
	count := e.recordFault(op, severity, LayerReactive, signature, primaryErr, attempts)

	if fallback != nil && ctx.Err() == nil {
		fallbackOutput, fallbackErr := fallback(ctx)
		if fallbackErr == nil {
			report := ExecutionReport{State: StateDegraded, Attempts: attempts, FallbackUsed: true, CircuitState: e.cfg.Breaker.State(), FaultSignature: signature, FaultCount: count}
			return fallbackOutput, report, nil
		}
		e.recordFault(op, severity, LayerCorrective, "compute.fallback_failure", fallbackErr, 1)
		return nil, ExecutionReport{State: StateFailed, Attempts: attempts, CircuitState: e.cfg.Breaker.State(), FaultSignature: signature, FaultCount: count}, errors.Join(primaryErr, fallbackErr)
	}

	return nil, ExecutionReport{State: StateFailed, Attempts: attempts, CircuitState: e.cfg.Breaker.State(), FaultSignature: signature, FaultCount: count}, primaryErr
}

func (e *Engine) recordFault(operation string, severity Severity, layer Layer, signature string, err error, attempt int) int {
	fault := Fault{ID: fmt.Sprintf("fault-%d", time.Now().UnixNano()), Signature: signature, Operation: operation, Severity: severity, Layer: layer, State: StateFailed, Attempt: attempt, Error: err.Error(), Stack: string(debug.Stack()), OccurredAt: time.Now().UTC()}
	countBefore := e.cfg.Store.Count(signature)
	remediationID := ""
	if e.cfg.Remediation.Armed && countBefore+1 >= e.cfg.Remediation.MinimumOccurrences && severity <= e.cfg.Remediation.MaximumAutomaticSeverity {
		for _, runbook := range e.cfg.Runbooks {
			if err := runbook.Execute(context.Background(), fault); err == nil {
				remediationID = runbook.ID()
				e.cfg.Metrics.Inc("remediation."+remediationID, StateHealthy)
				break
			}
		}
	}
	fault.RemediationID = remediationID
	e.cfg.Store.Record(fault)
	e.cfg.Metrics.Inc(signature, StateFailed)
	return countBefore + 1
}

func (e *Engine) Faults() []Fault { return e.cfg.Store.Snapshot() }
func (e *Engine) Metrics() *Metrics { return e.cfg.Metrics }
func (e *Engine) CircuitState() BreakerState { return e.cfg.Breaker.State() }

// ChaosOperation returns a bounded failure injector for controlled tests/staging.
func ChaosOperation(failures int, err error, success func(context.Context) ([]float64, error)) func(context.Context) ([]float64, error) {
	remaining := failures
	if err == nil {
		err = errors.New("chaos injected failure")
	}
	return func(ctx context.Context) ([]float64, error) {
		if remaining > 0 {
			remaining--
			return nil, err
		}
		if success == nil {
			return nil, nil
		}
		return success(ctx)
	}
}
