
package aeternum

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/orchestrator"
)

type Status string

const (
	StatusNative   Status = "NATIVE"
	StatusAdapter  Status = "ADAPTER_REQUIRED"
	StatusBlocked  Status = "BLOCKED_INFRASTRUCTURE"
	StatusUnmapped Status = "UNMAPPED"
)

type ProcessorSpec struct {
	ID           string
	Function     string
	Authority    string
	Dependencies []string
	Notes        string
}

type ModuleSpec struct {
	ID        string
	Group     string
	Function  string
	Authority string
	Status    Status
	Operation string
	Notes     string
}

type HortaCore struct {
	engine  *orchestrator.Engine
	sara    *backend.SARAProxy
	modules map[string]ModuleSpec
}

func NewHortaCore(engine *orchestrator.Engine, sara *backend.SARAProxy) (*HortaCore, error) {
	if engine == nil {
		return nil, errors.New("N07 engine is required")
	}
	h := &HortaCore{engine: engine, sara: sara, modules: make(map[string]ModuleSpec)}
	for _, spec := range moduleCatalog() {
		if _, exists := h.modules[spec.ID]; exists {
			return nil, errors.New("duplicate Aeternum module: " + spec.ID)
		}
		h.modules[spec.ID] = spec
	}
	return h, nil
}

func (h *HortaCore) Processors() []ProcessorSpec {
	return []ProcessorSpec{
		{ID: "codex", Function: "context/memory/provenance", Authority: "SARA + N07", Dependencies: []string{"SARA RegenerativeMemory", "SARA TemporalVectorDB", "N07 capability inventory"}, Notes: "No duplicate memory authority is created in N07."},
		{ID: "blueprint", Function: "planning/admission/orchestration", Authority: "N07", Dependencies: []string{"N07 Orchestrator", "PrefrontalCortex"}, Notes: "Canonical executive planning remains N07."},
		{ID: "eru", Function: "reversibility/state comparison/meta-evidence", Authority: "SARA", Dependencies: []string{"ERU_Engine", "DecisionTrace", "ProvenanceTracker"}, Notes: "Uses the existing SARA ERU rather than a second implementation."},
		{ID: "audit", Function: "governance/validation/anomaly evidence", Authority: "SARA", Dependencies: []string{"GovernedSARA", "EthicalFilterChain", "CycleAuditor"}, Notes: "No keyword-only fake approval layer is introduced."},
		{ID: "guide", Function: "capability discovery/architecture explanation", Authority: "N07 + SARA", Dependencies: []string{"Mesh discovery", "SARA capabilities"}, Notes: "Guidance exposes observed capability state, not fabricated liveness."},
	}
}

