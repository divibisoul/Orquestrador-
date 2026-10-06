package orchestrator

import (
	"context"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
	"testing"
	"time"
)

func newEngine(t *testing.T, size int) (*Engine, *supergpu.Runtime) {
	t.Helper()
	n, err := neural.New(size, .1)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	return e, g
}
func TestEngineRuntime(t *testing.T) {
	e, _ := newEngine(t, 2)
	r, err := e.Execute(context.Background(), "compute.execute", []float64{2, 3}, map[string]string{"operation": "square"})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Payload) != 2 || r.Payload[0] != 4 {
		t.Fatal("orchestration result incorrect")
	}
	p, err := e.Execute(context.Background(), "cognitive.execute", []float64{2, 3}, map[string]string{"operation": "identity"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "ok" || len(p.Payload) != 2 {
		t.Fatal("cognitive pipeline failed")
	}
	if e.Status() != "ready" {
		t.Fatal("engine not ready")
	}
	if e.Health()["nucleus"] != "N07" {
		t.Fatal("wrong nucleus")
	}
	if err = e.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
func TestCancel(t *testing.T) {
	e, _ := newEngine(t, 1)
	started := make(chan struct{})
	if err := e.Register("wait@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		close(started)
		<-ctx.Done()
		return protocol.Result{TraceID: m.TraceID, Status: "cancelled", Error: ctx.Err().Error()}, ctx.Err()
	}); err != nil {
		t.Fatal(err)
	}
	m := protocol.NewMessage("N01", "N07", "command", "wait@1.0.0", nil)
	done := make(chan error, 1)
	go func() { _, err := e.Submit(context.Background(), m); done <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	if err := e.Cancel(m.TraceID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancellation did not propagate")
	}
	_ = e.Shutdown(context.Background())
}
func TestVersionedRouting(t *testing.T) {
	e, _ := newEngine(t, 1)
	versions := []string{"1.0.0", "2.0.0", "10.0.0"}
	for _, v := range versions {
		v := v
		if err := e.Register("test.operation@"+v, func(context.Context, protocol.Message) (protocol.Result, error) {
			return protocol.Result{Status: v}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	m := protocol.NewMessage("N01", "N07", "command", "test.operation@2.0.0", []float64{1})
	r, err := e.Submit(context.Background(), m)
	if err != nil || r.Status != "2.0.0" {
		t.Fatalf("explicit version did not route to v2: %v %v", r, err)
	}
	m.Operation = "test.operation"
	r, err = e.Submit(context.Background(), m)
	if err != nil || r.Status != "10.0.0" {
		t.Fatalf("unversioned route did not select highest semantic version: %v %v", r, err)
	}
	if len(e.Operations()) < 7 {
		t.Fatalf("expected versioned operation inventory, got %d", len(e.Operations()))
	}
	if len(N07Agents()) < 6 {
		t.Fatal("N07 agent surface incomplete")
	}
	if SOULTopology()["directional"] != 12 {
		t.Fatal("N07 topology is not six-peer bidirectional")
	}
	_ = e.Shutdown(context.Background())
}

func TestSemanticVersionRoutingHonorsPreReleasePrecedence(t *testing.T) {
	cases := []struct {
		higher string
		lower  string
	}{
		{"1.0.0", "1.0.0-beta"},
		{"1.0.0-beta", "1.0.0-alpha"},
		{"1.0.0-alpha.10", "1.0.0-alpha.2"},
		{"1.0.0-alpha.2", "1.0.0-alpha.1"},
		{"2.0.0-rc.1", "2.0.0-beta.99"},
	}
	for _, tc := range cases {
		if !semverGreater(tc.higher, tc.lower) {
			t.Fatalf("expected %s > %s", tc.higher, tc.lower)
		}
		if semverGreater(tc.lower, tc.higher) {
			t.Fatalf("comparison is not antisymmetric for %s and %s", tc.higher, tc.lower)
		}
	}

	if semverGreater("1.0.0+build.2", "1.0.0+build.1") || semverGreater("1.0.0+build.1", "1.0.0+build.2") {
		t.Fatal("build metadata must not affect SemVer precedence")
	}
}

func TestSplitOperationAcceptsBuildMetadataAndRejectsMalformedSemVer(t *testing.T) {
	name, version, err := splitOperation("model.route@1.2.3+linux.amd64")
	if err != nil {
		t.Fatal(err)
	}
	if name != "model.route" || version != "1.2.3+linux.amd64" {
		t.Fatalf("unexpected parsed operation: %q %q", name, version)
	}

	for _, operation := range []string{
		"model.route@01.2.3",
		"model.route@1.02.3",
		"model.route@1.2.03",
		"model.route@1.2.3-01",
		"model.route@1.2.3-",
	} {
		if _, _, err := splitOperation(operation); err == nil {
			t.Fatalf("expected malformed SemVer rejection: %s", operation)
		}
	}
}

func TestSemanticVersionHandlesVeryLargeNumericPreReleaseIdentifiers(t *testing.T) {
	large := "12345678901234567890123456789012345678901234567890"
	greater := "12345678901234567890123456789012345678901234567891"
	higher := "1.0.0-alpha." + greater
	lower := "1.0.0-alpha." + large
	if !semverGreater(higher, lower) {
		t.Fatalf("large numeric prerelease identifier comparison failed: %s <= %s", higher, lower)
	}
	if semverGreater(lower, higher) {
		t.Fatalf("large numeric prerelease comparison is not antisymmetric: %s > %s", lower, higher)
	}
}

func TestRequestRejectionsDoNotOpenExecutionBreaker(t *testing.T) {
	e, _ := newEngine(t, 1)
	e.failureThreshold = 2

	for i := 0; i < 5; i++ {
		message := protocol.NewMessage("N01", "N07", "command", "missing.operation@1.0.0", []float64{1})
		if _, err := e.Submit(context.Background(), message); err == nil {
			t.Fatal("unknown operation must be rejected")
		}
		if e.breakerOpen() {
			t.Fatalf("request rejection poisoned the execution breaker after attempt %d", i+1)
		}
	}
	if got := e.failures.Load(); got != 0 {
		t.Fatalf("request rejections must not increment execution failures, got %d", got)
	}

	if err := e.Register("failing.operation@1.0.0", func(context.Context, protocol.Message) (protocol.Result, error) {
		return protocol.Result{Status: "error"}, context.Canceled
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		message := protocol.NewMessage("N01", "N07", "command", "failing.operation@1.0.0", []float64{1})
		_, _ = e.Submit(context.Background(), message)
	}
	if !e.breakerOpen() {
		t.Fatal("real execution failures must still trip the execution breaker")
	}
	_ = e.Shutdown(context.Background())
}
