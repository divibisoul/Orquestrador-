# SOUL Cognitive Capabilities R1

This layer is additive over the current N07 federation and does not replace N01-N07 or SARA.

## Implemented structurally

- Goal contract with objective, success criteria, capabilities, correlation and expiry.
- Discovery-first planner: a capability must be executable in peer discovery before a planned step is accepted.
- Mesh executor: delegates through the existing N07 PeerClient and dynamic routing; no second Mesh.
- Observer: records peer, latency, correlation and deterministic error.
- Working memory: bounded item count, TTL and relevance eviction.
- Goal store: bounded by goal expiry.
- Local critique: reuses the existing N07 Prefrontal Cortex policy gate.
- SARA policy gate: irreversible goals require SARA audit when policy enforcement is enabled; missing SARA is fail-closed.
- Long-term observation persistence: reuses the existing Supabase run store; no second database.
- Feature flag: N07_COGNITIVE_LOOP_ENABLED=false by default.
- Main-process registration: wired into cmd/nexus/main.go behind the explicit feature flag; disabled-by-default remains fail-closed.

## Authority

N07 remains orchestration/execution authority. Peer nuclei remain owners of their native capabilities. SARA remains the authority for regeneration and the regenerative policy path.

## Not yet claimed

- Parallel execution of independent planned cognitive steps is not yet enabled by this layer; Octacore already owns bounded parallel waves.
- Full input/output JSON Schema validation is not yet claimed because current peer discovery exposes executable capability names but does not consistently publish schemas.
- Long-term memory read/recall is available at the Supabase store layer, but the cognitive loop has not yet promoted it to an automatic context-recall policy.
- No E2E seven-nucleus cognitive transaction is claimed.

## State

STRUCTURAL: code, unit tests and main-process registration exist on this branch.

VALIDATED: requires exact-head CI format/vet/test/race/build evidence.

INTEGRATED: requires live N07 -> peer capability transaction with preserved correlation.

ONLINE: not claimed.


## 2026-09-28 reconciliation

The main runtime registration is present behind the existing feature flag. A CI vet failure caused by a missing `strconv` import was corrected, followed by a planner contract correction that adds the preserved `Target` field to `Step`. Exact-head CI must still re-run and pass before this layer is considered VALIDATED.
