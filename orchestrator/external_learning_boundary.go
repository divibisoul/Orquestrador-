package orchestrator

// ExternalLearningBoundary describes the canonical ownership boundaries for
// FedML and Hivemind without embedding or duplicating either upstream.
type ExternalLearningBoundary struct {
	FederatedProvider       string
	DecentralizedDependency string
	ControlPlane            string
	PersistenceBoundary     string
	StrategyBoundary        string
	TransportBoundary       string
}

func DefaultExternalLearningBoundary() ExternalLearningBoundary {
	return ExternalLearningBoundary{
		FederatedProvider:       "FedML",
		DecentralizedDependency: "Hivemind",
		ControlPlane:            "N07",
		PersistenceBoundary:     "N01.HortaCore",
		StrategyBoundary:        "N07.Prefrontal",
		TransportBoundary:       "SARA.NervoVago",
	}
}
