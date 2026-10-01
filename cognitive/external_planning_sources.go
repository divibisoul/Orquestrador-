package cognitive

// ExternalPlanningSource records a planning primitive discovered in an external
// source. It is descriptive/provenance metadata; execution remains owned by
// the canonical N07 Planner and MeshExecutor.
type ExternalPlanningSource struct {
	ID             string
	Repository     string
	SourcePath     string
	SourceRevision string
	Pattern        string
	UseInN07       string
	Strategy       string
	LicenseScope   string
}

// ExternalPlanningSources contains the planning primitives selected for
// comparison against N07's existing Goal -> Plan -> Act flow.
var ExternalPlanningSources = []ExternalPlanningSource{
	{
		ID:             "BABYAGI-TASK-LOOP",
		Repository:     "yoheinakajima/babyagi",
		SourcePath:     "babyagi/functionz",
		SourceRevision: "fa8930ebe72a82e5ad57b356e7cbec96290e5bb2",
		Pattern:        "task/function registration and execution loop",
		UseInN07:       "compare task decomposition and function execution with Goal/Step",
		Strategy:       "SELECTIVE",
		LicenseScope:   "VERIFY SOURCE FILE LICENSE",
	},
	{
		ID:             "AUTOGPT-PLAN-EXECUTE",
		Repository:     "Significant-Gravitas/AutoGPT",
		SourcePath:     "classic/original_autogpt/autogpt/agents/prompt_strategies/plan_execute.py",
		SourceRevision: "b958f5ba76a558952fcbe976f66c273436e6642c",
		Pattern:        "plan/execute decomposition and re-planning strategy",
		UseInN07:       "compare explicit planning/replanning stages without creating a second planner",
		Strategy:       "SELECTIVE",
		LicenseScope:   "MIT (classic path)",
	},
	{
		ID:             "DEERFLOW-SUBAGENTS",
		Repository:     "bytedance/deer-flow",
		SourcePath:     "backend/packages/harness/deerflow/subagents",
		SourceRevision: "d3e8e78f3f6e841990429980070d2ccaecb987b8",
		Pattern:        "subagent registry/runtime/batch execution",
		UseInN07:       "compare delegated subtask execution with Mesh peer ownership",
		Strategy:       "ADAPTER",
		LicenseScope:   "MIT",
	},
	{
		ID:             "SUPERAGI-TASK-QUEUE",
		Repository:     "TransformerOptimus/SuperAGI",
		SourcePath:     "superagi/agent/task_queue.py",
		SourceRevision: "6acddf3f3cde7d46724a6cb84562fa6a476ff110",
		Pattern:        "persistent current/completed task queues with explicit status",
		UseInN07:       "compare queue semantics with existing finite execution queue and RunStore",
		Strategy:       "SELECTIVE",
		LicenseScope:   "MIT",
	},
	{
		ID:             "DEERFLOW-BATCH-TASK",
		Repository:     "bytedance/deer-flow",
		SourcePath:     "backend/packages/harness/deerflow/tools/builtins/batch_task_tool.py",
		SourceRevision: "d3e8e78f3f6e841990429980070d2ccaecb987b8",
		Pattern:        "bounded batch task dispatch",
		UseInN07:       "compare bounded parallel delegation against existing SuperGPU helpers",
		Strategy:       "ADAPTER",
		LicenseScope:   "MIT",
	},
]

// ExternalPlanningSourceIDs returns a stable copy suitable for diagnostics.
func ExternalPlanningSourceIDs() []string {
	out := make([]string, 0, len(ExternalPlanningSources))
	for _, source := range ExternalPlanningSources {
		out = append(out, source.ID)
	}
	return out
}
