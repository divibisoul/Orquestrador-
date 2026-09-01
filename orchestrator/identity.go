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
		{ID: "n07.discovery", Role: "discovery", Capabilities: []string{"mesh.discovery", "mesh.capabilities", "mesh.capability.resolve"}, Inputs: []string{"capability", "peer-state"}, Outputs: []string{"capability-owner", "peer-candidates"}},
		{ID: "n07.router", Role: "router", Capabilities: []string{"mesh.delegate", "dynamic-routing"}, Inputs: []string{"capability", "health", "latency", "load"}, Outputs: []string{"selected-peer", "route-score"}},
		{ID: "n07.executor", Role: "executor", Capabilities: []string{"local-execution", "mesh.supergpu.execute", "mesh.supergpu.parallel"}, Inputs: []string{"task", "payload"}, Outputs: []string{"result", "duration"}},
		{ID: "n07.composer", Role: "composer", Capabilities: []string{"mesh.fusion.execute", "capability-composition"}, Inputs: []string{"component-capabilities", "dependencies"}, Outputs: []string{"composed-result", "component-trace"}},
		{ID: "n07.storage", Role: "content-storage", Capabilities: []string{"storage.web3.upload@1.0.0", "storage.web3.status@1.0.0"}, Tools: []string{"web3.storage", "IPFS", "Filecoin"}, Inputs: []string{"file", "content-addressed-data", "cid"}, Outputs: []string{"cid", "storage-status", "ipfs-reference"}},
		{ID: "n07.validator", Role: "validator", Capabilities: []string{"result-validation", "contract-validation"}, Inputs: []string{"result", "correlationId", "contractVersion"}, Outputs: []string{"validated-result", "validation-error"}},
		{ID: "n07.observer", Role: "observer", Capabilities: []string{"mesh.health", "metrics", "tracing"}, Inputs: []string{"events", "latency", "errors"}, Outputs: []string{"health", "metrics", "trace"}},
	}
}

func N07Identity() map[string]any {
	return map[string]any{"nucleus": protocol.N07, "identity": "SOUL-N07-Orchestrator", "role": "distributed-orchestrator", "independent": true, "basePeers": []string{protocol.N01, protocol.N02, protocol.N03, protocol.N04, protocol.N05, protocol.N06}, "agents": N07Agents(), "execution": []string{"local", "delegated", "parallel", "composed"}, "memory": []string{"route-state", "peer-health", "correlation", "metrics", "content-addressed-storage"}, "input": []string{"SOUL_MESSAGE", "CAPABILITY_REQUEST", "TASK", "EVENT", "CONTENT"}, "output": []string{"TASK_RESULT", "ERROR", "EVENT", "METRICS", "CID"}, "discovery": true, "delegation": true, "composition": true, "observability": true}
}
