package orchestrator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

func TestGRCERegistrationRemainsFailClosedWithoutExecutor(t *testing.T) {
	err := orchestrator.RegisterGRFGRCEOperations(nil)
	if err == nil || !strings.Contains(err.Error(), "orchestrator engine is required") {
		t.Fatalf("unexpected registration result: %v", err)
	}

	_ = context.Background()
}
