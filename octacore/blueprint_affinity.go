package octacore

import "github.com/divibisoul/Orquestrador-/blueprint"

// ResolveBlueprint exposes the ATLAS functional-affinity layer through the
// existing OctaCore processor. It does not transfer capability ownership.
func (p *Processor) ResolveBlueprint(query string, limit int) ([]blueprint.Match, error) {
	return blueprint.Resolve(query, limit)
}

// ComposeBlueprint builds an additive route plan from existing capabilities.
// It never removes or replaces an already registered nucleus capability.
func (p *Processor) ComposeBlueprint(query string, limit int) (blueprint.RoutePlan, error) {
	return blueprint.Compose(query, limit)
}

// BlueprintManifest returns a defensive copy of the current functional map.
func (p *Processor) BlueprintManifest() blueprint.Manifest {
	return blueprint.DefaultManifest()
}
