package protocol

import "testing"

func TestFederatedMessageCarriesDataAndCorrelation(t *testing.T) {
    m := NewMessage("N01", "N07", "command", "fusion.invoke", nil)
    m.Data = map[string]any{"capability":"android.device_info", "payload":map[string]any{"request":"device"}}
    encoded, err := Encode(m)
    if err != nil { t.Fatal(err) }
    decoded, err := Decode(encoded)
    if err != nil { t.Fatal(err) }
    if decoded.CorrelationID != m.CorrelationID { t.Fatal("correlation id not preserved") }
    if decoded.Data["capability"] != "android.device_info" { t.Fatal("federated capability data not preserved") }
}
