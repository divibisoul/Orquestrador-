# SOUL v1 — Release Closure Gates

## Objective
Deliver the current SOUL architecture as an operational, verified v1. This document is a finite release gate, not an invitation to add indefinite new features.

## Target
Primary operational target: **2026-09-07**.
Contingency window for release-blocking defects only: **2026-09-08 through 2026-09-10**.

## Mandatory gates

1. N01–N06 current main branches contain the intended cumulative nucleus, Mesh and N07 neural integration changes.
2. N07 is the canonical orchestration/fusion layer; no duplicate peer runtime is introduced.
3. Mesh contracts are compatible across all active peers, including identity, correlation, authentication and payload semantics.
4. N07 discovery, capability routing, delegation, composition and federated parallel execution are executable and covered by regression tests.
5. Exact-head CI is green for the release revisions: format, vet/typecheck, unit/integration tests, race checks and build.
6. Security checks pass without secrets committed to source.
7. Live bidirectional commissioning proves N07 can communicate with N01, N02, N03, N04, N05 and N06 using their configured endpoints and credentials.
8. Neural federation proves forward/parallel paths with bounded deadlines, finite payload validation and propagated correlation.
9. SuperCompute/SuperGPU proves parallel fan-out, per-task timing, aggregation and partial-failure behavior.
10. Final smoke test proves startup, health, Mesh ingress, discovery and a representative delegated/fused execution path.

## Non-blocking future work

Hardware-specific CUDA/HIP/NPU providers, additional transports, new capabilities, performance research and v2 features are outside the v1 closure gate unless they are required for a failing mandatory gate.

## Completion states

- `INCOMPLETE`: one or more mandatory gates are not implemented.
- `STRUCTURALLY READY`: implementation and local regression coverage exist, but live/E2E evidence is absent.
- `RELEASE CANDIDATE`: all mandatory code/CI gates are green; live commissioning is the only remaining evidence.
- `READY / ONLINE`: all mandatory gates and runtime smoke tests pass.

## Evidence rule
Never promote a state based solely on documentation, declarations, branch names or commit messages. Evidence must come from the exact revision, executable tests/CI and live runtime checks where required.
