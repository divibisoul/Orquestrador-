package superagi

import (
	"context"
	"testing"
)

type fakeProvider struct{}
func (fakeProvider) Generate(context.Context, string) (string, error) { return "generated", nil }
func (fakeProvider) Embed(context.Context, string) ([]float64, error) { return []float64{1, 2, 3}, nil }

func TestRuntimeUsesProvider(t *testing.T) {
	r := NewRuntime().WithProvider(fakeProvider{})
	out, err := r.GenerateText(context.Background(), "hello")
	if err != nil || out != "generated" { t.Fatalf("out=%q err=%v", out, err) }
	v, err := r.GenerateEmbedding(context.Background(), "hello")
	if err != nil || len(v) != 3 { t.Fatalf("embedding=%v err=%v", v, err) }
}

func TestLearningLifecycle(t *testing.T) {
	r := NewRuntime()
	ctx := context.Background()
	if err := r.TrainOnline(ctx, []string{"a", "b"}); err != nil { t.Fatal(err) }
	if err := r.FineTuneLoRA(ctx, "adapter", "example"); err != nil { t.Fatal(err) }
	if err := r.SwapLoRA(ctx, "adapter"); err != nil { t.Fatal(err) }
	if err := r.ReplayExperience(ctx, []string{"c"}); err != nil { t.Fatal(err) }
}

func TestLearningInputValidation(t *testing.T) {
	r := NewRuntime()
	ctx := context.Background()
	cases := []struct{name string; call func() error}{
		{"TrainOnline", func() error { return r.TrainOnline(ctx, nil) }},
		{"FineTuneLoRA", func() error { return r.FineTuneLoRA(ctx, "", "") }},
		{"SwapLoRA", func() error { return r.SwapLoRA(ctx, "missing") }},
		{"ReplayExperience", func() error { return r.ReplayExperience(ctx, nil) }},
	}
	for _, tc := range cases { t.Run(tc.name, func(t *testing.T) { if err := tc.call(); err == nil { t.Fatal("expected validation error") } }) }
}
