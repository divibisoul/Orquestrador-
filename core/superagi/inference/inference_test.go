package inference

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

type provider struct{}
func (provider) Generate(context.Context, string) (string, error) { return "ok", nil }

func TestInferenceAdapter(t *testing.T) {
	r := superagi.NewRuntime().WithProvider(provider{})
	e := New(r)
	got, err := e.Generate(context.Background(), "test-model", "hello")
	if err != nil || got != "ok" { t.Fatalf("got=%q err=%v", got, err) }
}
