package protocol

import("encoding/json";"errors";"strings")

type MeshEnvelope struct{ContractVersion string `json:"contractVersion"`;Operation string `json:"operation"`;Payload map[string]any `json:"payload,omitempty"`;CorrelationID string `json:"correlationId"`;Source string `json:"source"`;Target string `json:"target"`;Metadata map[string]string `json:"metadata,omitempty"`}
func(m MeshEnvelope)Validate()error{if m.ContractVersion==""{return errors.New("contractVersion is required")};if m.Operation==""{return errors.New("operation is required")};if strings.TrimSpace(m.CorrelationID)==""{return errors.New("correlationId is required")};if strings.TrimSpace(m.Source)==""||strings.TrimSpace(m.Target)==""{return errors.New("source and target are required")};return nil}
func EncodeMesh(m MeshEnvelope)([]byte,error){if err:=m.Validate();err!=nil{return nil,err};return json.Marshal(m)}
func DecodeMesh(b []byte)(MeshEnvelope,error){var m MeshEnvelope;if err:=json.Unmarshal(b,&m);err!=nil{return m,err};return m,m.Validate()}
