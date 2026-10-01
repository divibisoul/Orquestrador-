package octacore

// AgentProfile records an upstream agent/tool contribution without importing or
// copying upstream runtime code into N07. Invocation occurs through canonical
// N02 capability handlers over the SOUL Mesh.
type AgentProfile struct {
	Name          string   `json:"name"`
	Source        string   `json:"source"`
	SourcePaths   []string `json:"source_paths"`
	Capability    string   `json:"capability"`
	Role          string   `json:"role"`
	TargetNucleus string   `json:"target_nucleus"`
	AdapterOwner  string   `json:"adapter_owner"`
	UseInCore     string   `json:"use_in_core"`
}

var agentProfiles = []AgentProfile{
	{
		Name:          "SuperAGI Task Queue / Agent Iteration",
		Source:        "TransformerOptimus/SuperAGI",
		SourcePaths:   []string{"superagi/agent/task_queue.py", "superagi/agent/agent_iteration_step_handler.py", "superagi/agent/queue_step_handler.py"},
		Capability:    "mlfg",
		Role:         "meta-learning design and bounded task iteration advisory",
		TargetNucleus: "N06",
		AdapterOwner: "N02",
		UseInCore:     "optional collaborative advisory stage before resource estimation",
	},
	{
		Name:          "SuperAGI Tool Executor",
		Source:        "TransformerOptimus/SuperAGI",
		SourcePaths:   []string{"superagi/agent/tool_executor.py", "superagi/agent/tool_builder.py"},
		Capability:    "skill_acquisition",
		Role:          "reusable skill specification from demonstrations",
		TargetNucleus: "N04",
		AdapterOwner: "N02",
		UseInCore:     "optional collaborative skill-learning stage when demonstrations are supplied",
	},
	{
		Name:          "Xun function-tool / sub-agent model",
		Source:        "MenxLi/xun",
		SourcePaths:   []string{"src/xun/toolbox.py", "src/xun/toolcall.py", "src/xun/agent_factory.py"},
		Capability:    "emergent_cognition",
		Role:          "bounded subtask decomposition and synthesis",
		TargetNucleus: "N07",
		AdapterOwner: "N02",
		UseInCore:     "optional collaborative decomposition stage over Mesh",
	},
	{
		Name:          "Hermes protocol adapter pattern",
		Source:        "NousResearch/hermes-agent",
		SourcePaths:   []string{"agent/acp_openai_bridge.py", "agent/backend_identity.py", "agent/delegation_context.py"},
		Capability:    "uci",
		Role:          "universal protocol/capability translation advisory",
		TargetNucleus: "N04",
		AdapterOwner: "N02",
		UseInCore:     "optional contract-translation stage for heterogeneous tool requests",
	},
	{
		Name:          "CrewAI tool-hooks / agent-task composition",
		Source:        "crewAIInc/crewAI",
		SourcePaths:   []string{"src/crewai/agent.py", "src/crewai/task.py", "src/crewai/tools/base_tool.py"},
		Capability:    "scre",
		Role:          "code/task review proposal and verification planning",
		TargetNucleus: "N04",
		UseInCore:     "optional verification advisory stage before dispatch",
	},
}

func AgentProfiles() []AgentProfile {
	out := make([]AgentProfile, len(agentProfiles))
	copy(out, agentProfiles)
	for i := range out {
		out[i].SourcePaths = append([]string(nil), agentProfiles[i].SourcePaths...)
	}
	return out
}

func FindAgentProfile(capability string) *AgentProfile {
	for i := range agentProfiles {
		if agentProfiles[i].Capability == capability {
			profile := agentProfiles[i]
			profile.SourcePaths = append([]string(nil), profile.SourcePaths...)
			return &profile
		}
	}
	return nil
}
