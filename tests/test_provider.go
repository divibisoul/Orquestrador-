package tests

import "context"

type testProvider struct{}
func (testProvider) Generate(_ context.Context, prompt string) (string, error) { return "test:" + prompt, nil }
func (testProvider) Embed(_ context.Context, text string) ([]float64, error) { return []float64{float64(len(text)), 1, 0, 0}, nil }
