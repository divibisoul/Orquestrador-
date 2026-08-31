package soul

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/divibisoul/Orquestrador-/mesh"
    "github.com/divibisoul/Orquestrador-/mesh/transport"
)

func TestRegisterDefaultsAndDispatch(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost { t.Fatalf("method=%s", r.Method) }
        if got := r.Header.Get("X-Soul-Event-Type"); got != "soul.task.request" { t.Fatalf("event type=%q", got) }
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"event_id":"reply-1","event_type":"soul.task.response","trace_id":"trace-1","source":"N06","target":"orquestrador","timestamp":1}`))
    }))
    defer srv.Close()

    endpoints := map[NucleusID]string{N06: srv.URL}
    fabric := NewFabric(nil, mesh.NewRegistry(), transport.Client{})
    if err := fabric.RegisterDefaults(endpoints); err != nil { t.Fatal(err) }
    if got := len(fabric.Snapshot()); got != 6 { t.Fatalf("nuclei=%d want 6", got) }

    out, err := fabric.Dispatch(context.Background(), N06, "trace-1", "test", map[string]interface{}{"k": "v"})
    if err != nil { t.Fatal(err) }
    if out.TraceID != "trace-1" { t.Fatalf("trace=%q", out.TraceID) }
}

func TestValidateRejectsUnknownNucleus(t *testing.T) {
    if err := validateNucleus(Nucleus{ID: "N07", Repository: "x"}); err == nil { t.Fatal("expected rejection") }
}
