package backend

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strings"

  "github.com/divibisoul/Orquestrador-/orchestrator"
  "github.com/divibisoul/Orquestrador-/protocol"
)

// RegisterVagusOperation exposes the canonical NervoVago boundary through N07.
// SARA remains the implementation authority; N07 is the authenticated federation bridge.
func RegisterVagusOperation(e *orchestrator.Engine, proxy *SARAProxy) error {
  if e == nil || proxy == nil {
    return errors.New("NervoVago operation requires engine and SARA proxy")
  }
  handler := func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
    raw := strings.TrimSpace(message.Metadata["sara_vagus_json"])
    if raw == "" {
      return protocol.Result{}, errors.New("metadata.sara_vagus_json is required")
    }
    var event map[string]any
    if err := json.Unmarshal([]byte(raw), &event); err != nil {
      return protocol.Result{}, fmt.Errorf("invalid NervoVago event: %w", err)
    }
    out, err := proxy.PublishVagus(ctx, event, message.CorrelationID)
    return saraResult(message, out, err)
  }
  return e.Register("nervo.vago.publish@1.0.0", handler)
}
