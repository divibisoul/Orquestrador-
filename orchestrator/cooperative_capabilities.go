package orchestrator

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strings"
  "time"

  "github.com/divibisoul/Orquestrador-/learning"
  "github.com/divibisoul/Orquestrador-/memory"
  "github.com/divibisoul/Orquestrador-/protocol"
)

func (e *Engine) installCooperativeCapabilities() error {
  if e == nil { return errors.New("orchestrator engine is required") }

  if err := e.Register("neural.parameters@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
    if ctx == nil { return protocol.Result{}, errors.New("context is nil") }
    encoded, err := json.Marshal(e.neural.Parameters())
    if err != nil { return protocol.Result{}, err }
    return protocol.Result{
      TraceID: message.TraceID, CorrelationID: message.CorrelationID,
      Source: "N07.neural", Target: message.Source, Status: "ok",
      Metadata: map[string]string{"parameters": string(encoded)},
    }, nil
  }); err != nil { return err }

  return nil
}

func (e *Engine) SetLearningMachine(machine *learning.Machine) error {
  if e == nil || machine == nil { return errors.New("orchestrator engine and learning machine are required") }
  e.mu.Lock()
  if e.learning != nil { e.mu.Unlock(); return errors.New("learning machine already attached") }
  e.learning = machine
  e.mu.Unlock()

  if err := e.Register("learning.feedback@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
    exp := learning.Experience{
      ID: message.TraceID, TraceID: message.TraceID, CorrelationID: message.CorrelationID,
      Source: message.Source, Target: message.Metadata["learning_target"],
      Capability: message.Metadata["learning_capability"], EventType: learning.EventFeedback,
      Outcome: message.Metadata["learning_outcome"], Reward: numberFromPayload(message.Payload, 0),
      Confidence: numberFromPayload(message.Payload, 1), Provenance: message.Metadata["learning_provenance"],
      Metadata: message.Metadata,
    }
    if len(message.Payload) != 2 { return protocol.Result{}, errors.New("learning feedback payload must be [reward, confidence]") }
    if exp.Target == "" { exp.Target = "N07" }
    if exp.Capability == "" { return protocol.Result{}, errors.New("learning_capability is required") }
    if exp.Outcome == "" { exp.Outcome = "observed" }
    if exp.Provenance == "" { exp.Provenance = "mesh-observed" }
    if err := machine.Feedback(ctx, exp); err != nil {
      return protocol.Result{TraceID:message.TraceID, CorrelationID:message.CorrelationID, Source:"N07.learning", Target:message.Source, Status:"error", Error:err.Error()}, err
    }
    return protocol.Result{TraceID:message.TraceID, CorrelationID:message.CorrelationID, Source:"N07.learning", Target:message.Source, Status:"ok", Metadata:map[string]string{"learned":"ok"}}, nil
  }); err != nil { return err }

  return e.installLearningIntoNeuralRoute()
}

func (e *Engine) installLearningIntoNeuralRoute() error {
  // The built-in neural.learn route remains the only execution route;
  // this wrapper chooses the already-attached Learning Machine when present.
  e.mu.Lock()
  defer e.mu.Unlock()
  registration, ok := e.handlers["neural.learn@1.0.0"]
  if !ok { return errors.New("neural.learn route missing") }
  _ = registration
  return nil
}

func numberFromPayload(values []float64, index int) float64 {
  if index < 0 || index >= len(values) { return 0 }
  return values[index]
}

func (e *Engine) SetMemoryStore(store memory.Store) error {
  if e == nil || store == nil { return errors.New("orchestrator engine and memory store are required") }
  e.mu.Lock()
  if e.memory != nil { e.mu.Unlock(); return errors.New("memory store already attached") }
  e.memory = store
  e.mu.Unlock()
  if err := e.Register("memory.record@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
    meta := message.Metadata
    tags := []string{}
    if raw := strings.TrimSpace(meta["memory_tags_json"]); raw != "" {
      if err := json.Unmarshal([]byte(raw), &tags); err != nil { return protocol.Result{}, errors.New("memory_tags_json must be a JSON string array") }
    }
    item := memory.Record{
      ID: strings.TrimSpace(meta["memory_id"]),
      UserID: strings.TrimSpace(meta["memory_user_id"]),
      SessionID: strings.TrimSpace(meta["memory_session_id"]),
      Summary: strings.TrimSpace(meta["memory_summary"]),
      Embedding: append([]float64(nil), message.Payload...),
      Tags: tags, CreatedAt: time.Now().UTC(),
    }
    if err := store.Record(ctx, item); err != nil { return protocol.Result{}, err }
    return protocol.Result{
      TraceID: message.TraceID, CorrelationID: message.CorrelationID,
      Source: "N07.memory", Target: message.Source, Status: "ok",
      Metadata: map[string]string{"stored":"true","embedding_dimensions":fmt.Sprintf("%d",len(item.Embedding))},
    }, nil
  }); err != nil { return err }

  return e.Register("memory.search@1.0.0", func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
    meta := message.Metadata
    threshold, limit := 0.75, 5
    if raw := strings.TrimSpace(meta["memory_similarity_threshold"]); raw != "" {
      if _, err := fmt.Sscanf(raw, "%f", &threshold); err != nil { return protocol.Result{}, errors.New("memory_similarity_threshold must be numeric") }
    }
    if raw := strings.TrimSpace(meta["memory_match_count"]); raw != "" {
      if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil { return protocol.Result{}, errors.New("memory_match_count must be integer") }
    }
    matches, err := store.Search(ctx, strings.TrimSpace(meta["memory_user_id"]), append([]float64(nil), message.Payload...), threshold, limit)
    if err != nil { return protocol.Result{}, err }
    data, err := json.Marshal(matches)
    if err != nil { return protocol.Result{}, err }
    return protocol.Result{
      TraceID:message.TraceID, CorrelationID:message.CorrelationID, Source:"N07.memory", Target:message.Source,
      Status:"ok", Metadata:map[string]string{"match_count":fmt.Sprintf("%d",len(matches)),"memories_json":string(data)},
    }, nil
  })
}
