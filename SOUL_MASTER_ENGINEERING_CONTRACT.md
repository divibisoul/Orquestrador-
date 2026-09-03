# SOUL — Master Engineering Contract

## Authority
This document is the executable-governance source that consolidates the SOUL engineering directives. It does not replace runtime code; CI and the Architectural Integrity Engine enforce its invariants.

## Absolute rules
1. No mocks, fakes, fabricated success, simulated hardware, or unverified connectivity presented as production behavior.
2. Seven nuclei only: N01..N07. Each nucleus keeps native ownership of its runtime, agents, tools, and capabilities.
3. Exactly one canonical authority per responsibility. Legacy names may remain only as deprecated compatibility adapters.
4. Canonical Mesh protocol: `soul-mesh/1`, contract `1.1.0`.
5. Structural fusion topology is linear and adjacent: N01↔N02↔N03↔N04↔N05↔N06↔N07. Each edge is bidirectional (2 directional channels).
6. Dynamic fusion must be capability-mediated, provenance-aware, executable, and reject non-adjacent direct fusion.
7. N06 canonical execution authority is `N06CapabilityEngine`; historical processors/runtimes must not own independent registries.
8. N07 is the governance, orchestration, federation, and SuperGPU control plane. N01..N06 remain independent runtimes.
9. `PrefrontalNeocortex` is the executive admission boundary for risk-bearing dispatch and must be connected to a real neural signal provider.
10. SuperGPU federation may grant real compute leases to N01..N07, but physical hardware must never be fabricated. Detectable-but-unsupported GPU hardware is `DEGRADED`.
11. CI must fail closed. `continue-on-error`, `|| true`, silent lockfile mutation, or artificial success are forbidden.
12. Every failure is a system defect until root-caused. The response is detect → isolate → reproduce → correct → harden → validate → document.
13. Every claim of readiness requires evidence from real source, tests, CI, and runtime where applicable.
14. `PASS`, `FAIL`, and `DEGRADED` must remain explicit, evidence-backed states.
15. `sourceRef` must identify the exact commit represented by the manifest. Drift is a governance failure, not documentation noise.

## Required validation layers
- source/HEAD integrity
- protocol and contract integrity
- authority uniqueness
- topology symmetry and adjacency
- dependency/lockfile determinism
- typecheck/build/test/race where applicable
- cross-nucleus contract tests
- real E2E when runtime endpoints exist
- hardware/backend reality checks
- immutable diagnostics for infrastructure failures

## Failure handling
Infrastructure failures such as runner provisioning failures, missing workflow logs, and `BlobNotFound` must never be converted into application-code failures without evidence. They require their own forensic classification and diagnostic artifact.

## Non-destructive convergence
Before integrating any branch or PR, compare its base, head, changed files, ancestry, tests, CI evidence, and responsibilities against the current canonical `main`. Complementary work is merged; duplicate authority is consolidated; unsafe or unverified work remains outside the canonical line until corrected.
