package learning

import (
	"context"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

// Engine exposes the canonical runtime learning lifecycle without duplicating state.
type Engine struct { Runtime *superagi.Runtime }

func New(r *superagi.Runtime) *Engine {
	if r == nil { r = superagi.NewRuntime() }
	return &Engine{Runtime:r}
}

func (e *Engine) TrainOnline(ctx context.Context, batch []string) error {
	if e == nil || e.Runtime == nil { return errors.New("learning runtime unavailable") }
	return e.Runtime.TrainOnline(ctx, batch)
}
func (e *Engine) Replay(ctx context.Context, batch []string) error {
	if e == nil || e.Runtime == nil { return errors.New("learning runtime unavailable") }
	return e.Runtime.ReplayExperience(ctx, batch)
}
func (e *Engine) FineTuneLoRA(ctx context.Context, name, data string) error {
	if e == nil || e.Runtime == nil { return errors.New("learning runtime unavailable") }
	return e.Runtime.FineTuneLoRA(ctx, strings.TrimSpace(name), data)
}
func (e *Engine) ActivateLoRA(ctx context.Context, name string) error {
	if e == nil || e.Runtime == nil { return errors.New("learning runtime unavailable") }
	return e.Runtime.SwapLoRA(ctx, strings.TrimSpace(name))
}
func (e *Engine) Snapshot(n int) []string {
	if e == nil || e.Runtime == nil { return nil }
	return e.Runtime.ExperienceSnapshot(n)
}
