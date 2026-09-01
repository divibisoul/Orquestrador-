# N07 — 40 Function Execution Audit

This document is an evidence map, not a substitute for executable code. A function is considered implemented only when its code path exists and tests exercise it.

## Orchestrator
1 New — dependency validation and runtime construction.
2 Register — semantic version validation and duplicate/version isolation.
3 Route — exact-version and highest-semver resolution with TTL cache.
4 Submit — context, deadline, rate, breaker, trace, metrics and cleanup.
5 Execute — message construction plus schema gate.
6 Cancel — active trace cancellation.
7 Status — runtime state.
8 Health — component and circuit health.
9 Stats — runtime/telemetry aggregation.
10 Shutdown — active cancellation and compute shutdown.

## Neural
11 New — bounded learning-rate and runtime initialization.
12 AddEdge — bounds, self-cycle and graph-cycle validation.
13 RemoveEdge — safe deletion and cache invalidation.
14 Activate — finite-value validation and exact batch-multiple processing.
15 Forward — context-aware propagation and bounded cache.
16 Learn — SGD/RMSProp/Adam parameter updates, regularization and gradient clipping.
17 Normalize — stable z-score normalization with finite-value checks.
18 Attention — scaled attention with finite-value checks and configured head weighting.
19 Backprop — derivative calculation and gradient clipping.
20 Health — topology, learning, cache and gradient metrics.

## Prefrontal
21 New — threshold/capacity/policy initialization.
22 Evaluate — weighted multicriteria decision scoring.
23 Plan — Pareto filtering plus score ordering.
24 Prioritize — score/urgency ordering.
25 Inhibit — risk/utility policy gate.
26 Select — policy-aware candidate selection.
27 ValidateAction — complete policy validation before commit.
28 Commit — decision record with justification and pending outcome.
29 Recall — bounded decision history retrieval.
30 Health — evaluation, inhibition and latency metrics.

## SuperGPU
31 New — backend/resource runtime initialization.
32 Discover — cached hardware/backend discovery and capabilities.
33 Select — preference-aware available-device selection.
34 Reserve — exclusive reservation with expiry.
35 Release — owner-checked resource release.
36 Execute — context-aware backend execution with capability enforcement.
37 Batch — ordered bounded batch execution.
38 MemoryStats — explicit backend memory-support state.
39 Health — device/reservation/discovery health.
40 Shutdown — stop admission, clear reservations and drain executions.

## Backend and storage integration

- Unified HTTP backend exposes health, capabilities, execution, intent and storage endpoints.
- Supabase persistence is server-side only and uses `service_role`; client roles cannot directly access N07 run/artifact tables.
- Current Storacha path uses the Guppy client with a configured Space and persistent `--data-dir`; legacy Web3.Storage-compatible support remains explicit compatibility behavior.
- Upload size limits, CID validation, cancellation handling and HTTP error propagation are tested.
- Production container build is validated in CI; container publication to GHCR is automated on main/tag pushes.

## Cross-cutting verification

- Mesh protocol recognizes N01..N07.
- Neural bridges exist for N01..N06 and target N07.
- N07 exposes identity and bidirectional topology metadata.
- N07 provides canonical Mesh ingress plus direct control-plane endpoints.
- Federated parallel execution has parent/child correlation and required-task semantics.
- Current-head CI is an acceptance gate; live seven-runtime E2E remains separate and requires concurrently deployed runtimes.
- Legacy artifacts are removed only after dependency analysis proves they are outside the active execution graph.

## Release boundary

N07 source, backend, Supabase persistence, storage adapters and production packaging are structurally implemented and automated-tested. Final ONLINE status still requires actual deployed runtime endpoints, configured secrets/Space credentials and a real non-ping seven-nucleus capability transaction.
