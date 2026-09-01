package mesh

import (
	"context"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/supergpu"
	"testing"
)

func TestEnhancedFederatedGatewayConstruction(t *testing.T) {
	n, _ := neural.New(2, .1)
	c, _ := prefrontal.New(.1, 4)
	g := supergpu.New(nil)
	g.Discover()
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	gw := NewEnhancedFederatedHTTPGateway(e)
	if gw == nil || gw.base == nil {
		t.Fatal("enhanced gateway not initialized")
	}
	if len(e.Operations()) < 6 {
		t.Fatal("built-in operations missing")
	}
	if err := e.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
