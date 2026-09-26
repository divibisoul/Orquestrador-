package shared

// ClareiraPacket is the application-level packet carried by the existing soul-mesh/1 fabric.
type ClareiraPacket struct {
	ID                 string                 `json:"id"`
	Data               string                 `json:"data"`
	InformationalValue float64                `json:"informationalValue"`
	Criticality        float64                `json:"criticality"`
	PacketType         string                 `json:"packetType"`
	SourceID           string                 `json:"sourceId"`
	DestinationHint    string                 `json:"destinationHint,omitempty"`
	Timestamp          int64                  `json:"timestamp"`
	CorrelationID      string                 `json:"correlationId"`
	Metadata           map[string]interface{} `json:"metadata"`
}