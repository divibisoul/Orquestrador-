package mesh

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type HTTPGateway struct {
	Engine                    *orchestrator.Engine
	Secret                    string
	AllowUnauthenticatedLocal bool
}
type canonicalWireEnvelope struct {
	Protocol        string            `json:"protocol"`
	ContractVersion string            `json:"contractVersion"`
	ID              string            `json:"id"`
	MessageID       string            `json:"messageId"`
	CorrelationID   string            `json:"correlationId"`
	Source          string            `json:"source"`
	Target          string            `json:"target"`
	Kind            string            `json:"kind"`
	Type            string            `json:"type"`
	Capability      string            `json:"capability"`
	Payload         map[string]any    `json:"payload"`
	Timestamp       int64             `json:"timestamp"`
