# SOUL Cross-Front — Orquestrador / Compute

## Authority
GitHub `main` is the authoritative repository state. Parallel ChatGPT fronts are independent workers and may modify the six nuclei concurrently.

## Current compute state
- Component: Transcendental Compute Engine (TCE)
- Package: `compute/transcendental`
- Isolation: preserved; no dependency on Prefrontal, Neural Fabric, router, cortex or nucleus repositories.
- Feature flag default: disabled.
- Runtime integration: not started because the target runtime/router/cortex files are not present in this repository.

## Current fixes applied
- Removed duplicate `models/catalog.go` implementation so the canonical catalog lives in `models/model.go` and remains config-aware.
- Fixed selector scoring type error exposed by CI.
- Enforced explicit selector strategies and validation.
- Enforced Trillium preference for small workloads under automatic selection.
- Preserved per-workload execution errors in `ExecutePlan`.
- Added nil/zero-value executor protection.
- Expanded selector and executor edge-case tests.

## Current verification
- GitHub Actions runs `vet`, isolated tests, race test, full build and vulnerability scan.
- The latest run is currently executing against the current `main` head. No green result is claimed until all required jobs finish successfully.

## Cross-front consumption
Other fronts must not recreate or replace the TCE catalog/selector/executor merely because their conversation has an older snapshot. Read this file and the current tree before changing the component.

## Next work
1. Finish CI validation.
2. Expand deterministic model/metrics assertions and model reference metadata.
3. Add an integration-neutral capability descriptor for the Orquestrador without importing external nuclei.
4. Only then connect optional adapters to the real Prefrontal/Neural Fabric files once those files exist in this repository.
5. Preserve the disabled-by-default path and interface-only coupling.

## Design principle
TCE selects and simulates compute; it does not become the cognitive orchestrator. The future Orquestrador must compose TCE + agents + tools + nucleus capabilities rather than replacing nucleus ownership.
