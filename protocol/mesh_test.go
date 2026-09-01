package protocol
import "testing"
func TestMeshRoundTrip(t *testing.T){m:=MeshEnvelope{ContractVersion:"1.2",Operation:"execute",Payload:map[string]any{"x":1},CorrelationID:"corr-1",Source:"N07",Target:"N01"};b,err:=EncodeMesh(m);if err!=nil{t.Fatal(err)};d,err:=DecodeMesh(b);if err!=nil{t.Fatal(err)};if d.CorrelationID!=m.CorrelationID||d.Operation!=m.Operation{t.Fatal("mesh round-trip failed")}}
