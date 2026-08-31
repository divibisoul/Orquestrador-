package cognitive

import (
	"context"

	"github.com/divibisoul/Orquestrador-/compute"
	"github.com/divibisoul/Orquestrador-/core/neuralfabric"
)

type Cortex struct {
	Fabric   compute.Fabric
	Encoder  neuralfabric.Encoder
	Predictor neuralfabric.Predictor
	Router   neuralfabric.Router
	Learner  neuralfabric.Learner
}

func (c *Cortex) DecideAndExecute(ctx context.Context, goal string, devices []compute.Device) (compute.Result, error) {
	if len(devices) == 0 {
		return compute.Result{}, context.Canceled
	}

	s := neuralfabric.State{Goal: goal}
	candidates := make([]neuralfabric.Route, 0, len(devices))
	for _, d := range devices {
		candidates = append(candidates, neuralfabric.Route{NodeID: d.ID, DeviceID: d.ID, Precision: "fp16", BatchSize: 1})
	}

	r, err := c.Router.Route(ctx, s, candidates)
	if err != nil {
		return compute.Result{}, err
	}

	var selected compute.Device
	for _, d := range devices {
		if d.ID == r.DeviceID {
			selected = d
			break
		}
	}
	if selected.ID == "" {
		return compute.Result{}, context.Canceled
	}

	job := compute.Job{ID: goal, Model: "external-provider", Precision: compute.Precision(r.Precision), QualityTarget: .8}
	return c.Fabric.Execute(ctx, job, selected)
}
