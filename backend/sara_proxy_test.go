package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func newSARARegistrationEngine(t *testing.T) *orchestrator.Engine {
	t.Helper()
	n, err := neural.New(4, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(0.1, 4)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpu.New(nil)
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestSARAProxyRegistrationOnlyWhenConfigured(t *testing.T) {
	e := newSARARegistrationEngine(t)

	if err := RegisterSARAOperations(e, NewSARAProxy(Config{})); err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{
		"sara.cycle@1.0.0",
		"sara.audit@1.0.0",
		"sara.regenerate@1.0.0",
		"sara.state@1.0.0",
		"sara.capabilities@1.0.0",
		"sara.trace@1.0.0",
	} {
		// Registration itself is deterministic and independent of network reachability.
		if !contains(e.Operations(), op) {
			t.Fatalf("operation %s was not registered", op)
		}
	}
}

func TestSARAProxyPropagatesCorrelationHeader(t *testing.T) {
    var gotCorrelation string
    var gotBody map[string]any
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        gotCorrelation = r.Header.Get("X-Correlation-ID")
        if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
            t.Fatalf("decode request: %v", err)
        }
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"cycle_id":"c-1","correlation_id":"corr-1"}`))
    }))
    defer server.Close()

    proxy := NewSARAProxy(Config{
        SARAServiceURL: server.URL,
        SARAServiceToken: "token",
    })
    out, err := proxy.Cycle(context.Background(), "input", "cycle-1", "corr-1")
    if err != nil {
        t.Fatal(err)
    }
    if gotCorrelation != "corr-1" {
        t.Fatalf("expected correlation header corr-1, got %q", gotCorrelation)
    }
    if gotBody["cycle_id"] != "cycle-1" {
        t.Fatalf("expected cycle_id cycle-1, got %#v", gotBody["cycle_id"])
    }
    if out["cycle_id"] != "c-1" {
        t.Fatalf("expected response cycle_id c-1, got %#v", out["cycle_id"])
    }
}


func TestSARAProxyLiveCycleIsOptIn(t *testing.T) {
	baseURL := strings.TrimSpace(os.Getenv("SARA_E2E_URL"))
	token := strings.TrimSpace(os.Getenv("SARA_E2E_TOKEN"))
	if baseURL == "" || token == "" {
		t.Skip("SARA_E2E_URL and SARA_E2E_TOKEN are required for live federation test")
	}

	proxy := NewSARAProxy(Config{
		SARAServiceURL:     baseURL,
		SARAServiceToken:   token,
		SARARequestTimeout: 30 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	out, err := proxy.Cycle(ctx, "promover autonomia e transparência", "n07-live-sara", "corr-live-sara")
	if err != nil {
		t.Fatal(err)
	}
	if out["cycle_id"] == nil {
		t.Fatalf("live SARA result missing cycle_id: %#v", out)
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

var _ = protocol.Message{}
