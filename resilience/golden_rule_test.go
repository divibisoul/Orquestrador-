package resilience

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestReactiveRetryBackoffAndFallback(t *testing.T) {
	calls := 0
	sleeps := make([]time.Duration, 0, 2)
	policy := RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   5 * time.Millisecond,
		MaxDelay:    20 * time.Millisecond,
		Sleep: func(_ context.Context, delay time.Duration) error {
			sleeps = append(sleeps, delay)
			return nil
		},
	}
	engine := NewEngine(Config{Retry: policy, Breaker: NewCircuitBreaker(3, time.Second)})
	output, report, err := engine.ExecuteCompute(
		context.Background(),
		"supergpu.execute",
		SeverityHigh,
		ChaosOperation(3, errors.New("gpu transport failed"), func(context.Context) ([]float64, error) { return []float64{4, 9}, nil }),
		func(context.Context) ([]float64, error) { return []float64{4, 9}, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(output) != 2 || output[0] != 4 || output[1] != 9 {
		t.Fatalf("unexpected fallback output: %#v", output)
	}
	if !report.FallbackUsed || report.State != StateDegraded {
		t.Fatalf("expected degraded fallback report, got %#v", report)
	}
	if calls != 0 {
		t.Fatal("unused probe counter indicates an invalid test setup")
	}
	if len(sleeps) != 2 || sleeps[0] != 5*time.Millisecond || sleeps[1] != 10*time.Millisecond {
		t.Fatalf("unexpected exponential backoff: %#v", sleeps)
	}
	if len(engine.Faults()) != 1 {
		t.Fatalf("expected one primary fault record, got %d", len(engine.Faults()))
	}
}

func TestCorrectiveCircuitBreakerAndForensicEvidence(t *testing.T) {
	engine := NewEngine(Config{Retry: RetryPolicy{MaxAttempts: 1}, Breaker: NewCircuitBreaker(1, time.Hour)})
	_, report, err := engine.ExecuteCompute(context.Background(), "supergpu.execute", SeverityCritical, func(context.Context) ([]float64, error) { return nil, errors.New("synthetic hardware failure") }, nil)
	if err == nil {
		t.Fatal("expected primary failure")
	}
	if report.State != StateFailed || engine.CircuitState() != BreakerOpen {
		t.Fatalf("expected failed/open state, report=%#v circuit=%s", report, engine.CircuitState())
	}
	faults := engine.Faults()
	if len(faults) != 1 || faults[0].Stack == "" || faults[0].Signature != "compute.primary_failure.circuit_opened" {
		t.Fatalf("forensic evidence incomplete: %#v", faults)
	}
	if !strings.Contains(engine.Metrics().Prometheus(), "soul_resilience_faults_total") {
		t.Fatal("Prometheus resilience metric missing")
	}
}

func TestGenerativeChaosInjectionAndMetricLearning(t *testing.T) {
	engine := NewEngine(Config{
		Retry:   RetryPolicy{MaxAttempts: 1},
		Breaker: NewCircuitBreaker(10, time.Second),
	})
	for i := 0; i < 3; i++ {
		_, report, err := engine.ExecuteCompute(
			context.Background(),
			"peer.call",
			SeverityMedium,
			func(context.Context) ([]float64, error) { return nil, errors.New("peer unavailable") },
			func(context.Context) ([]float64, error) { return []float64{1}, nil },
		)
		if err != nil || !report.FallbackUsed {
			t.Fatalf("expected controlled fallback on iteration %d: report=%#v err=%v", i, report, err)
		}
	}
	if count := engine.Faults(); len(count) != 3 {
		t.Fatalf("expected three learned fault records, got %d", len(count))
	}
	prom := engine.Metrics().Prometheus()
	if !strings.Contains(prom, `signature="compute.primary_failure"`) {
		t.Fatalf("missing fault signature metric: %s", prom)
	}
}

type testRunbook struct {
	called bool
}

func (r *testRunbook) ID() string { return "test-remediation" }
func (r *testRunbook) Execute(context.Context, Fault) error {
	r.called = true
	return nil
}

func TestGenerativeRunbookArmedOnlyAndSeverityBounded(t *testing.T) {
	runbook := &testRunbook{}
	engine := NewEngine(Config{
		Retry:     RetryPolicy{MaxAttempts: 1},
		Breaker:   NewCircuitBreaker(10, time.Second),
		Store:     &FaultStore{},
		Metrics:   NewMetrics(),
		Runbooks:  []Runbook{runbook},
		Remediation: RemediationPolicy{
			Armed:                    true,
			MinimumOccurrences:       1,
			MaximumAutomaticSeverity: SeverityLow,
		},
	})
	_, _, err := engine.ExecuteCompute(context.Background(), "safe-operation", SeverityLow, func(context.Context) ([]float64, error) { return nil, errors.New("safe fault") }, func(context.Context) ([]float64, error) { return []float64{1}, nil })
	if err != nil {
		t.Fatal(err)
	}
	if !runbook.called {
		t.Fatal("armed runbook was not executed")
	}
}
