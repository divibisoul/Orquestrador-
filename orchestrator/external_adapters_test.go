package orchestrator

import (
	"context"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
	"os"
	"strings"
	"testing"
)

func newExternalTestEngine(t *testing.T) (*Engine, ExternalAdapterRegistry) {
	t.Helper()
	n, _ := neural.New(4, .05)
	c, _ := prefrontal.New(.1, 8)
	g := supergpu.New(nil)
	e, err := New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewExternalAdapterRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if err := RegisterExternalAdapterOperations(e, r); err != nil {
		t.Fatal(err)
	}
	return e, r
}
func TestExternalAdapterManifestHasSixteenProviders(t *testing.T) {
	_, r := newExternalTestEngine(t)
	if len(r.byID) != 16 {
		t.Fatalf("providers=%d", len(r.byID))
	}
}
func TestExternalAdapterExecutionFailsClosedByDefault(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	old := os.Getenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH")
	t.Cleanup(func() { _ = os.Setenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH", old) })
	_ = os.Unsetenv("SOUL_EXTERNAL_EXECUTE_LANGGRAPH")
	_, err := e.Execute(context.Background(), "external.langgraph.execute@1.0.0", nil, map[string]string{"external_operation": "workflow.invoke", "input": "x"})
	if err == nil || !strings.Contains(err.Error(), "EXTERNAL_ADAPTER_DISABLED") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestExternalAdapterMetadataCannotOverrideExecutionContext(t *testing.T) {
	metadata := map[string]string{
		"provider":   "superagi",
		"mode":       "probe",
		"operation":  "agent.execute",
		"root":       "/tmp/untrusted-upstream",
		"metadata":   "untrusted-metadata-replacement",
		"prompt":     "preserve the actual task payload",
	}
	request := externalAdapterRequest(
		"langgraph",
		"execute",
		"workflow.invoke",
		"/repo/integrations/external/langgraph",
		metadata,
	)

	if request["provider"] != "langgraph" {
		t.Fatalf("provider identity overridden by request metadata: %#v", request["provider"])
	}
	if request["mode"] != "execute" {
		t.Fatalf("adapter mode overridden by request metadata: %#v", request["mode"])
	}
	if request["operation"] != "workflow.invoke" {
		t.Fatalf("operation overridden by request metadata: %#v", request["operation"])
	}
	if request["root"] != "/repo/integrations/external/langgraph" {
		t.Fatalf("pinned source root overridden by request metadata: %#v", request["root"])
	}
	if request["prompt"] != "preserve the actual task payload" {
		t.Fatalf("ordinary provider payload was lost: %#v", request["prompt"])
	}
	nested, ok := request["metadata"].(map[string]string)
	if !ok || nested["root"] != "/tmp/untrusted-upstream" {
		t.Fatalf("original caller metadata was not preserved as nested evidence: %#v", request["metadata"])
	}
}

func TestExternalAdapterProbeRegistered(t *testing.T) {
	e, _ := newExternalTestEngine(t)
	if _, err := e.Execute(context.Background(), "external.langgraph.probe@1.0.0", nil, nil); err != nil {
		t.Fatalf("probe failed: %v", err)
	}
}
