# N07 system state

## Current role
N07 is the SOUL Orchestrator. It coordinates N01–N06 and connects neural processing, prefrontal decision policy and compute execution.

## Cumulative engineering rule
PRESERVE → AUDIT → CORRECT → COMPLETE → CONNECT → CROSS → FUSE → OPTIMIZE → VALIDATE → DOCUMENT → ADVANCE.

A declaration, type, route or commit is not proof of runtime completion. Completion requires executable behavior and evidence.

## Cross-front contract
Every change made in another nucleus must be reviewed for N07 compatibility continuously. Relevant dimensions: protocol version, nucleus IDs, correlation/trace propagation, capability ownership, agent/tool contracts, input/output schemas, transport, timeout/cancel semantics, authentication, retries, circuit state and observability.

## Single N07 runtime surface
N07 has one canonical orchestrator runtime. Versioned operations are stored as `name@semver`; explicit versions route exactly, while unversioned calls select the newest registered version. The route cache is keyed by the same canonical operation identity. This prevents versioned registrations from becoming unreachable through the router.

## Neural fabric
N07 exposes the canonical neural execution surface through Mesh. N01–N06 consume `neural.forward@1.0.0` and `neural.learn@1.0.0` through their local adapters while N07 remains the execution owner. The adapters preserve correlation, contract version, authentication and deadlines; they do not duplicate the neural runtime.

## Peer and Mesh surface
N07 participates in bidirectional Mesh integration with N01, N02, N03, N04, N05 and N06. The runtime exposes one inbound canonical Mesh gateway and one outbound peer adapter contract. Discovery and capability delegation use the canonical Mesh path, with HMAC, replay protection, retry/backoff and circuit control.

## Current integration strategy
N07 advances concurrently with N01–N06. New peer capabilities, functions, tools, agents, transports or resilience changes are reconciled into this runtime when structurally compatible. Conflicting implementations are not merged by duplication; ownership and adapter boundaries are preferred.

## Compute boundary
The compute layer remains backend-driven. Hardware detection does not imply executable GPU support. A selected device must be supported by the active backend before execution and resource reservations are released on all exit paths.

## Integration gates
Continuously validate:

1. canonical Mesh contract across N01–N07;
2. five-peer bidirectional relationships for N01–N06 and six peers around N07;
3. capability discovery and ownership;
4. versioned routing and canonical operation identity;
5. neural-fabric invocation from every nucleus adapter;
6. delegation and response correlation;
7. failure, timeout, cancellation and fallback behavior;
8. inter-nucleus parallel execution;
9. composed capabilities and function/tool fusion;
10. distributed Super GPU execution;
11. final N01–N06–N07 integrated runtime commissioning.

## Evidence policy
Current repository state, latest relevant commits, executable tests and CI results take precedence over prior conversation claims. When CI execution is unavailable, keep the affected item explicitly unverified rather than promoting structural completion to runtime completion.
