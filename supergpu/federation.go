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
	Nucleus   string
	Operation string
	Payload   []float64
	Device    string
}

type FederatedResult struct {
	Nucleus string
	Device  Device
	Output  []float64
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

// Execute grants any SOUL nucleus a real execution lease on N07 compute hardware.
// No synthetic GPU is created: unavailable hardware is returned as an explicit error.
func (f *Federation) Execute(ctx context.Context, req FederatedRequest) (FederatedResult, error) {
	if f == nil || f.Runtime == nil {
		return FederatedResult{}, errors.New("supergpu federation is unavailable")
	}
	req.Nucleus, req.Operation, req.Device = strings.TrimSpace(req.Nucleus), strings.TrimSpace(req.Operation), strings.TrimSpace(req.Device)
	if !validClient(req.Nucleus) {
		return FederatedResult{}, fmt.Errorf("SUPERGPU_NUCLEUS_INVALID: %s", req.Nucleus)
	}
	if req.Operation == "" {
		return FederatedResult{}, errors.New("supergpu operation is required")
	}
	if len(req.Payload) == 0 {
		return FederatedResult{}, errors.New("supergpu payload is empty")
	}
	device, err := f.Runtime.Select(req.Device)
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

func (f *Federation) Health() map[string]any {
	if f == nil || f.Runtime == nil {
		return map[string]any{"status": "degraded", "error": "runtime unavailable"}
	}
	h := f.Runtime.Health()
	h["federation"] = "N01..N07"
	h["resource_policy"] = "N07-control-plane-with-nucleus-scoped-leases"
	return h
}
