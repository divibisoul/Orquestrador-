package execution

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Task func(context.Context) error

type Executor struct {
	Workers int
	mu      sync.Mutex
	cb      circuitBreaker
	limits  map[string]*rateBucket
}

type rateBucket struct {
	tokens float64
	last   time.Time
	rate   float64
	cap    float64
}

type circuitState string

const (
	CircuitClosed   circuitState = "closed"
	CircuitOpen     circuitState = "open"
	CircuitHalfOpen circuitState = "half_open"
)

type circuitBreaker struct {
	state         circuitState
	successes     int
	failures      int
	windowStart   time.Time
	openedAt      time.Time
	cooldown      time.Duration
	window        time.Duration
	minSamples    int
	failureRatio  float64
	halfOpenProbe bool
}

func New(workers int) *Executor {
	if workers < 1 {
		workers = 1
	}
	return &Executor{
		Workers: workers,
		cb: circuitBreaker{
			state:        CircuitClosed,
			window:       10 * time.Second,
			cooldown:     30 * time.Second,
			minSamples:   4,
			failureRatio: 0.5,
		},
		limits: map[string]*rateBucket{},
	}
}

func (e *Executor) ExecuteParallel(ctx context.Context, tasks []Task) []error {
	if len(tasks) == 0 {
		return nil
	}
	jobs := make(chan Task)
	errs := make(chan error, len(tasks))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-jobs:
				if !ok {
					return
				}
				if t == nil {
					errs <- errors.New("nil task")
					continue
				}
				if err := t(ctx); err != nil {
					errs <- err
				}
			}
		}
	}

	workers := e.Workers
	if workers < 1 {
		workers = 1
	}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	for _, task := range tasks {
		select {
		case jobs <- task:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			close(errs)
			return collectErrors(errs)
		}
	}
	close(jobs)
	wg.Wait()
	close(errs)
	return collectErrors(errs)
}

func collectErrors(errs <-chan error) []error {
	out := make([]error, 0)
	for err := range errs {
		out = append(out, err)
	}
	return out
}

// ExecuteDistributed dispatches tasks through a caller-provided node dispatcher.
// The dispatcher remains responsible for actual remote transport; this method
// respects cancellation and rejects nil dispatchers instead of panicking.
func (e *Executor) ExecuteDistributed(ctx context.Context, tasks []Task, dispatch func(context.Context, Task) error) []error {
	if dispatch == nil {
		return []error{errors.New("nil dispatcher")}
	}
	out := make([]error, 0)
	for _, t := range tasks {
		select {
		case <-ctx.Done():
			out = append(out, ctx.Err())
			return out
		default:
		}
		if t == nil {
			out = append(out, errors.New("nil task"))
			continue
		}
		if err := dispatch(ctx, t); err != nil {
			out = append(out, err)
		}
	}
	return out
}

// CircuitBreaker records real successes and failures in a rolling window.
// Return value remains backwards-compatible: true means the circuit is open.
func (e *Executor) CircuitBreaker(success bool) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	b := &e.cb
	if b.windowStart.IsZero() {
		b.windowStart = now
	}

	if b.state == CircuitOpen {
		if now.Sub(b.openedAt) < b.cooldown {
			return true
		}
		b.state = CircuitHalfOpen
		b.halfOpenProbe = false
	}

	if b.state == CircuitHalfOpen {
		if b.halfOpenProbe {
			return true
		}
		b.halfOpenProbe = true
		if success {
			b.state = CircuitClosed
			b.successes = 0
			b.failures = 0
			b.windowStart = now
			b.halfOpenProbe = false
			return false
		}
		b.state = CircuitOpen
		b.openedAt = now
		b.halfOpenProbe = false
		return true
	}

	if now.Sub(b.windowStart) >= b.window {
		b.successes = 0
		b.failures = 0
		b.windowStart = now
	}
	if success {
		b.successes++
	} else {
		b.failures++
	}

	samples := b.successes + b.failures
	if samples >= b.minSamples && float64(b.failures)/float64(samples) >= b.failureRatio {
		b.state = CircuitOpen
		b.openedAt = now
		return true
	}
	return false
}

func (e *Executor) CircuitState() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return string(e.cb.state)
}

// Bulkhead isolates a workload to a bounded worker count.
func (e *Executor) Bulkhead(ctx context.Context, workers int, tasks []Task) []error {
	if workers < 1 {
		workers = 1
	}
	local := New(workers)
	return local.ExecuteParallel(ctx, tasks)
}

// RateLimiter implements a token bucket and returns whether a request is allowed.
func (e *Executor) RateLimiter(key string, limit int) bool {
	if limit <= 0 || key == "" {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	b := e.limits[key]
	if b == nil {
		b = &rateBucket{tokens: float64(limit), last: now, rate: float64(limit), cap: float64(limit)}
		e.limits[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * b.rate
	if b.tokens > b.cap {
		b.tokens = b.cap
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func RetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := time.Duration(1<<min(attempt-1, 6)) * time.Second
	if d > 60*time.Second {
		d = 60 * time.Second
	}
	return d
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var ErrCircuitOpen = errors.New("circuit breaker open")
