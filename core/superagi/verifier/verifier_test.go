package verifier

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

func TestVerifierAdapter(t *testing.T) {
	r := superagi.NewRuntime()
	v := New(r)
	ok, confidence, err := v.Safety(context.Background(), "hello")
	if err != nil || !ok || confidence <= 0 { t.Fatalf("ok=%v confidence=%v err=%v", ok, confidence, err) }
}
