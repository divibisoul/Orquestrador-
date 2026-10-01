package mesh

import (
	"context"
	"testing"

	"github.com/divibisoul/Orquestrador-/protocol"
)

type enhancedGatewayTestEngine struct {
	operations []string
}

func (e *enhancedGatewayTestEngine) Operations() []string {
	return append([]string(nil), e.operations...)
}

func (e *enhancedGatewayTestEngine) Submit(context.Context, protocol.Message) (protocol.Result, error) {
	return protocol.Result{Status: "ok"}, nil
}

func TestEnhancedFederatedGatewayConstruction(t *testing.T) {
	e := &enhancedGatewayTestEngine{
		operations: []string{
			"mesh.ping@1.0.0",
			"mesh.describe@1.0.0",
			"mesh.supergpu.parallel@1.0.0",
			"core.health@1.0.0",
		},
	}
	gw := NewEnhancedFederatedHTTPGateway(e)
	if gw == nil || gw.base == nil {
		t.Fatal("enhanced gateway not initialized")
	}
	if len(e.Operations()) < 4 {
		t.Fatal("built-in operations missing")
	}
}
