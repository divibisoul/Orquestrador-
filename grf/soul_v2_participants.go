package grf

// SoulV2ParticipantSpec preserves the additive external-participant contract
// without granting external sources native SOUL authority.
type SoulV2ParticipantSpec struct {
	ID         string
	Source     string
	Role       string
	Stage      string
	Capability string
	State      EpistemicState
}

func SoulV2ParticipantSpecs() []SoulV2ParticipantSpec {
	return []SoulV2ParticipantSpec{
		{ID: "bijux-dag-runtime", Source: "bijux/bijux-core", Role: "deterministic DAG execution and provenance", Stage: "0-12", Capability: "validated-DAG-execution", State: PROJECTED},
		{ID: "ouro-loop", Source: "VictorVVedtion/ouro-loop", Role: "bounded remediation loop", Stage: "7,9", Capability: "verify-remediate-loop", State: PROJECTED},
		{ID: "recurs", Source: "Gen-Verse/Recuris", Role: "experiential failure-localization memory", Stage: "3,11", Capability: "failure-localization", State: PROJECTED},
		{ID: "fedml", Source: "FedML-AI/FedML", Role: "federated learning runtime", Stage: "learning", Capability: "federated-training", State: PROJECTED},
		{ID: "hivemind", Source: "learning-at-home/hivemind", Role: "decentralized peer learning", Stage: "learning", Capability: "decentralized-training", State: PROJECTED},
		{ID: "temporal", Source: "temporalio/temporal", Role: "durable workflow execution", Stage: "orchestration", Capability: "durable-workflow", State: PROJECTED},
		{ID: "hora-graph-core", Source: "Vivien83/hora-graph-core", Role: "HortaCore graph-memory reinforcement", Stage: "memory", Capability: "bi-temporal-memory", State: PROJECTED},
		{ID: "cognitive-workspace", Source: "tao-hpu/cognitive-workspace", Role: "Clareira global workspace reinforcement", Stage: "cognition", Capability: "global-workspace", State: PROJECTED},
		{ID: "ravana", Source: "OpenSource-Syndicate/RAVANA", Role: "proactive autonomous-agent reinforcement", Stage: "cognition", Capability: "proactive-agent", State: PROJECTED},
		{ID: "ray", Source: "ray-project/ray", Role: "distributed compute runtime", Stage: "compute", Capability: "distributed-runtime", State: PROJECTED},
		{ID: "nervo-vago", Source: "SOUL internal unified component", Role: "canonical nervous signal boundary", Stage: "cross-cutting", Capability: "nervo-vago.event@1.0.0", State: PROJECTED},
	}
}

// NewSoulV2Participants materializes the eleven additive contracts against the
// existing GRF boundary participant implementation. No external code is copied.
func NewSoulV2Participants() (map[string]*BoundaryParticipant, error) {
	out := make(map[string]*BoundaryParticipant, len(SoulV2ParticipantSpecs()))
	for _, spec := range SoulV2ParticipantSpecs() {
		p, err := NewBoundaryParticipant(ParticipantDescriptor{
			ID: spec.ID,
			Source: spec.Source,
			Role: spec.Role,
			State: spec.State,
		})
		if err != nil {
			return nil, err
		}
		out[spec.ID] = p
	}
	return out, nil
}
