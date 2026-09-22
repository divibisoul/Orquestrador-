package supergpu

import (
	"context"
	"testing"
)

func TestFederationUsesRealRuntimeForNuclei(t *testing.T) {
	r := New(nil)
	devices := r.Discover()
	if len(devices) == 0 {
		t.Fatal("expected discovered host compute device")
	}
	f, err := NewFederation(r)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.Execute(context.Background(), FederatedRequest{Nucleus: "N01", Operation: "square", Payload: []float64{2, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Output) != 2 || got.Output[0] != 4 || got.Output[1] != 9 {
		t.Fatalf("unexpected result: %#v", got.Output)
	}
	if got.Device.ID == "" {
		t.Fatal("missing leased device")
	}
	if r.Health()["reservations"].(int) != 0 {
		t.Fatal("federated lease was not released")
	}
}
func TestFederationRejectsUnknownNucleus(t *testing.T) {
	r := New(nil)
	r.Discover()
	f, _ := NewFederation(r)
	if _, err := f.Execute(context.Background(), FederatedRequest{Nucleus: "N99", Operation: "identity", Payload: []float64{1}}); err == nil {
		t.Fatal("expected invalid nucleus rejection")
	}
}
