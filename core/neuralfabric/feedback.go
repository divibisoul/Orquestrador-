package neuralfabric

import (
	"context"
	"errors"
	"github.com/divibisoul/Orquestrador-/core/trinity"
)

func (f *Fabric) Feedback(ctx context.Context, fb trinity.Feedback) error {
	if f == nil { return errors.New("nil neural fabric") }
	if ctx == nil { return errors.New("nil context") }
	if err := ctx.Err(); err != nil { return err }
	if fb.WorkloadID == "" || fb.Route.Model == "" { return errors.New("invalid feedback") }
	reward := fb.Quality
	if reward == 0 { if fb.Success { reward = 1 } else { reward = -1 } }
	f.tree.Update(fb.Route.Model, reward, f.cfg.LearningRate)
	return nil
}
