# Evolution and Cross-Front Coordination

This document is the shared state ledger for all concurrent engineering fronts. GitHub state is authoritative; chat history is context only.

## Operating rules

1. Audit the real current state before every mutation: HEAD SHA, branch, PRs, overlapping paths, recent commits, CI/security and relevant runtime evidence.
2. One responsibility has one canonical implementation. Merge or adapt only when contracts and ownership are compatible; never clone an existing capability.
3. Any compiler error, test failure, broken contract, dead path, unsafe behavior, incomplete integration or inconsistent state discovered during audit is an engineering task: correct it at the owning location when safe, then re-read and revalidate it.
4. An error or severe blocker never ends the engineering cycle. Search authoritative documentation, upstream sources, issues and alternative designs when needed, and use a technically valid alternative rather than abandoning the requirement.
5. Every prompt/refinement is additive to the same project. Earlier unfinished work remains active unless explicitly cancelled.
6. Never assume concurrent work is integrated until the target branch actually contains its commit. A changed SHA is a synchronization event; never overwrite newer work blindly.
7. After every write, read the file back, verify the resulting SHA/commit, and run the narrowest useful executable validation before expanding the change.
8. Time is an engineering signal. While an external CI/security run is processing, continue independently actionable work; do not repeatedly waste the same cycle on an unchanged failure.
9. The dashboard and graphs are control instruments, not decoration. Update them when evidence materially changes and before claiming a milestone complete.
10. “Done” requires implementation + test + integration evidence + relevant operational gates. Tool invocation or documentation alone never proves completion.

## Current topology

```text
                         SOUL ORCHESTRATOR
                                |
                 +--------------+--------------+
                 |                             |
          CONTROL / COGNITIVE              COMPUTE
                 |                             |
        +--------+--------+              +-----+------+
        |        |        |              |            |
     Planner   Registry  Health         TCE         Fabric
        |        |        |              |            |
        +--------+--------+              +-----+------+
                 |                             |
                 +-------------+---------------+
                               |
                          SOUL MESH
                               |
                 +------+------+------+------+------+
                 |      |      |      |      |      |
                N01    N02    N03    N04    N05    N06
                               |
                               N07
                     active hardening + future fusion
```

## Parallel-front state

### Nexus consolidation
`feature/nexus-finalization-v1` is the principal consolidation line in open PR #10 against `integration/soul-six-nuclei`. The branch is continuously changing and must be re-read before each write.

Recent verified work includes security hardening, state version monotonicity, mesh route validation, Neural Fabric learned-routing integration, REST metrics exposure and automated regression coverage. The HEAD after the latest dashboard/governance update is `3284f3a14ca9956047b0aedb7bb23496b22c768d`.

### N07
Active N07 work exists on `feature/n07-orchestrator-real-runtime` and multiple `upgrade/n07-*` branches. Many `upgrade/n07-*` branches converge on the same production-history SHA and therefore must not be treated as separate implementations.

The N07 production line currently includes executable neural, prefrontal, protocol, orchestrator and SuperGPU layers plus tests. Recent verified fixes on `upgrade/n07-production-v7` include:

- compute discovery performed during runtime construction so the default path starts with a usable CPU device;
- degraded health reporting when no device is discovered;
- restoration of the N07 format gate;
- an executable blocking Security workflow;
- input-safety regression coverage for neural operations;
- explicit continuous-audit, time, graph and no-duplication rules in the N07 blueprint.

The latest N07 production-v7 work is still a parallel front. It must be hardened now and fused later; “fusion later” never means “leave N07 behind”.

## Verified cross-front correction history

The advanced Nexus branch previously had a real Go type error in `core/superagi/runtime.go`; it was corrected at its owning location. A Security failure caused by a Go-toolchain mismatch was also corrected so Security uses the project's Go 1.25 toolchain.

The latest Nexus validation cycle caught a real REST metrics test defect: `5` was passed as a raw duration (5ns) while the assertion expected milliseconds. The test was corrected to use `5*time.Millisecond`; the resulting current HEAD completed CI and Security successfully.

## Compute reconciliation

The isolated `compute/transcendental` layer remains the deterministic boundary for reference accelerator modeling, cost and metrics. Trinity/Fabric owns orchestration and device abstraction. Reconciliation must produce one canonical compute contract and use adapters only where contracts genuinely overlap.

```text
TCE: reference models / estimation / metrics
              |
              | canonical adapter
              v
      UNIFIED COMPUTE CONTRACT
              ^
              |
Trinity/Fabric: routing / device abstraction / execution
```

## N07-specific fusion contract

N07 must eventually consume and reconcile validated inputs, outputs, events, protocols, capabilities, tools and functions exposed by N01–N06 and the control plane. Before fusion:

```text
N01 ─┐
N02 ─┤
N03 ─┤
N04 ─┼──> capability/contract map ──> N07 fusion layer ──> N01 + N06
N05 ─┤
N06 ─┘
               |
             tools/functions
```

Equivalent functions/tools use a single canonical owner. Complementary capabilities may be composed into new functions when the composition is tested, observable and demonstrably useful. No whole-subtree copy is acceptable when the capability already exists elsewhere.

## Progress dashboard

These are engineering tracking indicators, not claims of end-to-end readiness.

```text
AREA                          STATE
Nexus CI                      GREEN on current verified HEAD
Nexus Security                GREEN on current verified HEAD
Persistent dashboard          ACTIVE
Cross-front audit             ACTIVE
N07 runtime                   ACTIVE / being hardened
N07 format gate               ACTIVE
N07 security gate             ACTIVE
N07 final fusion              DEFERRED, not abandoned
Six-nucleus integration       PARTIALLY PROVEN; real external runtime evidence remains
External tracing              PENDING
Real CUDA/ROCm/NPU            PENDING
Real SPIFFE/workload identity PENDING
Distributed Raft recovery     PENDING
Final activation gate         PENDING
```

## Execution-quality dashboard

```text
GitHub-first auditing          ██████████
Parallel-front inspection      ██████████
Immediate defect correction    ██████████
Alternative-path recovery      ██████████
Post-write verification        ██████████
Duplicate avoidance            ██████████
Prompt-to-prompt continuity    ██████████
Time-aware execution           ██████████
Graph/dashboard continuity    ██████████
Objective CI verification      ██████████
```

A failed gate becomes progress only after the underlying defect is corrected and a replacement validation is inspected. Documentation never upgrades a component to complete.
