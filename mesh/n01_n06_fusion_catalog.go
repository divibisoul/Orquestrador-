package mesh

import "github.com/divibisoul/Orquestrador-/protocol"

type FusionCapability struct {
    ID string `json:"id"`
    Owner string `json:"owner"`
    Kind string `json:"kind"`
    Consumers []string `json:"consumers"`
    Composable bool `json:"composable"`
}

var n01FusionCapabilities = []string{
    "android.device_info", "android.battery", "android.memory", "android.network",
    "android.events", "shizuku.bridge", "brightness", "wifi", "bluetooth",
    "airplane", "background-process",
}

var n06FusionCapabilities = []string{
    "support.context", "support.artifacts", "support.documents", "support.tool-execution",
    "support.streaming", "support.mesh", "support.ai-pilot",
    "tool:createDocument", "tool:updateDocument", "tool:getWeather", "tool:requestSuggestions",
}

func VerifiedN01N06Catalog() []FusionCapability {
    out := make([]FusionCapability, 0, len(n01FusionCapabilities)+len(n06FusionCapabilities))
    for _, id := range n01FusionCapabilities { out = append(out, FusionCapability{ID:id,Owner:protocol.N01,Kind:"native",Consumers:[]string{protocol.N07},Composable:true}) }
    for _, id := range n06FusionCapabilities { out = append(out, FusionCapability{ID:id,Owner:protocol.N06,Kind:"capability",Consumers:[]string{protocol.N07},Composable:true}) }
    return out
}

func KnownFusionOwner(id string) string {
    for _, c := range VerifiedN01N06Catalog() { if c.ID == id { return c.Owner } }
    return ""
}
