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

func TestEngineRuntime(t *testing.T) {
	n, _ := neural.New(2, .1)
	c, _ := prefrontal.New(.1, 4)
	g := supergpu.New(nil)
	g.Discover()
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
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
	n, _ := neural.New(1, .1)
	c, _ := prefrontal.New(.1, 2)
	g := supergpu.New(nil)
	g.Discover()
	e, _ := New(n, c, g)
	started := make(chan struct{})
	_ = e.Register("wait", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		close(started)
		<-ctx.Done()
		return protocol.Result{TraceID: m.TraceID, Status: "cancelled", Error: ctx.Err().Error()}, ctx.Err()
	})
	m := protocol.NewMessage("N01", "N07", "command", "wait", nil)
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