func (h *HortaCore) Capabilities() []ModuleSpec {
	out := make([]ModuleSpec, 0, len(h.modules))
	for _, module := range h.modules {
		out = append(out, module)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (h *HortaCore) Health() map[string]any {
	saraConfigured := h.sara != nil && h.sara.Configured()
	status := "DEGRADED"
	if saraConfigured && h.engine.Status() == "ready" {
		status = "READY"
	}
	return map[string]any{
		"system": "AETERNUM_HORTACORE",
		"status": status,
		"module_count": len(h.modules),
		"processor_count": len(h.Processors()),
		"sara_configured": saraConfigured,
		"n07_engine_status": h.engine.Status(),
		"timestamp": time.Now().UTC(),
		"principles": []string{
			"no_mock",
			"no_fabricated_success",
			"single_authority_per_responsibility",
			"blocked_capabilities_remain_explicit",
		},
	}
}

func (h *HortaCore) Execute(ctx context.Context, moduleID string, payload []float64, metadata map[string]string) (map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	id := strings.TrimSpace(moduleID)
	spec, ok := h.modules[id]
	if !ok {
		return nil, errors.New("unknown Aeternum module: " + id)
	}
	switch spec.Status {
	case StatusNative:
		result, err := h.engine.Execute(ctx, spec.Operation, payload, metadata)
		return map[string]any{"module": id, "status": string(spec.Status), "operation": spec.Operation, "result": result}, err
	case StatusAdapter:
		return nil, errors.New("AETERNUM_ADAPTER_REQUIRED:" + id)
	case StatusBlocked:
		return nil, errors.New("AETERNUM_BLOCKED_INFRASTRUCTURE:" + id)
	default:
		return nil, errors.New("AETERNUM_UNMAPPED:" + id)
	}
}

func moduleCatalog() []ModuleSpec {
	const (
		native  = StatusNative
		adapter = StatusAdapter
		blocked = StatusBlocked
	)
	return []ModuleSpec{
		{ID: "acai", Group: "perception", Function: "audio/text intent and conversational interpretation", Authority: "N03/N04 native runtimes", Status: adapter, Notes: "Exact peer capability contract must be discovered before execution."},
		{ID: "mpvs", Group: "perception", Function: "visual perception", Authority: "N04 native runtime", Status: adapter, Notes: "No synthetic object detector is introduced."},
		{ID: "multimodal_cortex", Group: "perception", Function: "multimodal fusion", Authority: "N04 native runtime", Status: adapter, Notes: "Requires the peer real multimodal capability contract."},
		{ID: "autonomous_embodiment", Group: "perception", Function: "physical/digital actuation", Authority: "native hardware runtime", Status: blocked, Notes: "No hardware backend is present in N07; coordinates are never fabricated."},
		{ID: "neural_forge", Group: "advanced", Function: "neural architecture generation", Authority: "N07 neural/composition", Status: adapter, Notes: "Existing neural runtime executes numerical graphs; it is not a code generator."},
		{ID: "asc", Group: "advanced", Function: "scientific hypothesis methodology", Authority: "research-capable peer/runtime", Status: blocked, Notes: "No real scientific execution backend is owned by N07."},
		{ID: "biomolecular_designer", Group: "advanced", Function: "biomolecular design", Authority: "external scientific backend", Status: blocked, Notes: "Affinity and structure outputs require a real validated backend."},
		{ID: "reality_synthesis", Group: "advanced", Function: "environment/physics synthesis", Authority: "external simulation/graphics backend", Status: blocked, Notes: "No fabricated physics engine is admitted."},
		{ID: "strategic_planning", Group: "advanced", Function: "constraint-aware planning", Authority: "N07 Prefrontal + Orchestrator", Status: adapter, Notes: "Use prefrontal admission and orchestration contracts; no fixed optimality percentage."},
		{ID: "csae", Group: "optimization", Function: "architecture optimization", Authority: "N07 fusion/orchestrator", Status: adapter, Notes: "Optimization must be backed by measured alternatives and execution evidence."},
		{ID: "dcrs", Group: "optimization", Function: "resource allocation", Authority: "N07 SuperGPU/compute", Status: adapter, Notes: "Allocation is real only when a device lease is granted."},
		{ID: "adaptation_module", Group: "optimization", Function: "parameter adaptation", Authority: "N07 neural learning", Status: adapter, Notes: "Uses neural.learn when a concrete learning contract is supplied."},
		{ID: "scre", Group: "optimization", Function: "code refactoring/generation", Authority: "code-capable peer or external provider", Status: adapter, Notes: "No placeholder source is accepted as refactoring."},
		{ID: "ecas", Group: "optimization", Function: "component/architecture composition", Authority: "N07 mesh.fusion", Status: adapter, Notes: "Composition executes only through registered fusion components."},
		{ID: "eus", Group: "cognitive", Function: "cross-domain knowledge synthesis", Authority: "SARA + N01..N07 capability graph", Status: adapter, Notes: "Knowledge must derive from real indexed evidence."},
		{ID: "mlfg", Group: "cognitive", Function: "meta-learning framework generation", Authority: "N07 neural + external training runtime", Status: adapter, Notes: "No hardcoded framework name is treated as generated evidence."},
		{ID: "emergent_cognition", Group: "cognitive", Function: "multi-capability synthesis", Authority: "N07 fusion", Status: adapter, Notes: "Confidence must derive from contributing component evidence."},
		{ID: "bnc_v2", Group: "cognitive", Function: "plastic neural architecture", Authority: "N07 neural", Status: adapter, Notes: "Real metrics come from Neural.Health/LearningSteps, not fixed Hz."},
		{ID: "skill_acquisition", Group: "cognitive", Function: "skill learning", Authority: "N02/N07 learning runtime", Status: adapter, Notes: "Requires an explicit training/evaluation contract."},
		{ID: "uci", Group: "infrastructure", Function: "protocol translation/interconnect", Authority: "N07 Mesh", Status: native, Operation: "mesh.health@1.0.0", Notes: "Exposes canonical Mesh health rather than synthetic throughput."},
		{ID: "ethical_governance", Group: "governance", Function: "ethical/governance review", Authority: "SARA GovernedSARA", Status: adapter, Notes: "Delegates to real SARA governance; never hardcodes approval."},
		{ID: "strategic_defense", Group: "governance", Function: "threat/security analysis", Authority: "SARA security + SOUL Mesh security", Status: adapter, Notes: "Threat-free status requires real observations."},
		{ID: "existential_safety", Group: "governance", Function: "high-impact safety gate", Authority: "SARA + N07 Prefrontal", Status: adapter, Notes: "Risk is evidence-derived; no fixed epsilon is emitted."},
		{ID: "einstein_reasoning", Group: "einstein", Function: "theory/logic validation", Authority: "N07 cognitive + scientific providers", Status: adapter, Notes: "Validation requires explicit premises and evidence."},
		{ID: "einstein_code", Group: "einstein", Function: "code generation/audit", Authority: "code-capable provider", Status: adapter, Notes: "No placeholder source code is generated as success."},
		{ID: "einstein_quantum", Group: "einstein", Function: "quantum computation", Authority: "real quantum backend", Status: blocked, Notes: "N07 exposes CPU/SuperGPU only; it does not fabricate qubits or fidelity."},
		{ID: "cot_arhd", Group: "cot", Function: "resource/latency anticipation", Authority: "N07 observability + router", Status: adapter, Notes: "Requires measured history rather than fixed predictions."},
		{ID: "cot_drc", Group: "cot", Function: "task decomposition/parallelism", Authority: "N07 Orchestrator/Federation", Status: adapter, Notes: "Uses bounded workers and cancellation; no fixed parallelism percentage."},
		{ID: "cot_area", Group: "cot", Function: "algorithm evolution", Authority: "N07 neural learning", Status: adapter, Notes: "Requires real training and evaluation samples."},
	}
}

func ValidateCatalog() error {
	catalog := moduleCatalog()
	seen := make(map[string]struct{}, len(catalog))
	for _, spec := range catalog {
		if strings.TrimSpace(spec.ID) == "" || strings.TrimSpace(spec.Group) == "" || strings.TrimSpace(spec.Function) == "" {
			return errors.New("invalid Aeternum module specification")
		}
		if _, ok := seen[spec.ID]; ok {
			return errors.New("duplicate Aeternum module: " + spec.ID)
		}
		seen[spec.ID] = struct{}{}
		if spec.Status == StatusNative && strings.TrimSpace(spec.Operation) == "" {
			return errors.New("native Aeternum module requires a real operation: " + spec.ID)
		}
	}
	if len(catalog) != 29 {
		return errors.New("Aeternum catalog must contain exactly 29 modules")
	}
	return nil
}
