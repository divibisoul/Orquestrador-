# N07 system state

## Current phase
PHASE 1 — MESH ONLINE MINIMUM (PARTIAL)

## Ownership
N07 owns orchestration, neural federation, routing, capability composition and the SuperGPU control plane. N01–N06 remain independent runtimes. SARA remains the independent Python regenerative authority.

## Phase 0 — evidence reconciliation
| Unit | Status | Current evidence |
|---|---|---|
| N01 | REAL | Latest main validation/build workflows completed successfully on 2026-09-28 at head `ed3eebdd590155244212dc725bc0f6812754b3d3`. |
| N02 | REAL | Latest validation/forensics workflows completed successfully on 2026-09-27 at head `7d977635b376cfd72d600d056ee8bb4123cd15cc`. |
| N03 | INCOMPLETO | Main currently has `package-lock.json`, but the latest dependency-resolution and dependency-review runs failed. GitHub did not expose persisted job-step logs for those failed runs, so no root cause is asserted and no blind dependency rewrite was made. |
| N04 | REAL | Latest N04 CI and Soul Mesh checks completed successfully on 2026-09-27 at head `8c457291181b3295b50615bf3abe01599fe0226e`. |
| N05 | REAL | Latest N05 Mesh/bridge checks completed successfully on 2026-09-27 at head `2c934084e96e1831e845848cced498dc69f853d6`. |
| N06 | REAL | Latest Mesh/authority/validation checks completed successfully on 2026-09-27 at head `1e29f630f47572945c08056c666a1b360d037972`. |
| N07 | REAL / BLOCKED-ENV | Go vet/test/race/build and the local runtime checks pass on the current release line; externally configured staging fails closed when required secrets are absent. |
| SARA | REAL | Latest SARA CI and validation completed successfully on 2026-09-27 at head `4ae3e66e7cf7907cd6008e1f7e34e9919f193f9e`. |

## Critical forensic finding
The previous `mesh/n01_n07_federation_e2e_test.go` used `httptest.NewServer` peers and fabricated `e2e.n04`, `e2e.n05`, and `e2e.n06` capabilities. That was synthetic evidence and could not prove real federation.

That test is now integration-only (`//go:build integration`) and no longer starts peer doubles. It requires real N04/N05/N06 URLs plus `SOUL_MESH_HMAC_SECRET`, discovers executable native capabilities, invokes them through the real N07 PeerClient, and verifies correlation/source/target/payload. The default `go test ./...` path therefore remains deterministic while the real gate fails closed when runtime configuration is missing.

## Phase 1 evidence
| Gate | Status | Evidence boundary |
|---|---|---|
| Canonical protocol | REAL | N07 and peer code use `soul-mesh/1` and contract `1.1.0`. |
| Correlation/HMAC | REAL | N07 PeerClient signs outbound requests and verifies response contract, correlation and HMAC. |
| Real health/discovery surfaces | REAL | N01–N06 expose Mesh discovery/health surfaces in source; external reachability remains environment-dependent. |
| Real N04 native execution | BLOCKED / NOT MEASURED | Integration gate targets native `core.health`; requires live `SOUL_MESH_N04_URL`. |
| Real N05 native execution | BLOCKED / NOT MEASURED | Integration gate targets native `core.health`; requires live `SOUL_MESH_N05_URL`. |
| Real N06 native execution | BLOCKED / NOT MEASURED | Integration gate targets native `support.mesh`; requires live `SOUL_MESH_N06_URL`. |
| Real N07 execution | REAL | N07 CI validates engine/runtime, including vet/test/race/build. |
| Real N01 native execution through N07 | NOT MEASURED | No N01 live URL is present in N07 environment configuration; no endpoint is inferred. |

## Phase 2
NOT STARTED. Circuit breaker/retry/timeout code exists and has unit/runtime coverage, but the seven-core real smoke gate is not opened until Phase 1 external execution evidence exists.

## Phase 3
NOT STARTED. Shared context and N07↔SARA handoff remain bounded by their existing contracts. No claim of cross-runtime operational handoff is made without live evidence.

## Phase 4
NOT STARTED. No planner/DAG feature is activated before the Phase 1 real Mesh gate closes.

## Phase 5
NOT STARTED. Failure→Finding→correction→test remains the governing workflow. Historical failures may be converted into permanent tests/capabilities only after their concrete failure evidence is preserved.

## Phase 6
NOT STARTED. No Super AGI gate is marked operational.

## Release rule
SOUL is **NOT ONLINE as a fully commissioned external federation** until the real integration gate has executed successfully against configured live peers. Missing URLs/secrets are reported as BLOCKED, never simulated as green.

## Non-destructive guarantees
- N01–N07 remain independent repositories.
- SARA remains independent and authoritative for regeneration/governance contracts already present.
- No native capability is removed, renamed or replaced.
- New federation execution evidence is opt-in via the `integration` build tag.
- Existing default `go test ./...` remains the baseline regression gate.

## Internal systems must rise with the Mesh

The Mesh gate measures federation edges. It does not replace or subsume the internal systems of a nucleus.

Every later phase must carry forward the preserved internal substrate of each participating nucleus. In particular, Clareira is an active N01 subsystem and must remain executable, observable and compatible as the federation advances. Its internal processing graph, homeostasis, channels and event path are protected from simplification or replacement.

The system-design inventory also contains 72 nódulos as protected scope. This document does not invent a node-by-node runtime status where source evidence is absent.

See SOUL_EVOLUTION_PRESERVATION.md for the change classification: PRESERVE / ADAPT / EXTEND / RETIRE-WITH-REPLACEMENT.
