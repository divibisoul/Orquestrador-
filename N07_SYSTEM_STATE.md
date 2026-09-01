# N07 — LIVE SYSTEM STATE

Date: 2026-09-01
Branch: `fusion/n07-complete-live-v3`

## Current role

N07 is the SOUL Orchestrator, an independent AI/runtime and an active parallel engineering front. It coordinates N01–N06 and connects neural processing, prefrontal policy, compute, Mesh federation, capability composition and distributed execution.

## Cumulative engineering rule

PRESERVE → AUDIT → CORRECT → COMPLETE → CONNECT → CROSS → FUSE → OPTIMIZE → VALIDATE → DOCUMENT → ADVANCE.

A declaration, type, route or commit is not proof of runtime completion. Completion requires executable behavior and evidence.

## Implemented in the current line

- canonical Soul Mesh envelope `1.1.0`;
- source/target validation, correlation identity, nonce/replay protection, clock-skew validation and TTL enforcement;
- HMAC authentication on peer transport;
- N01–N06 structural peer relationships;
- canonical `mesh.discovery` with `mesh.describe` compatibility fallback;
- peer health/latency/failure-aware routing;
- bounded retry/backoff and circuit protection;
- bounded request/response payloads;
- local-first execution with federated fallback;
- explicit N07 agent identities for discovery, routing, execution, composition, validation and observation;
- explicit N07↔N01..N06 topology contract;
- federated parallel execution with per-task child correlation and timing;
- capability composition/fusion path;
- federation statistics and Prometheus-compatible runtime metrics;
- bounded TTL discovery cache with invalidation after peer failure;
- runnable control-plane endpoints and regression coverage.

## Cross-front synchronization

N07 explicitly consumes the verified N01/N06 fusion ownership model instead of copying their runtime implementations. Active N02/N03/N04/N05 fronts are treated as evolving peers; N07 is designed to discover their live capabilities through the canonical Mesh and adapt to contract changes rather than assuming static ownership.

## Known integration boundaries

- Full six-runtime E2E requires N01–N06 endpoints to be simultaneously configured, reachable and authenticated with compatible contract versions.
- Real accelerator execution remains provider/backend dependent; device discovery alone is not reported as executable CUDA/HIP.
- GitHub Actions evidence is the authoritative operational gate. A missing workflow result remains unverified.

## Closure gates

1. Exact current-head CI build, vet, test and race evidence.
2. Live N01–N06 peer configuration and multi-runtime E2E.
3. Measured SuperGPU/fusion concurrency, latency and failure-recovery behavior.
4. Final N01×N06×N07 fusion after upstream capability/tool contracts stabilize.

## Handoff

WHAT_CHANGED: protocol hardening, active federation, peer discovery/routing/delegation, agent identity, topology, composition, parallel SuperGPU, caching and observability.

WHAT_WAS_FOUND: stale discovery operation naming, missing gateway methods, missing TTL enforcement, branch/commit concurrency hazards and gaps between documented and executable surfaces.

WHAT_REMAINS: exact-head CI proof, live multi-runtime E2E, stabilization of external capability/provider contracts and final N01×N06×N07 fusion closure.

WHAT_NEXT_AGENT_SHOULD_DO: verify the current branch HEAD and all active peer-front changes before editing; preserve the canonical Mesh; fix the next evidence-backed defect and repeat CI validation.
