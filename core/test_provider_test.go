package core_test

import "context"

type integrationProvider struct{}
func (integrationProvider) Generate(_ context.Context, prompt string) (string, error) { return "test:" + prompt, nil }
func (integrationProvider) Embed(_ context.Context, text string) ([]float64, error) { return []float64{float64(len(text)), 1, 0, 0}, nil }
