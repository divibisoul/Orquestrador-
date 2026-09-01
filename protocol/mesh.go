package protocol

import("encoding/json";"errors";"strings")

const MeshContractVersion = "soul-mesh/1"
type MeshEnvelope struct{ContractVersion string `json:"contractVersion"`;Operation string `json:"operation"`;Payload map[string]any `json:"payload,omitempty"`;CorrelationID string `json:"correlationId"`;Source string `json:"source"`;Target string `json:"target"`;Metadata map[string]string `json:"metadata,omitempty"`}
func(m MeshEnvelope)Validate()error{if m.ContractVersion!=MeshContractVersion{return errors.New("unsupported mesh contract version")};if m.Operation==""{return errors.New("operation is required")};if strings.TrimSpace(m.CorrelationID)==""{return errors.New("correlationId is required")};if strings.TrimSpace(m.Source)==""||strings.TrimSpace(m.Target)==""{return errors.New("source and target are required")};return nil}
func EncodeMesh(m MeshEnvelope)([]byte,error){if err:=m.Validate();err!=nil{return nil,err};return json.Marshal(m)}
func DecodeMesh(b []byte)(MeshEnvelope,error){var m MeshEnvelope;if err:=json.Unmarshal(b,&m);err!=nil{return m,err};return m,m.Validate()}

// MessageFromMesh adapts the canonical SOUL mesh envelope into N07's protocol.
// Generic payloads are preserved in PayloadJSON while a numeric "values" field is
// promoted to the legacy Payload channel for existing N07 operations.
func MessageFromMesh(m MeshEnvelope) (Message,error) {
	if err:=m.Validate();err!=nil{return Message{},err}
	msg:=Message{Version:Version,TraceID:m.CorrelationID,Source:m.Source,Target:m.Target,Kind:"mesh",Operation:m.Operation,Metadata:cloneMetadata(m.Metadata)}
	if m.Payload!=nil {
		b,err:=json.Marshal(m.Payload);if err!=nil{return Message{},err};msg.PayloadJSON=b
		if raw,ok:=m.Payload["values"];ok {
			if err:=json.Unmarshal(mustJSON(raw),&msg.Payload);err!=nil{return Message{},errors.New("mesh values must be numeric array")}
		}
	}
	return msg,msg.Validate()
}

// MeshFromMessage converts N07 messages back into the mesh-neutral envelope.
// PayloadJSON is preferred because it preserves arbitrary tool/function payloads.
func MeshFromMessage(m Message) (MeshEnvelope,error) {
	if err:=m.Validate();err!=nil{return MeshEnvelope{},err}
	payload:=map[string]any{}
	if len(m.PayloadJSON)>0 {if err:=json.Unmarshal(m.PayloadJSON,&payload);err!=nil{return MeshEnvelope{},err};if payload==nil{payload=map[string]any{}}}
	if len(m.Payload)>0 {payload["values"]=append([]float64(nil),m.Payload...)}
	return MeshEnvelope{ContractVersion:MeshContractVersion,Operation:m.Operation,Payload:payload,CorrelationID:m.TraceID,Source:m.Source,Target:m.Target,Metadata:cloneMetadata(m.Metadata)},nil
}

func cloneMetadata(in map[string]string)map[string]string{if len(in)==0{return nil};out:=make(map[string]string,len(in));for k,v:=range in{out[k]=v};return out}
func mustJSON(v any)[]byte{b,_:=json.Marshal(v);return b}
