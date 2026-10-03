package supergpu

import "testing"

func TestSelectAcceleratorDoesNotFallBackToCPU(t *testing.T) {
	r := New(nil)
	r.Discover()
	if _, err := r.SelectAccelerator(""); err == nil {
		t.Fatal("CPU fallback must not satisfy accelerator selection")
	}
	if r.AcceleratorAvailable() {
		t.Fatal("CPU-only runtime must report no accelerator")
	}
}

func TestSelectAcceleratorRejectsCPUPreference(t *testing.T) {
	r := New(nil)
	r.Discover()
	if _, err := r.SelectAccelerator("cpu-0"); err == nil {
		t.Fatal("CPU preference must be rejected for accelerator-only selection")
	}
}
