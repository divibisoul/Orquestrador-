package orchestrator

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strings"

  "github.com/divibisoul/Orquestrador-/protocol"
)

const (
  SuperpowersAgentRouteOperation = "superpowers.agent.route@1.0.0"
  SuperpowersAgentDescribeOperation = "superpowers.agent.describe@1.0.0"
)

type SuperpowersAgentBinding struct {
  ID string
  Bridge string
  Role string
  Targets []string
  Operations []string
  Skills []string
  Reporting []string
  FailClosed bool
}

var superpowersBindings = []SuperpowersAgentBinding{
  {ID:"superpowers.sara-trinity", Bridge:"SARA", Role:"regenerative-engineering-agent", Targets:[]string{"ARA","ETR","ITR","RGO","MMD"}, Operations:[]string{"sara.cycle","sara.audit","sara.regenerate","rgo.ingest"}, Skills:[]string{"brainstorming","writing-plans","subagent-driven-development","test-driven-development","systematic-debugging","requesting-code-review","verification-before-completion"}, FailClosed:true},
  {ID:"superpowers.cortex-orbital-supergpu", Bridge:"N07", Role:"executive-reasoning-compute-agent", Targets:[]string{"prefrontal-neocortex","orbital-reasoning","SuperGPU"}, Operations:[]string{"transcendental.estimate@1.0.0","prefrontal.orbital.evaluate@1.0.0","supergpu.execute@1.0.0"}, Skills:[]string{"brainstorming","writing-plans","subagent-driven-development","test-driven-development","systematic-debugging","verification-before-completion"}, FailClosed:true},
  {ID:"superpowers.mesh-clareira", Bridge:"Mesh", Role:"federation-observer-agent", Targets:[]string{"SOUL-Mesh","Projeto-Clareira"}, Operations:[]string{"mesh.discovery","mesh.capability.resolve@1.0.0","mesh.health"}, Skills:[]string{"systematic-debugging","verification-before-completion","requesting-code-review"}, Reporting:[]string{"clareira.ingest","clareira.metrics"}, FailClosed:true},
  {ID:"superpowers.octacore", Bridge:"N07", Role:"octacore-coordination-agent", Targets:[]string{"Octacore","Mesh","SuperGPU"}, Operations:[]string{"octacore.fusion.describe@1.0.0","octacore.fusion.execute@1.0.0","octacore.submit","octacore.batch"}, Skills:[]string{"writing-plans","subagent-driven-development","test-driven-development","systematic-debugging","verification-before-completion"}, FailClosed:true},
}

func SuperpowersAgentBindings() []SuperpowersAgentBinding {
  out := make([]SuperpowersAgentBinding, len(superpowersBindings))
  copy(out, superpowersBindings)
  return out
}

func RegisterSuperpowersAgentOperations(e *Engine) error {
  if e == nil { return errors.New("engine is nil") }
  if err := e.Register(SuperpowersAgentDescribeOperation, superpowersAgentDescribe); err != nil { return err }
  return e.Register(SuperpowersAgentRouteOperation, superpowersAgentRoute(e))
}

func superpowersAgentDescribe(ctx context.Context, m protocol.Message) (protocol.Result, error) {
  raw, err := json.Marshal(map[string]any{
    "provider": map[string]any{
      "id":"superpowers",
      "source":"https://github.com/obra/superpowers",
      "revision":"8ca22dba9a94f28898bbce59f2537ff4d87c747d",
      "integration_mode":"upstream-methodology-via-explicit-adapter",
      "runtime_activation_requires_explicit_adapter":true,
    },
    "agents": SuperpowersAgentBindings(),
    "policy": map[string]any{"no_fake_runtime_success":true,"preserve_upstream_attribution":true,"preserve_existing_authority":true,"no_secondary_mesh":true,"no_secondary_sara_authority":true},
  })
  if err != nil { return protocol.Result{}, err }
  return protocol.Result{TraceID:m.TraceID, CorrelationID:m.CorrelationID, Source:"N07.superpowers", Target:m.Source, Status:"ok", Metadata:map[string]string{"bindings_json":string(raw)}}, nil
}

func superpowersAgentRoute(e *Engine) Handler {
  return func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
    agentID := strings.TrimSpace(m.Metadata["agent_id"])
    operation := strings.TrimSpace(m.Metadata["operation"])
    if agentID == "" || operation == "" { return superpowersAgentFail(m, "SUPERPOWERS_AGENT_INPUT_REQUIRED", errors.New("agent_id and operation are required")) }
    var binding *SuperpowersAgentBinding
    for i := range superpowersBindings {
      if superpowersBindings[i].ID == agentID { binding = &superpowersBindings[i]; break }
    }
    if binding == nil { return superpowersAgentFail(m, "SUPERPOWERS_AGENT_NOT_FOUND", errors.New(agentID)) }
    allowed := false
    for _, op := range binding.Operations { if op == operation { allowed = true; break } }
    if !allowed { return superpowersAgentFail(m, "SUPERPOWERS_AGENT_OPERATION_NOT_BOUND", fmt.Errorf("%s -> %s", agentID, operation)) }
    result, err := e.Execute(ctx, operation, m.Payload, m.Metadata)
    if err != nil { return superpowersAgentFail(m, "SUPERPOWERS_AGENT_DELEGATE_FAILED", err) }
    if result.Status != "ok" { return superpowersAgentFail(m, "SUPERPOWERS_AGENT_DELEGATE_NOT_OK", errors.New(result.Error)) }
    if result.Metadata == nil { result.Metadata = map[string]string{} }
    result.Metadata["superpowers_agent"] = agentID
    result.Metadata["superpowers_skill_boundary"] = strings.Join(binding.Skills, ",")
    return result, nil
  }
}

func superpowersAgentFail(m protocol.Message, code string, err error) (protocol.Result, error) {
  return protocol.Result{TraceID:m.TraceID, CorrelationID:m.CorrelationID, Source:"N07.superpowers", Target:m.Source, Status:"error", Error:code+":"+err.Error()}, err
}
