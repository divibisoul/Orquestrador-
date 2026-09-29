# Front 1 ↔ Front 2 — Canonical Mesh Contract

## Purpose

Front 1 owns the real communication substrate and its evidence. Front 2 owns the additive cognitive layer. The two fronts must not create divergent communication mechanisms.

## Mandatory rule

Every future cognitive Act must execute through the same canonical N07 PeerClient / Mesh discovery / HMAC / correlation path proven by Phase 1.

The cognitive layer may add Goal, Plan, Observe, bounded working memory and policy logic, but it must not:

- create a second Mesh;
- create a second peer transport;
- call a peer by a private HTTP path that bypasses Mesh authentication/discovery;
- replace a nucleus-native capability with a duplicate implementation in N07;
- infer a peer is ONLINE from source declarations alone.

## Evidence boundary

PR #42 currently contains an additive MeshExecutor backed by *mesh.PeerClient and a discovery-first planner. Its documentation states that main-process registration and live seven-nucleus E2E are not yet claimed. Therefore Front 1 records the compatibility contract but does not merge or activate Front 2 here.

Before cognitive activation, the exact integrated head must pass CI and then a real peer transaction must preserve correlation, source, target, capability and response evidence.

## Preservation rule

Cognitive evolution must raise the integration level without lowering the internal capability of N01–N07 or SARA. Clareira, native agents, tools, registries, capabilities, processing nodes and the broader 72-node protected scope remain owned by their native authorities.

## Gate

No cognitive feature changes the Phase-1 communication state by documentation alone. Only reproducible runtime evidence changes the communication matrix.
