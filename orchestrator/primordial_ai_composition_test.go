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

func TestPrimordialCompositionOperationExposesAIProfilesAndHigherOrderComposition(t *testing.T) {
	n, _ := neural.New(8, .05)
	c, _ := prefrontal.New(.1, 16)
	e, err := New(n, c, supergpu.New(nil))
	if err != nil { t.Fatal(err) }

	dir := t.TempDir()
	ledger := filepath.Join(dir, "soul-nuclei.json")
	contents := `{"nuclei":{
		"N01":{"repository":"N01","role":"core","entrypoints":["n01"],"capabilityOwnership":"native"},
		"N02":{"repository":"N02","role":"language","entrypoints":["n02"],"capabilityOwnership":"native"},
		"N03":{"repository":"N03","role":"perception","entrypoints":["n03"],"capabilityOwnership":"native"},
		"N04":{"repository":"N04","role":"tools","entrypoints":["n04"],"capabilityOwnership":"native"}
	},"transversal":{
		"SARA":{"repository":"SARA","role":"regeneration","entrypoints":["sara"],"capabilityOwnership":"transversal"}
	},"fusion":{"topology":"test","policy":"preserve"},"primordialEssence":{
		"definition":"test","preservationRule":"preserve",
		"essences":{
			"N01":{"nativeRole":"core","essence":"coordination","evidence":["n01"]},
			"N02":{"nativeRole":"language","essence":"language","evidence":["n02"]},
			"N03":{"nativeRole":"perception","essence":"perception","evidence":["n03"]},
			"N04":{"nativeRole":"tools","essence":"tools","evidence":["n04"]},
			"SARA":{"nativeRole":"regeneration","essence":"regeneration","evidence":["sara"]}
		},
		"compositionSeeds":[
			{"participants":["N02","N03","N04"],"mode":"higher-order","derivedFunction":"perceive -> reason -> act","existingEvidence":["triple"],"status":"STRUCTURAL"},
			{"participants":["N01","N05"],"mode":"orchestrated","derivedFunction":"context -> inference","existingEvidence":["pair"],"status":"STRUCTURAL"},
			{"participants":["N04","SARA"],"mode":"transversal","derivedFunction":"tool execution under governance","existingEvidence":["transversal"],"status":"STRUCTURAL"}
		]
	}}}`
	// Add N05 to the fixture without changing the runtime code under test.
	contents = stringReplace(contents, `"N04":{"repository":"N04","role":"tools","entrypoints":["n04"],"capabilityOwnership":"native"}`, `"N04":{"repository":"N04","role":"tools","entrypoints":["n04"],"capabilityOwnership":"native"},"N05":{"repository":"N05","role":"inference","entrypoints":["n05"],"capabilityOwnership":"native"}`)
	contents = stringReplace(contents, `"N04":{"nativeRole":"tools","essence":"tools","evidence":["n04"]}`, `"N04":{"nativeRole":"tools","essence":"tools","evidence":["n04"]},"N05":{"nativeRole":"inference","essence":"inference","evidence":["n05"]}`)
	if err := os.WriteFile(ledger, []byte(contents), 0o600); err != nil { t.Fatal(err) }

	if err := RegisterPrimordialCompositionOperation(e); err != nil { t.Fatal(err) }

	compositionMessage := protocol.NewMessage("N03", "N07", "command", "composition.primordial.resolve@1.0.0", nil)
	compositionMessage.Metadata = map[string]string{"ledger_path": ledger, "mode": "composition", "participants": "N03,N02,N04"}
	result, err := e.Submit(context.Background(), compositionMessage)
	if err != nil { t.Fatal(err) }
	if result.Metadata["mode"] != "composition" || result.Metadata["primordial_json"] == "" {
		t.Fatalf("unexpected composition result: %#v", result.Metadata)
	}

	profileMessage := protocol.NewMessage("N01", "N07", "command", "composition.primordial.resolve@1.0.0", nil)
	profileMessage.Metadata = map[string]string{"ledger_path": ledger, "mode": "profile", "participant": "SARA"}
	profileResult, err := e.Submit(context.Background(), profileMessage)
	if err != nil { t.Fatal(err) }
	if profileResult.Metadata["mode"] != "profile" || profileResult.Metadata["primordial_json"] == "" {
		t.Fatalf("unexpected profile result: %#v", profileResult.Metadata)
	}
}

func stringReplace(value, old, replacement string) string {
	if old == replacement { return value }
	for {
		idx := indexOf(value, old)
		if idx < 0 { return value }
		value = value[:idx] + replacement + value[idx+len(old):]
	}
}

func indexOf(value, needle string) int {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle { return i }
	}
	return -1
}
