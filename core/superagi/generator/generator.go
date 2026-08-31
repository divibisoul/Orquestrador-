package generator

import (
	"context"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

// Generator is the canonical generation facade. Providers stay owned by Runtime.
type Generator struct { Runtime *superagi.Runtime }

func New(r *superagi.Runtime) *Generator {
	if r == nil { r = superagi.NewRuntime() }
	return &Generator{Runtime:r}
}

func (g *Generator) Text(ctx context.Context, prompt string) (string, error) {
	if g == nil || g.Runtime == nil { return "", errors.New("generator runtime unavailable") }
	return g.Runtime.GenerateText(ctx, strings.TrimSpace(prompt))
}
func (g *Generator) Code(ctx context.Context, spec, language string) (string, error) {
	if g == nil || g.Runtime == nil { return "", errors.New("generator runtime unavailable") }
	return g.Runtime.GenerateCode(ctx, strings.TrimSpace(spec), strings.TrimSpace(language))
}
func (g *Generator) Summary(ctx context.Context, text string) (string, error) {
	if g == nil || g.Runtime == nil { return "", errors.New("generator runtime unavailable") }
	return g.Runtime.Summarize(ctx, text)
}
func (g *Generator) Translate(ctx context.Context, text, from, to string) (string, error) {
	if g == nil || g.Runtime == nil { return "", errors.New("generator runtime unavailable") }
	return g.Runtime.Translate(ctx, text, from, to)
}
func (g *Generator) Image(ctx context.Context, prompt string) ([]byte, error) {
	if g == nil || g.Runtime == nil { return nil, errors.New("generator runtime unavailable") }
	return g.Runtime.GenerateImage(ctx, strings.TrimSpace(prompt))
}
