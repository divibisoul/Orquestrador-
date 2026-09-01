package mesh

import (
    "testing"
    "github.com/divibisoul/Orquestrador-/protocol"
)

func TestVerifiedN01N06Catalog(t *testing.T) {
    catalog := VerifiedN01N06Catalog()
    if len(catalog) != len(n01FusionCapabilities)+len(n06FusionCapabilities) { t.Fatalf("catalog size=%d", len(catalog)) }
    seen := map[string]bool{}
    for _, c := range catalog {
        if c.ID == "" || seen[c.ID] { t.Fatalf("invalid or duplicate capability: %q", c.ID) }
        seen[c.ID] = true
        if c.Owner != protocol.N01 && c.Owner != protocol.N06 { t.Fatalf("invalid owner %q", c.Owner) }
        if len(c.Consumers) != 1 || c.Consumers[0] != protocol.N07 { t.Fatalf("N07 consumer missing for %s", c.ID) }
        if !c.Composable { t.Fatalf("capability not composable: %s", c.ID) }
    }
    if KnownFusionOwner("android.device_info") != protocol.N01 { t.Fatal("N01 ownership mismatch") }
    if KnownFusionOwner("tool:createDocument") != protocol.N06 { t.Fatal("N06 tool ownership mismatch") }
}
