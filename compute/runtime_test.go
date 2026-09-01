package compute

import "testing"

func TestAddDeviceIsIdempotentByID(t *testing.T) {
	fabric := NewLocalFabric(nil)
	fabric.AddDevice(Device{ID: "cpu-1", Kind: CPU, Ready: false})
	fabric.AddDevice(Device{ID: "cpu-1", Kind: CPU, Ready: true, FLOPs: 123})

	devices, err := fabric.Devices(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 {
		t.Fatalf("device count=%d want 1", len(devices))
	}
	if !devices[0].Ready || devices[0].FLOPs != 123 {
		t.Fatalf("device was not replaced: %+v", devices[0])
	}
}
