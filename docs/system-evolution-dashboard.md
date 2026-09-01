# SOUL Orchestrator — System Evolution Dashboard

## Purpose
Persistent engineering dashboard for the six concurrent development fronts. GitHub state is authoritative; prose is never used as proof of completion.

## Operational rule
Before every change: inspect current branch/PR/commit state, compare concurrent work, identify the canonical owner of the area, and modify only the necessary owner branch. If a defect, conflict, stub, inactive function, broken contract, or incomplete area is found, fix it before returning to the original task. After every write: reread the file, record the commit, and validate with CI/security when available.

## Anti-loop rule
Never repeat the same action without new evidence. New evidence is a commit/branch change, test result, security finding, integration failure, uncovered gap, or new requirement.

## No-duplication rule
Do not create a second implementation when an equivalent capability already exists. Prefer contract-level integration or a minimal adapter. Preserve distinct responsibilities even when names are similar.

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

## Current work-front discipline

```text
AUDIT -> DIFF -> OWNER -> PATCH -> READ-BACK -> CI/SECURITY -> NEXT GAP
```

Concurrent fronts must be treated as live collaborators. A changed SHA is a synchronization event, not a reason to overwrite another change.

## Milestone history

### Baseline
TCE isolated simulation created with interfaces, deterministic models, selector, executor, configuration and tests.

### Hardening 1
Validation, nil-safety, error propagation, selector policy, configuration limits and regression tests.

### Hardening 2
Cross-front security fixes, Neural Fabric checkpoint/update behavior, Prefrontal control functions, SOUL Fabric readiness semantics.

### Hardening 3
Six-nucleus transport E2E, deterministic Meta-RL seed handling, security workflow refinement and cross-front reconciliation.

### Current objective
Consolidate TCE, Trinity and Fabric by responsibility rather than duplication, then close the remaining end-to-end communication and operational-validation gaps.

## Completion gate
The system is not considered complete from documentation alone. A milestone is green only when the corresponding code exists, tests cover it, and the relevant GitHub validation gate passes.
