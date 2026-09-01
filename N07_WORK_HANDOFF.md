# N07 concurrent engineering handoff

## Permanent contract
PRESERVE → AUDIT → CORRECT → COMPLETE → CONNECT → CROSS → FUSE → OPTIMIZE → VALIDATE → DOCUMENT → ADVANCE.

GitHub `main` is the source of truth. Multiple fronts can change N07 concurrently. Before writing, reread current file SHA; on SHA conflict, reread and integrate instead of overwriting.

## Current implementation focus
N07 remains the single orchestration runtime. It owns routing, federation, neural access, prefrontal policy and compute coordination; N01–N06 retain their native ownership.

## Current neural fabric
N01–N06 have N07 bridge contracts for `neural.forward@1.0.0` and `neural.learn@1.0.0`, using Soul Mesh `1.1.0`, correlation identity, bounded numeric payloads, deadlines and HMAC where configured.

## Current N07 surfaces
- canonical Mesh ingress: `/api/soul-mesh`
- health: `/health`
- status: `/status`
- metrics: `/metrics`
- identity: `/identity`
- topology: `/topology`
- direct execution: `/execute`

## Current known convergence work
1. Keep one canonical N07 gateway path; do not create parallel gateway implementations with overlapping authority.
2. Keep one canonical neural runtime; bridges are adapters only.
3. Reconcile concurrent federation/SuperGPU branches into `main` using exact-head validation.
4. Ensure versioned operation routing, response contracts and HMAC canonicalization match across all bridges.
5. Validate local and federated execution with deterministic unit/race tests before claiming runtime closure.

## Other-front coordination
N01–N06 each contain `N07_NEURAL_FABRIC_HANDOFF.md`. Those files are the communication contract for the concurrent fronts. Changes in one nucleus that affect protocol, capabilities, tools, agents, transport, timeout, security or output shape must be reflected in N07 before the integration is considered coherent.

## Evidence rule
Green CI applies only to the exact tested SHA. A branch being ahead/behind another branch is not evidence of integration. Live six-runtime E2E remains separately gated on configured reachable endpoints.
