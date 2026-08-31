# Architecture — Recovery Baseline

## Canonical direction

The recovery branch treats the `core/*` packages as the canonical domain layer:

- `core/orchestrator` — workflow planning/execution domain.
- `core/prefrontal` — executive decision/policy layer.
- `core/neuralfabric` — route scoring and learning abstractions.
- `core/superagi` — generation/agent capability boundaries.
- `compute` — device/job execution abstraction.
- `mesh` — node discovery/heartbeat runtime.
- `state` — state/runtime abstraction.
- `security` — execution policy boundary.
- `api/proto` — provider-neutral inter-core contracts.

`internal/*` packages are compatibility/integration surfaces until their consumers are migrated. They must not become a second source of truth for domain behavior.

## Target flow

```text
                    +----------------------+
                    |   API / Contracts    |
                    | REST / Proto / gRPC  |
                    +----------+-----------+
                               |
                               v
                    +----------------------+
                    |    Orchestrator      |
                    | core/orchestrator    |
                    +----------+-----------+
                               |
                 +-------------+-------------+
                 |                           |
                 v                           v
        +------------------+        +------------------+
        | Neo-Cortex       |        | Neural Fabric    |
        | core/prefrontal  |        | core/neuralfabric|
        +--------+---------+        +--------+---------+
                 |                           |
                 +-------------+-------------+
                               |
                               v
                    +----------------------+
                    | Compute / Super AGI |
                    +----------+-----------+
                               |
                 +-------------+-------------+
                 |                           |
                 v                           v
            +---------+                 +---------+
            |  Mesh   |                 | State   |
            +---------+                 +---------+

Security policy is an enforcement boundary around externally observable effects.
```

## Duplication policy

No duplicate package is to be deleted merely because it exists. Before removal, all imports and tests must be migrated and the replacement must preserve behavior. `internal/nexus`, `internal/mesh`, and `internal/state` are therefore compatibility candidates, not automatically safe deletion targets.

## Runtime maturity

The repository contains real implementations mixed with explicit boundaries/stubs. A contract, interface, or method existing in source code is not evidence that the underlying distributed/GPU/learning behavior is operational.

## Communication

Protobuf definitions are the canonical inter-core contract when cross-process communication is required. Generated bindings and a live gRPC server are a separate implementation milestone.

## Existing function map

Orchestrator: 1–15. Neo-Cortex: 16–29. Super AGI: 30–60. The original detailed contract map remains preserved in git history and should be kept synchronized with implementation changes.
