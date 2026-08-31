package learning

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

func TestLearningAdapter(t *testing.T) {
	e := New(superagi.NewRuntime())
	if err := e.TrainOnline(context.Background(), []string{"a", "b"}); err != nil { t.Fatal(err) }
	if err := e.FineTuneLoRA(context.Background(), "adapter", "delta"); err != nil { t.Fatal(err) }
	if err := e.ActivateLoRA(context.Background(), "adapter"); err != nil { t.Fatal(err) }
	if len(e.Snapshot(2)) != 2 { t.Fatalf("snapshot length mismatch") }
}
