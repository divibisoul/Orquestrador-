# Architecture — Orquestrador-Nexus Recovery

## Canonical domain

The recovery branch has one canonical source of truth for each domain plane:

- `core/orchestrator` — canonical workflow engine. Its public `Engine` exposes workflow lifecycle plus execution, distributed dispatch, circuit breaker, bulkhead, rate limiting and fractal orchestration capabilities; specialized mechanics live in its `execution`, `workflow` and `fractal` subpackages.
- `core/prefrontal` — canonical executive decision layer.
- `core/neuralfabric` — canonical route scoring, prediction and feedback layer.
- `core/superagi` — canonical provider-neutral generation, memory and verification boundary.
- `compute` — device/job execution abstraction.
- `mesh` — canonical node discovery, registration, heartbeat and stale tracking runtime.
- `state` — canonical versioned state and CAS runtime.
- `security` — policy/auth/audit boundary.
- `api/proto` — provider-neutral inter-core contracts.

The former `internal/nexus`, `internal/mesh` and `internal/state` implementations were removed from the recovery runtime only after consumer-graph checks. Their historical source remains available through Git history and the untouched foundation branch.

## Six-plane flow

```text
                +---------------------------+
                | API / Proto / Control     |
                +-------------+-------------+
                              |
                              v
                    +-------------------+
                    |   Orchestrator    |
                    | core/orchestrator |
                    +---------+---------+
                              |
              +---------------+---------------+
              |                               |
              v                               v
      +---------------+                +---------------+
      |  Neo-Cortex   |                | Neural Fabric |
      | core/prefrontal|               | core/neuralfabric|
      +-------+-------+                +-------+-------+
              |                                |
              +---------------+----------------+
                              |
                              v
                +---------------------------+
                | Super AGI / Compute       |
                +-------------+-------------+
                              |
                +-------------+-------------+
                |                             |
                v                             v
          +-----------+                 +-----------+
          |   Mesh    |                 |   State   |
          | discovery |                 | versioned |
          +-----------+                 +-----------+

Security is an enforcement boundary around external effects.
Observability surrounds all six planes without owning domain decisions.
```

## Orchestrator implementation model

`core/orchestrator/runtime.go` is the public engine facade. It delegates specialized behavior to:

```text
Engine
 ├─ workflow lifecycle / checkpoints
 ├─ execution.Executor
 │   ├─ parallel execution
 │   ├─ distributed dispatch
 │   ├─ circuit breaker
 │   ├─ bulkhead
 │   └─ token-bucket rate limiting
 └─ fractal.Manager
     ├─ spawn sub-orchestrator
     ├─ kill sub-orchestrator
     └─ deterministic rebalance
```

This keeps one public Orchestrator engine while preventing resilience and fractal mechanics from becoming duplicate domain implementations.

## Honest capability boundaries

A source-level API is not proof of a production backend. The following remain extension layers: real distributed Raft, live generated gRPC runtime, mTLS/SPIFFE, durable audit, physical GPU/NPU execution and concrete external model runtimes.

Neural Fabric persistence/adaptive learning remains incomplete. Super AGI training, LoRA and replay boundaries explicitly return `not implemented` until real backends exist.

## Communication

`api/proto/*.proto` defines the provider-neutral contract for future inter-process communication. Generated bindings and live gRPC serving can be added without changing the canonical decision/execution domain.

## Self-hosted CI

Recovery validation, security scans and benchmarks target the repository's self-hosted Actions runner. See `docs/SELF_HOSTED_RUNNER.md`.

## Acceptance rule

No merge to `main` is permitted until current recovery-head CI and security runs pass and the six-plane E2E test succeeds. Performance targets are benchmark claims only when accompanied by reproducible measurements.
