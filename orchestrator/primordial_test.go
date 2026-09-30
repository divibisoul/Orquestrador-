package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

func TestPrimordialCompositionOperationUsesLedger(t *testing.T) {
	n, _ := neural.New(8, .05)
	c, _ := prefrontal.New(.1, 16)
	e, err := New(n, c, supergpu.New(nil))
	if err != nil { t.Fatal(err) }

	dir := t.TempDir()
	ledger := filepath.Join(dir, "soul-nuclei.json")
	if err := os.WriteFile(ledger, []byte(`{"nuclei":{"N01":{"repository":"N01","role":"core","entrypoints":[],"capabilityOwnership":"N01"},"N02":{"repository":"N02","role":"neural","entrypoints":[],"capabilityOwnership":"N02"}},"transversal":{},"fusion":{"topology":"test","policy":"preserve"},"primordialEssence":{"definition":"test","preservationRule":"preserve","essences":{"N01":{"nativeRole":"core","essence":"a","evidence":["test"]},"N02":{"nativeRole":"neural","essence":"b","evidence":["test"]}},"compositionSeeds":[{"participants":["N01","N02"],"mode":"adjacent","derivedFunction":"core+neural","existingEvidence":["test"],"status":"STRUCTURAL"}]}}`), 0o600); err != nil { t.Fatal(err) }
	}
	if err := RegisterPrimordialCompositionOperation(e); err != nil { t.Fatal(err) }
	m := protocol.NewMessage("N01", "N07", "command", "composition.primordial.resolve@1.0.0", nil)
	m.Metadata = map[string]string{"ledger_path": ledger, "participant_a": "N01", "participant_b": "N02"}
	result, err := e.Submit(context.Background(), m)
	if err != nil { t.Fatal(err) }
	if result.Metadata["mode"] != "pair" || result.Metadata["primordial_json"] == "" { t.Fatalf("unexpected result: %#v", result.Metadata) }
}
