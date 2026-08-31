package orchestrator

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/divibisoul/Orquestrador-/state"
)

func TestDurableWorkflowRestore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "workflow.json")
	store, err := state.OpenDurable(path)
	if err != nil { t.Fatal(err) }
	e := NewEngineWithStore(2, store)
	steps := []Step{{ID:"one", Run:func(context.Context) error{return nil}}, {ID:"two", Run:func(context.Context) error{return nil}}}
	if err := e.CreateWorkflow("wf", steps); err != nil { t.Fatal(err) }
	if err := e.ExecuteStep(context.Background(), "wf"); err != nil { t.Fatal(err) }

	e2 := NewEngineWithStore(2, store)
	if err := e2.RestoreWorkflow("wf", steps); err != nil { t.Fatal(err) }
	w, err := e2.GetWorkflow("wf")
	if err != nil { t.Fatal(err) }
	if w.Current != 1 || w.Status != Pending { t.Fatalf("restored workflow=%+v", w) }
	if err := e2.ExecuteStep(context.Background(), "wf"); err != nil { t.Fatal(err) }
	status, err := e2.GetWorkflowStatus("wf"); if err != nil { t.Fatal(err) }
	if status != Completed { t.Fatalf("status=%s", status) }
}
