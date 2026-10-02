package orchestrator

import "github.com/divibisoul/Orquestrador-/protocol"

type AgentDescriptor struct {
	ID           string   `json:"id"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities"`
	Tools        []string `json:"tools"`
	Inputs       []string `json:"inputs"`
	Outputs      []string `json:"outputs"`
}

func N07Agents() []AgentDescriptor {
	return []AgentDescriptor{
		{ID: "n07.discovery", Role: "discovery", Capabilities: []string{"mesh.discovery", "mesh.capabilities", "mesh.capability.resolve@1.0.0"}, Inputs: []string{"capability", "peer-state"}, Outputs: []string{"capability-owner", "peer-candidates"}},
		{ID: "n07.affinity-resolver", Role: "functional-affinity-routing", Capabilities: []string{"mesh.capability.resolve@1.0.0", "dynamic-affinity-routing"}, Tools: []string{"SOUL capability authority", "OSS affinity manifest", "real peer discovery"}, Inputs: []string{"capability", "correlationId", "configured-peers"}, Outputs: []string{"REAL/PROJECTED/BLOCKED/UNMEASURABLE target states", "affinity-ranked routes"}},
		{ID: "n07.external-planning", Role: "external-planning-provenance", Capabilities: []string{"cognitive.planning.sources@1.0.0"}, Tools: []string{"BabyAGI pattern", "AutoGPT pattern", "DeerFlow subagent pattern", "SuperAGI queue pattern"}, Inputs: []string{"planning-source-catalog-request"}, Outputs: []string{"pinned-provenance-catalog"}},
		{ID: "n07.router", Role: "router", Capabilities: []string{"mesh.delegate", "dynamic-routing"}, Inputs: []string{"capability", "health", "latency", "load"}, Outputs: []string{"selected-peer", "route-score"}},
		{ID: "n07.executor", Role: "executor", Capabilities: []string{"local-execution", "mesh.supergpu.execute", "mesh.supergpu.parallel"}, Inputs: []string{"task", "payload"}, Outputs: []string{"result", "duration"}},
		{ID: "n07.composer", Role: "composer", Capabilities: []string{"mesh.fusion.execute", "capability-composition"}, Inputs: []string{"component-capabilities", "dependencies"}, Outputs: []string{"composed-result", "component-trace"}},
		{ID: "n07.storage", Role: "content-storage", Capabilities: []string{"storage.web3.upload@1.0.0", "storage.web3.status@1.0.0"}, Tools: []string{"web3.storage", "IPFS", "Filecoin"}, Inputs: []string{"file", "content-addressed-data", "cid"}, Outputs: []string{"cid", "storage-status", "ipfs-reference"}},
		{ID: "n07.validator", Role: "validator", Capabilities: []string{"result-validation", "contract-validation"}, Inputs: []string{"result", "correlationId", "contractVersion"}, Outputs: []string{"validated-result", "validation-error"}},
		{ID: "n07.observer", Role: "observer", Capabilities: []string{"mesh.health", "metrics", "tracing"}, Inputs: []string{"events", "latency", "errors"}, Outputs: []string{"health", "metrics", "trace"}},
		{ID: "n07.jev", Role: "decision-specialist", Capabilities: []string{"jev.systemone@1.0.0"}, Tools: []string{"Jev/System One"}, Inputs: []string{"state", "questions"}, Outputs: []string{"typed-decisions", "probabilities", "confidence"}},
		{ID: "n07.orbital-reasoner", Role: "orbital-resource-simulation", Capabilities: []string{"prefrontal.orbital.evaluate@1.0.0", "transcendental.estimate@1.0.0"}, Tools: []string{"TCE deterministic simulation"}, Inputs: []string{"workloads", "candidate", "neural-signal"}, Outputs: []string{"simulation-evidence", "prefrontal-decision", "resource-estimate"}},
		{ID: "n07.cooperation", Role: "cooperative-control-plane", Capabilities: []string{"cooperation.handshake@1.0.0", "cooperation.exchange@1.0.0", "cooperation.health@1.0.0"}, Tools: []string{"Mesh discovery", "capability negotiation", "correlation-preserving exchange", "learning route observer"}, Inputs: []string{"target", "capability", "payload", "correlation"}, Outputs: []string{"handshake", "exchange result", "route learning observation"}},
		{ID: "soul.engineer.provenance", Role: "provenance-engineer", Capabilities: []string{"source-verification", "revision-pinning", "attribution"}, Tools: []string{"external capability registry", "git provenance"}, Inputs: []string{"provider", "revision"}, Outputs: []string{"provenance-record", "source-state"}},
		{ID: "soul.engineer.adapter", Role: "capability-adapter-engineer", Capabilities: []string{"capability-contract", "adapter-design", "authority-preservation"}, Tools: []string{"SOUL capability contract", "OSS capability manifest"}, Inputs: []string{"provider", "capability", "target"}, Outputs: []string{"adapter-plan", "activation-boundary"}},
		{ID: "soul.engineer.runtime", Role: "runtime-integration-engineer", Capabilities: []string{"mesh-routing", "tool-binding", "dependency-isolation"}, Tools: []string{"canonical Mesh", "SuperGPU", "sandbox"}, Inputs: []string{"adapter", "target", "configuration"}, Outputs: []string{"runtime-route", "activation-state"}},
		{ID: "soul.engineer.verification", Role: "verification-engineer", Capabilities: []string{"contract-tests", "smoke-tests", "e2e-verification"}, Tools: []string{"GitHub Actions", "runtime evidence"}, Inputs: []string{"integration", "test-contract"}, Outputs: []string{"evidence-state", "failure-record"}},
		{ID: "soul.engineer.resilience", Role: "resilience-observability-engineer", Capabilities: []string{"fail-closed", "correlation", "bounded-execution", "observability", "re-audit"}, Tools: []string{"Clareira", "Mesh telemetry", "circuit breaker"}, Inputs: []string{"execution-event", "health", "failure"}, Outputs: []string{"health-state", "trace", "re-audit-trigger"}},
	}
}

func N07Identity() map[string]any {
	return map[string]any{"nucleus": protocol.N07, "identity": "SOUL-N07-Orchestrator", "role": "distributed-orchestrator", "independent": true, "basePeers": []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06}, "agents": N07Agents(), "execution": []string{"local", "delegated", "parallel", "composed"}, "memory": []string{"route-state", "peer-health", "correlation", "metrics", "content-addressed-storage"}, "input": []string{"SOUL_MESSAGE", "CAPABILITY_REQUEST", "TASK", "EVENT", "CONTENT"}, "output": []string{"TASK_RESULT", "ERROR", "EVENT", "METRICS", "CID"}, "discovery": true, "delegation": true, "composition": true, "observability": true}
}
