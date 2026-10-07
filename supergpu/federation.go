package supergpu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/divibisoul/Orquestrador-/protocol"
)

// FederatedRequest is a real compute lease from N07 on behalf of N01-N06.
type FederatedRequest struct {
	Nucleus            string
	Operation          string
	Payload            []float64
	Device             string
	RequireAccelerator bool
}

type FederatedBatchRequest struct {
	Nucleus            string
	Operation          string
	Inputs             [][]float64
	Device             string
	Workers            int
	RequireAccelerator bool
}

type FederatedResult struct {
	Nucleus string
	Device  Device
	Output  []float64
}

type FederatedBatchResult struct {
	Nucleus string
	Device  Device
	Outputs [][]float64
}

type Federation struct{ Runtime *Runtime }

func NewFederation(r *Runtime) (*Federation, error) {
	if r == nil {
		return nil, errors.New("supergpu runtime is required")
	}
	return &Federation{Runtime: r}, nil
}

func validClient(nucleus string) bool {
	switch nucleus {
	case protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06, protocol.N07:
		return true
	default:
		return false
	}
}

func (f *Federation) device(requireAccelerator bool, preferred string) (Device, error) {
	if requireAccelerator {
		return f.Runtime.SelectAccelerator(preferred)
	}
	return f.Runtime.Select(preferred)
}

// Execute grants any SOUL nucleus a real execution lease on N07 compute hardware.
// No synthetic GPU is created: unavailable hardware is returned as an explicit error.
func (f *Federation) Execute(ctx context.Context, req FederatedRequest) (FederatedResult, error) {
	if f == nil || f.Runtime == nil {
		return FederatedResult{}, errors.New("supergpu federation is unavailable")
	}
	req.Nucleus = strings.TrimSpace(req.Nucleus)
	req.Operation = strings.TrimSpace(req.Operation)
	req.Device = strings.TrimSpace(req.Device)
	if !validClient(req.Nucleus) {
		return FederatedResult{}, fmt.Errorf("SUPERGPU_NUCLEUS_INVALID: %s", req.Nucleus)
	}
	if req.Operation == "" {
		return FederatedResult{}, errors.New("supergpu operation is required")
	}
	if len(req.Payload) == 0 {
		return FederatedResult{}, errors.New("supergpu payload is empty")
	}
	device, err := f.device(req.RequireAccelerator, req.Device)
	if err != nil {
		return FederatedResult{}, err
	}
	if err := f.Runtime.Reserve(device.ID, req.Nucleus); err != nil {
		return FederatedResult{}, err
	}
	defer f.Runtime.Release(device.ID, req.Nucleus)
	output, err := f.Runtime.Execute(ctx, device, req.Operation, req.Payload)
	if err != nil {
		return FederatedResult{}, err
	}
	return FederatedResult{Nucleus: req.Nucleus, Device: device, Output: output}, nil
}

func (f *Federation) ExecuteBatch(ctx context.Context, req FederatedBatchRequest) (FederatedBatchResult, error) {
	if f == nil || f.Runtime == nil {
		return FederatedBatchResult{}, errors.New("supergpu federation is unavailable")
	}
	req.Nucleus = strings.TrimSpace(req.Nucleus)
	req.Operation = strings.TrimSpace(req.Operation)
	req.Device = strings.TrimSpace(req.Device)
	if !validClient(req.Nucleus) {
		return FederatedBatchResult{}, fmt.Errorf("SUPERGPU_NUCLEUS_INVALID: %s", req.Nucleus)
	}
	if req.Operation == "" {
		return FederatedBatchResult{}, errors.New("supergpu operation is required")
	}
	if len(req.Inputs) == 0 {
		return FederatedBatchResult{}, errors.New("supergpu inputs are empty")
	}
	if req.Workers <= 0 {
		req.Workers = 1
	}
	device, err := f.device(req.RequireAccelerator, req.Device)
	if err != nil {
		return FederatedBatchResult{}, err
	}
	if err := f.Runtime.Reserve(device.ID, req.Nucleus); err != nil {
		return FederatedBatchResult{}, err
	}
	defer f.Runtime.Release(device.ID, req.Nucleus)
	outputs, err := f.Runtime.BatchParallel(ctx, device, req.Operation, req.Inputs, req.Workers)
	if err != nil {
		return FederatedBatchResult{}, err
	}
	return FederatedBatchResult{Nucleus: req.Nucleus, Device: device, Outputs: outputs}, nil
}

func (f *Federation) Health() map[string]any {
	if f == nil || f.Runtime == nil {
		return map[string]any{"status": "degraded", "error": "runtime unavailable"}
	}
	h := f.Runtime.Health()
	h["federation"] = "N01..N07"
	h["resource_policy"] = "N07-control-plane-with-nucleus-scoped-leases"
	h["accelerator_only_api"] = true
	return h
}
