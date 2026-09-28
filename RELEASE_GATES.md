# SOUL Release Gates

This is the N07 release checklist. A gate is PASS only when its evidence is reproducible from source, CI, an E2E run, or a live health/execution probe. Missing runtime configuration is BLOCKED; it is never converted into simulated success.

| Gate | Status | Required evidence | Current boundary |
|---|---|---|---|
| G0 — repository inventory | PASS | Current GitHub source, workflows and package metadata | Completed 2026-09-28. |
| G0.1 — canonical Mesh contract | PASS | Source + Mesh tests | `soul-mesh/1`, contract `1.1.0` retained. |
| G0.2 — reproducible CI | PARTIAL | Current CI runs + lockfile inventory | N03 dependency-resolution/dependency-review failed and exposed no persisted job-step logs; no blind dependency rewrite was made. N02 lacks a committed lockfile in the current inventory. |
| G1 — N01 native E2E through N07/N01 | NOT MEASURED | Live N01 endpoint + native capability + correlation | No N01 live URL is configured for this N07 gate. |
| G1 — N04 native E2E | BLOCKED | Real integration transaction + correlation/HMAC | Requires `SOUL_MESH_N04_URL` and `SOUL_MESH_HMAC_SECRET`. |
| G1 — N05 native E2E | BLOCKED | Real integration transaction + correlation/HMAC | Requires `SOUL_MESH_N05_URL` and `SOUL_MESH_HMAC_SECRET`. |
| G1 — N06 native E2E | BLOCKED | Real integration transaction + correlation/HMAC | Requires `SOUL_MESH_N06_URL` and `SOUL_MESH_HMAC_SECRET`; target is native `support.mesh`. |
| G1 — N07 local runtime | PASS | Go vet/test/race/build | Current N07 validation path is green. |
| G1 — SARA local authority | PASS | SARA CI + self-check | Current SARA CI is green. |
| G1 — external staging | BLOCKED | Required staging credentials + live health/execution/storage | Current commissioning run stops during required-secret validation. |
| G2 — seven-core smoke | NOT OPEN | Real response matrix for all configured/deployable peers | Must wait for Phase 1 real execution evidence. |
| G2 — fail closed | PASS | Missing/unreachable peer yields explicit failure | N07 PeerClient rejects unconfigured peers and opens a circuit after repeated failures. |
| G2 — retry/timeout/circuit exercised | NOT MEASURED | Runtime/E2E evidence under failure | Implementation exists; real federation gate has not been commissioned. |
| G3 — operational handoff | NOT OPEN | Task crosses >=2 nuclei + SARA with one correlation | No live evidence yet. |
| G4 — planner DAG | NOT OPEN | Real decomposition + real execution | Must wait for G1/G2. |
| G5 — learning loop | NOT OPEN | >=3 historical failures converted to permanent tests/capabilities | Requires preserved failure evidence. |
| G6 — Super AGI gates | NOT OPEN | World state + meta-monitoring + governance + proof matrix | No operational claim before prior gates. |

## Hard release conditions

1. No fully-ONLINE declaration while a required external gate is BLOCKED or NOT MEASURED.
2. No capability is operational from declaration alone; an executable transaction must succeed with correlation preserved.
3. Integration tests that need live peers fail closed on missing configuration.
4. Default/offline CI remains independent of external secrets and preserves legacy behavior.
5. No later phase is opened by documentation alone.

## Current release decision

**PHASE 1 PARTIAL — NOT FULLY ONLINE.**

The repository family already contains real Mesh protocol handling, authentication, discovery, routing, retry/circuit logic and local runtime validation. The current work removes a synthetic federation proof from the N07 test path and replaces it with a real, opt-in integration gate. External peer URLs/secrets and the unresolved N03 dependency-review failure remain explicit blockers.
