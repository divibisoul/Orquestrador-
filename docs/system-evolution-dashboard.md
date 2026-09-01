# SOUL Orchestrator — System Evolution Dashboard

## Purpose
Persistent engineering dashboard for all concurrent development fronts. GitHub state is authoritative; prose is never used as proof of completion.

## Mandatory engineering directive
For every GitHub intervention, across every branch/front and every future session:

1. Inspect the real current state first: branch, PR, HEAD SHA, changed files, concurrent work, CI/security state and relevant runtime evidence.
2. Never stop merely because a severe error, blocker or failed test was found. Treat the failure as the next engineering task. Search documentation, upstream sources, issue trackers and alternative implementation strategies when useful, then apply the best safe fix.
3. Every identified defect, error, inconsistency, incomplete function, inactive path, broken contract, missing integration, unsafe behavior or false-success path must be corrected or completed before returning to the original task.
4. Never claim an action was completed from intention, tool invocation or prose. Read the changed file back, verify the resulting SHA/commit and use executable validation whenever possible.
5. Before every write, re-check whether another front changed the same area. A changed SHA is a synchronization event; never overwrite concurrent work blindly. Rebase/merge/consolidate by responsibility and preserve the newest valid implementation.
6. Do not create duplicate implementations. First locate existing equivalents and identify the canonical owner; integrate or adapt instead of cloning functionality.
7. If the preferred implementation path is blocked, immediately pursue a technically valid alternative rather than stopping. A blocker is evidence to change the method, not permission to abandon the requirement.
8. Use time as an engineering signal: track elapsed work, avoid spending the session repeatedly on the same failed action, and move to another independently actionable front while CI or external validation is running.
9. The dashboard/graph must remain live throughout the work. Update it whenever evidence materially changes and before claiming a milestone complete. A graph is a control instrument for scope, time, regressions and remaining work—not decoration.
10. A task is deliverable only when the relevant implementation, tests, integration evidence and operational gates support it. If a gate is pending, the state remains pending even when the code appears correct.

## N07 directive
N07 is a concurrent engineering front and must NOT be frozen until the end. It must continuously receive the same audit, hardening, test and integration treatment as every other subsystem.

The final N07 fusion is deferred until the integration point, but N07 development itself continues now. Its canonical work must be inspected across all active N07 branches before any new implementation is created.

At the fusion stage, N07 must ingest and reconcile the real inputs/outputs, protocols, events, tools and functions exposed by N01–N06 and the orchestrator control plane. Function/tool fusion must prefer shared contracts, capability negotiation and adapters over duplicate implementations. N07 must reach the same production-engineering standards as the other nuclei before final fusion is accepted.

## Current architectural map

```text
                         SOUL ORCHESTRATOR
                                  |
                    +-------------+-------------+
                    |             |             |
                PREfrontal   NEURAL FABRIC   COMPUTE
                    |             |             |
                    +-------------+-------------+
                                  |
                              TRINITY
                                  |
                              SOUL MESH
                                  |
             +--------+--------+--+--+--------+--------+
             |        |        |     |        |        |
            N01      N02      N03   N04      N05      N06
                                  |
                                 N07
                    fusion target: N01 + N06 + all
                    validated cross-system inputs/outputs
```

## Engineering completion matrix

| Area | Evidence-based state | Next verification |
|---|---:|---|
| TCE contracts | 100% | CI + integration |
| TCE validation | 100% | regression |
| TCE models | 90% | model/edge-case coverage |
| TCE selector | 90% | strategy matrix + regression |
| TCE executor | 90% | race/cancellation/limits |
| TCE tests | 90% | CI/race |
| Prefrontal | 90% | integration semantics |
| Neural Fabric | 90% | checkpoint/runtime regression |
| Trinity | 80% | unified contract integration |
| SOUL Fabric | 80% | real endpoint/readiness validation |
| Orchestrator runtime | 80% | E2E + recovery paths |
| Security | 80% | clean gosec + build/vet |
| Observability | 50% | end-to-end trace/metrics |
| TCE ↔ Trinity | 50% | canonical adapter/fusion |
| Trinity ↔ Fabric | 50% | route/feedback integration |
| Prefrontal ↔ Compute | 40% | CostEstimator integration |
| Neural Fabric ↔ Compute | 40% | backend/metrics integration |
| SOUL Mesh ↔ N01–N06 | 30% | real endpoint interoperability |
| Six-nucleus E2E | 30% | live/contract-tested endpoints |
| N07 runtime | active concurrent front | audit all N07 branches + contract/runtime tests |
| N07 fusion | intentionally deferred | cross-nucleus I/O/tool/function contract map |

## Current work-front discipline

```text
TIME/STATE CHECK -> AUDIT -> DIFF -> OWNER -> PATCH/FIX -> READ-BACK -> TEST -> CI/SECURITY -> UPDATE GRAPH -> NEXT GAP
```

Concurrent fronts are live collaborators. A changed SHA is a synchronization event, not a reason to overwrite another change.

## Active N07 branch inventory

```text
feature/n07-orchestrator-real-runtime
upgrade/n07-production-v7
upgrade/n07-final-v1 ... v8
upgrade/n07-prod-v4 ... v6
upgrade/n07-production-v8 ... v18
```

Many N07 branches currently converge on the same commits. Treat the shared commit as evidence already implemented, compare diverging commits before changing anything, and consolidate only the genuinely new behavior.

## Milestone history

### Baseline
TCE isolated simulation created with interfaces, deterministic models, selector, executor, configuration and tests.

### Hardening 1
Validation, nil-safety, error propagation, selector policy, configuration limits and regression tests.

### Hardening 2
Cross-front security fixes, Neural Fabric checkpoint/update behavior, Prefrontal control functions, SOUL Fabric readiness semantics.

### Hardening 3
Six-nucleus transport E2E, deterministic Meta-RL seed handling, security workflow refinement and cross-front reconciliation.

### Hardening 4
Persistent readiness dashboard and operational engineering rules; dashboard percentages tied to executable evidence.

### Current objective
Continue production hardening across all active fronts, keep N07 at the same engineering level in parallel, consolidate TCE/Trinity/Fabric by responsibility, then close remaining end-to-end communication, observability, recovery and activation gaps before final fusion.

## Completion gate
The system is not considered complete from documentation alone. A milestone is green only when the corresponding code exists, tests cover it, the relevant GitHub validation gate passes, the runtime path is exercised where applicable, and the dashboard reflects the new evidence.
