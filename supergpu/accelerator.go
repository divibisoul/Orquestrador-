package supergpu

import (
	"errors"
	"strings"
)

// SelectAccelerator returns only a real available accelerator device.
// CPU is never returned by this method.
func (r *Runtime) SelectAccelerator(preferred string) (Device, error) {
	if r == nil {
		return Device{}, errors.New("supergpu runtime unavailable")
	}
	devices := r.Discover()
	preferred = strings.TrimSpace(preferred)
	if preferred != "" {
		for _, d := range devices {
			if d.ID != preferred {
				continue
			}
			if d.Backend == "cpu" {
				return Device{}, errors.New("SUPERGPU_ACCELERATOR_REQUIRED")
			}
			if !d.Available {
				return Device{}, errors.New("SUPERGPU_ACCELERATOR_UNAVAILABLE")
			}
			return d, nil
		}
		return Device{}, errors.New("SUPERGPU_ACCELERATOR_NOT_FOUND")
	}
	for _, d := range devices {
		if d.Backend != "cpu" && d.Available {
			return d, nil
		}
	}
	return Device{}, errors.New("SUPERGPU_ACCELERATOR_UNAVAILABLE")
}

func (r *Runtime) AcceleratorAvailable() bool {
	if r == nil {
		return false
	}
	_, err := r.SelectAccelerator("")
	return err == nil
}
