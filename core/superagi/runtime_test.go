package superagi

import (
	"context"
	"testing"
)

func TestLearningBoundariesAreExplicit(t *testing.T) {
	r := NewRuntime()
	ctx := context.Background()

	cases := []struct {
		name string
		call func() error
	}{
		{"TrainOnline", func() error { return r.TrainOnline(ctx, nil) }},
		{"FineTuneLoRA", func() error { return r.FineTuneLoRA(ctx, "", "") }},
		{"SwapLoRA", func() error { return r.SwapLoRA(ctx, "") }},
		{"ReplayExperience", func() error { return r.ReplayExperience(ctx, nil) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err == nil {
				t.Fatalf("%s returned success without an implemented backend", tc.name)
			}
		})
	}
}
