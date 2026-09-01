package mesh

import (
    "testing"
    "github.com/divibisoul/Orquestrador-/protocol"
)

func TestVerifiedN01N06CatalogHasCanonicalOwnership(t *testing.T) {
    catalog := VerifiedN01N06Catalog()
    if len(catalog) != len(n01FusionCapabilities)+len(n06FusionCapabilities) { t.Fatalf("unexpected catalog size: %d", len(catalog)) }
    for _, capability := range catalog {
        if capability.Owner != protocol.N01 && capability.Owner != protocol.N06 { t.Fatalf("invalid owner %q for %s", capability.Owner, capability.ID) }
        if len(capability.Consumers) != 1 || capability.Consumers[0] != protocol.N07 { t.Fatalf("N07 consumer missing for %s", capability.ID) }
        if !capability.Composable { t.Fatalf("capability %s must remain composable", capability.ID) }
    }
    if KnownFusionOwner("android.device_info") != protocol.N01 { t.Fatal("android.device_info ownership drifted") }
    if KnownFusionOwner("tool:createDocument") != protocol.N06 { t.Fatal("createDocument ownership drifted") }
}
