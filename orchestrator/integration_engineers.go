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
	IntegrationEngineersDescribeOperation = "soul.integration.engineers.describe@1.0.0"
	IntegrationEngineerPrepareOperation   = "soul.integration.engineer.prepare@1.0.0"
)

type IntegrationEngineer struct {
	ID           string   `json:"id"`
	Role         string   `json:"role"`
	Scope        []string `json:"scope"`
	Capabilities []string `json:"capabilities"`
	Inputs       []string `json:"inputs"`
	Outputs      []string `json:"outputs"`
	Methodology  string   `json:"methodology"`
}

var integrationEngineers = []IntegrationEngineer{
	{ID: "soul.engineer.fabric", Role: "cross-repository-fusion-engineer", Scope: []string{"all-25-repositories"}, Capabilities: []string{"capability-composition", "cross-repo-reinforcement", "compatibility-analysis", "dependency-isolation"}, Inputs: []string{"provider-graph", "component-graph", "capability-request"}, Outputs: []string{"augmentation-plan", "compatibility-plan", "reconciliation-trigger"}, Methodology: "superpowers.octacore"},
	{ID: "soul.engineer.provenance", Role: "provenance-engineer", Scope: []string{"all-25-upstreams", "all-9-soul-components"}, Capabilities: []string{"source-verification", "revision-pinning", "attribution"}, Inputs: []string{"provider", "revision"}, Outputs: []string{"provenance-record", "source-state"}, Methodology: "superpowers.sara-trinity"},
	{ID: "soul.engineer.adapter", Role: "capability-adapter-engineer", Scope: []string{"all-25-upstreams"}, Capabilities: []string{"capability-contract", "adapter-design", "authority-preservation"}, Inputs: []string{"provider", "capability", "target"}, Outputs: []string{"adapter-plan", "activation-boundary"}, Methodology: "superpowers.octacore"},
	{ID: "soul.engineer.runtime", Role: "runtime-integration-engineer", Scope: []string{"N01", "N02", "N03", "N04", "N05", "N06", "N07", "SARA", "JEV"}, Capabilities: []string{"mesh-routing", "tool-binding", "sandboxed-execution", "dependency-isolation"}, Inputs: []string{"adapter", "target", "configuration"}, Outputs: []string{"runtime-route", "activation-state"}, Methodology: "superpowers.cortex-orbital-supergpu"},
	{ID: "soul.engineer.verification", Role: "verification-engineer", Scope: []string{"all-integrations"}, Capabilities: []string{"contract-tests", "smoke-tests", "e2e-verification", "verification-before-completion"}, Inputs: []string{"integration", "test-contract"}, Outputs: []string{"evidence-state", "failure-record"}, Methodology: "superpowers.sara-trinity"},
	{ID: "soul.engineer.resilience", Role: "resilience-observability-engineer", Scope: []string{"Mesh", "Clareira", "SuperGPU", "Octacore", "SARA"}, Capabilities: []string{"fail-closed", "correlation", "bounded-execution", "observability", "re-audit"}, Inputs: []string{"execution-event", "health", "failure"}, Outputs: []string{"health-state", "trace", "re-audit-trigger"}, Methodology: "superpowers.mesh-clareira"},
}

func IntegrationEngineers() []IntegrationEngineer {
	out := make([]IntegrationEngineer, len(integrationEngineers))
	copy(out, integrationEngineers)
	return out
}

func RegisterIntegrationEngineeringOperations(e *Engine) error {
	if e == nil {
		return errors.New("engine is nil")
	}
	if err := e.Register(IntegrationEngineersDescribeOperation, integrationEngineersDescribe); err != nil {
		return err
	}
	return e.Register(IntegrationEngineerPrepareOperation, integrationEngineerPrepare)
}

func integrationEngineersDescribe(ctx context.Context, m protocol.Message) (protocol.Result, error) {
	if err := ctx.Err(); err != nil {
		return protocol.Result{}, err
	}
	raw, err := json.Marshal(map[string]any{
		"system":             "SOUL",
		"canonical_owner":    "N07",
		"purpose":            "solidify functional integration rather than passive repository attachment",
		"agents":             IntegrationEngineers(),
		"evidence_states":    []string{"REAL", "PROJECTED", "BLOCKED", "UNMEASURABLE"},
		"mandatory_sequence": []string{"provenance", "capability-contract", "adapter", "routing", "controlled-execution", "evidence", "observability", "re-audit"},
	})
	if err != nil {
		return protocol.Result{}, err
	}
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.integration-engineering", Target: m.Source, Status: "ok", Metadata: map[string]string{"engineering_agents_json": string(raw)}}, nil
}

func integrationEngineerPrepare(ctx context.Context, m protocol.Message) (protocol.Result, error) {
	if err := ctx.Err(); err != nil {
		return protocol.Result{}, err
	}
	agentID := strings.TrimSpace(m.Metadata["engineer_id"])
	provider := strings.TrimSpace(m.Metadata["provider"])
	if agentID == "" || provider == "" {
		return integrationEngineerFail(m, "INTEGRATION_ENGINEER_INPUT_REQUIRED", errors.New("engineer_id and provider are required"))
	}
	var agent *IntegrationEngineer
	for i := range integrationEngineers {
		if integrationEngineers[i].ID == agentID {
			agent = &integrationEngineers[i]
			break
		}
	}
	if agent == nil {
		return integrationEngineerFail(m, "INTEGRATION_ENGINEER_NOT_FOUND", errors.New(agentID))
	}
	scopeOK := false
	for _, s := range agent.Scope {
		if s == "all-integrations" || s == "all-25-upstreams" || s == "all-9-soul-components" || s == provider {
			scopeOK = true
			break
		}
	}
	if !scopeOK {
		return integrationEngineerFail(m, "INTEGRATION_ENGINEER_SCOPE_MISMATCH", fmt.Errorf("%s cannot prepare %s", agentID, provider))
	}
	plan := map[string]any{
		"engineer":                agent,
		"provider":                provider,
		"requested_target":        strings.TrimSpace(m.Metadata["target"]),
		"requested_capability":    strings.TrimSpace(m.Metadata["capability"]),
		"required_gates":          []string{"contract", "adapter", "routing", "verification", "observability"},
		"fail_closed":             true,
		"no_fake_runtime_success": true,
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return integrationEngineerFail(m, "INTEGRATION_ENGINEER_PLAN_ENCODE_FAILED", err)
	}
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.integration-engineering", Target: m.Source, Status: "ok", Metadata: map[string]string{"plan_json": string(raw)}}, nil
}

func integrationEngineerFail(m protocol.Message, code string, err error) (protocol.Result, error) {
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.integration-engineering", Target: m.Source, Status: "error", Error: code + ":" + err.Error()}, err
}
