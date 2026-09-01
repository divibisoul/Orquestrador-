# N07 — LIVE SYSTEM STATE

Updated: 2026-09-01
Branch: `upgrade/n07-hybrid-supercompute-v2`
PR: #15

## Current state

N07 is an active parallel engineering front. It is not reserved for the end of the program. It remains an independent AI/runtime and a cooperative orchestration/fusion layer for N01–N06.

## Implemented in this branch

- canonical Soul Mesh envelope validation and HMAC boundary;
- Mesh ingress at `/api/soul-mesh`;
- six N01–N06 structural peer relationships;
- live capability discovery through `mesh.discovery` with legacy `mesh.describe` compatibility;
- capability-aware candidate routing using health, observed success, latency and in-flight load;
- bounded retry/backoff and circuit protection;
- local-first execution followed by federated delegation;
- arbitrary JSON-like peer payload preservation;
- explicit topology contract through `SOULTopology()` and `/topology`;
- federated parallel execution with per-task child trace IDs and timings;
- dependency-aware capability composition with explicit component provenance;
- federation statistics and Prometheus-compatible metrics;
- runnable N07 control plane;
- regression tests for routing, retry, HMAC, trace correlation, arbitrary payloads and parallel federation.

## Cross-front evidence used

The N01 engineering coordination contract requires GitHub state verification, corrective action on failures, Mesh reuse, capability/function/tool fusion, five-in/five-out peer links, distributed SuperCompute and evidence-based closure. N07 implements these as an active cooperative layer rather than waiting for final commissioning.

Observed active work in the other nucleus repositories includes hybrid Mesh, communication-fabric, capability-fusion, N02↔N06 linking, N01↔N06 fusion and SuperGPU-oriented branches. N07 therefore avoids hardcoded ownership of their internals and consumes capabilities through the canonical Mesh contract.

## Verification state

Source-level and package-level changes are present in the branch and are covered by executable tests. GitHub Actions must still be checked against this exact branch head before any operational readiness claim.

## Known boundaries

- A real six-runtime distributed E2E depends on N01–N06 endpoints being simultaneously configured and available; source presence is not treated as runtime proof.
- Transport classes are represented by an adapter contract. Only transports with an actual implementation and endpoint should be counted as operational.
- Native CUDA/HIP execution remains provider/backend dependent and is not claimed merely from device discovery.

## Handoff

WHAT_CHANGED: N07 federation, hybrid execution boundary, topology, composition, parallel distributed execution and telemetry.

WHAT_WAS_FOUND: previous discovery vocabulary mismatch; local-route false negative risk; duplicate federation stats implementation; branch ancestry divergence between earlier N07 work streams.

WHAT_REMAINS: exact-head CI green evidence; live six-nucleus E2E; integration of concrete peer endpoints/capabilities as their fronts stabilize; final N01×N06×N07 function/tool/capability fusion.

WHAT_NEXT_AGENT_SHOULD_DO: re-read this file, inspect the current HEAD and active peer-front changes, then correct the next verified gap without overwriting concurrent work.
