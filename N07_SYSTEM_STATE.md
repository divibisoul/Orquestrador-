# N07 system state

## Current role
N07 is the SOUL Orchestrator. It coordinates N01–N06 and connects neural processing, prefrontal decision policy and compute execution.

## Cumulative engineering rule
PRESERVE → AUDIT → CORRECT → COMPLETE → CONNECT → CROSS → FUSE → OPTIMIZE → VALIDATE → DOCUMENT → ADVANCE.

A declaration, type, route or commit is not proof of runtime completion. Completion requires executable behavior and evidence.

## Cross-front contract
Every change made in another nucleus must be reviewed for N07 compatibility continuously, not only at a final phase. Relevant dimensions: protocol version, nucleus IDs, correlation/trace propagation, capability ownership, agent/tool contracts, input/output schemas, transport, timeout/cancel semantics, authentication, retries, circuit state and observability.

## N07 peer surface
N07 participates in active bidirectional Mesh integration with N01, N02, N03, N04, N05 and N06. The runtime exposes an inbound canonical Mesh gateway and an outbound peer adapter contract. N07 is no longer reserved as a purely final-stage target; its integration surface advances concurrently with the other nuclei while final commissioning remains evidence-gated.

## Active integration strategy
N07 is developed simultaneously with N01–N06. Each new peer capability, function, tool, agent contract, transport or resilience change must be reconciled into N07 as soon as it is structurally compatible. This prevents the final fusion from becoming a large unverified merge and allows N07 to expose integration defects early.

## Current known boundary
The N07 outbound peer client uses the canonical Mesh contract and HMAC signing. Discovery is aligned to the canonical `mesh.discovery` capability. Full live N01–N07 commissioning remains dependent on both endpoints being concurrently reachable with compatible runtime configuration and secrets.

## Integration gates
Continuously validate:

1. canonical Mesh contract across N01–N07;
2. five-peer bidirectional logical relationships for N01–N06 and six peers around N07;
3. capability discovery and ownership;
4. delegation and response correlation;
5. failure/timeout/fallback behavior;
6. inter-nucleus parallel execution;
7. composed capabilities and function/tool fusion;
8. distributed Super GPU execution;
9. final N01–N06–N07 integrated tests and runtime commissioning.

## Evidence policy
Current repository state, latest relevant commits, executable tests and CI results take precedence over prior conversation claims. When CI execution is unavailable, keep the affected item explicitly unverified rather than promoting structural completion to runtime completion.
