package cognitive

import (
	"context"
	"errors"
	"fmt"
)

func (l *Loop) RunPlanned(ctx context.Context, g Goal) ([]Observation, error) {
	if !l.Enabled() {
		return nil, errors.New("COGNITIVE_LOOP_DISABLED")
	}
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	normalizeGoal(&g, l.cfg.GoalTTL)
	if len(g.Capabilities) == 0 {
		return nil, errors.New("goal requires capabilities")
	}
	if err := l.goals.Put(g); err != nil {
		return nil, err
	}
	if err := l.critic.Check(ctx, g); err != nil {
		return nil, err
	}
	planner, err := NewPlannerFromExecutor(l.executor)
	if err != nil {
		return nil, err
	}
	steps, err := planner.Plan(ctx, g)
	if err != nil {
		return nil, err
	}
	out := make([]Observation, 0, len(steps))
	for _, step := range steps {
		obs := l.executeStep(ctx, step)
		out = append(out, obs)
		if err := l.recordStep(ctx, step, obs); err != nil && obs.OK {
			obs.OK = false
			obs.Error = "COGNITIVE_OBSERVATION_PERSIST_FAILED: " + err.Error()
			out[len(out)-1] = obs
		}
		if !obs.OK {
			return out, fmt.Errorf("goal step %s failed: %s", step.ID, obs.Error)
		}
	}
	return out, nil
}

func NewPlannerFromExecutor(e Executor) (*Planner, error) {
	x, ok := e.(*MeshExecutor)
	if !ok || x == nil {
		return nil, errors.New("planner requires the canonical MeshExecutor")
	}
	return NewPlanner(x.peers)
}
