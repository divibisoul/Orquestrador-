# N07 — Orquestrador

N07 is the SOUL Orchestrator nucleus. Its responsibility is to coordinate the other nuclei through one typed, traceable protocol and connect the cognitive, federation, composition and compute layers:

```text
N01 ─┐
N02 ─┤
N03 ─┤
N04 ─┼──► N07 Orquestrador ──► Neural ──► Prefrontal ──► Compute
N05 ─┤          │
N06 ─┘          ├──► Soul Mesh / discovery / delegation
               ├──► Capability composition / aggregation
               └──► SuperGPU / parallel federation
```

The `cognitive.execute` route is the concrete local end-to-end path: neural forward pass → prefrontal risk/utility decision → decision commit → compute execution, with the same `trace_id` throughout. N07 now also provides a cooperative boundary for work that is not locally registered: local execution is preferred; otherwise the capability is discovered and delegated through the same canonical Soul Mesh.

## Runtime contract

- Language: Go 1.25.
- Protocol version: `N07.v1`; canonical Soul Mesh contract: `1.1.0`.
- Every message carries trace/correlation identity, source, target and execution metadata.
- No silent fallback: unavailable routes and invalid inputs return explicit errors.
- The compute layer is hardware-aware. It discovers NVIDIA/AMD driver tools when present and exposes a backend interface for native GPU implementations. The built-in CPU backend is an execution backend, not a GPU simulation.
- Shutdown is context-aware and releases compute reservations.
- Federation uses the existing Mesh transport and live capability discovery; it does not create a second Mesh.

## Public runtime surface

### Core N07
`New`, `Register`, `Route`, `Submit`, `Execute`, `Cancel`, `Status`, `Health`, `Stats`, `Shutdown`.

### Neural Network
`New`, `AddEdge`, `RemoveEdge`, `Activate`, `Forward`, `Learn`, `Normalize`, `Attention`, `Backprop`, `Health`.

### Neocórtex Pré-frontal
`New`, `Evaluate`, `Plan`, `Prioritize`, `Inhibit`, `Select`, `ValidateAction`, `Commit`, `Recall`, `Health`.

### SuperGPU
`New`, `Discover`, `Select`, `Reserve`, `Release`, `Execute`, `Batch`, `MemoryStats`, `Health`, `Shutdown`.

### Federation
`RegisterPeer`, `RemovePeer`, `Snapshot`, `Stats`, `DiscoverLive`, `Discover`, `Delegate`, `ExecuteParallel`.

### Composition
`NewCapabilityPlan`, `Validate`, `Execute` with dependency-aware serial/parallel waves and component provenance.

## Six-peer topology

N07 has explicit bidirectional structural channels for N01–N06. Every peer is represented as one IN and one OUT channel through the canonical Soul Mesh. The runtime topology is exposed by `GET /topology`.

The supported transport classes are represented as a resolution contract: `IN_PROCESS`, `LOOPBACK_HTTP`, `HTTP`, `REALTIME`, and `EVENT`. Concrete transport selection remains provider/adapter driven; N07 does not fake an unavailable transport.

## Runnable control plane

The repository contains an executable N07 server at `cmd/orchestrator` exposing:

- `POST /api/soul-mesh` — canonical Mesh ingress;
- `GET /health` — local runtime and federation health;
- `GET /stats` — local plus federation telemetry;
- `GET /topology` — explicit N07 ↔ N01..N06 topology contract;
- `GET /federation/peers` — configured peers and current state;
- `GET /federation/discovery` — concurrent live capability discovery;
- `POST /federation/execute-parallel` — parallel distributed tasks with aggregated results;
- `POST /federation/compose` — dependency-aware capability composition;
- `GET /metrics` — Prometheus-compatible scalar metrics.

Run:

```bash
go run ./cmd/orchestrator
```

The address defaults to `:8080` and can be overridden with `N07_ADDR`.

## Hybrid execution model

For a Mesh request N07 follows:

`validate → authenticate → capability resolve → local route if executable → otherwise federation discovery/selection → delegated execution → response/correlation`.

A capability is not considered operational merely because it is named in documentation. Peer discovery updates the live capability evidence. The router considers health, observed success, latency and in-flight load before selecting a peer, with bounded retry and circuit protection.

## Capability composition and SuperCompute

A `CapabilityPlan` represents a composed capability as a traceable set of component steps with explicit dependencies. Independent steps can execute in parallel; dependent steps execute only after prerequisites succeed. Required and optional steps are represented explicitly and failures remain observable.

`ExecuteParallel` provides the distributed SuperCompute primitive: one parent task can fan out independent capabilities to multiple nuclei, create child correlation identities, execute concurrently and return a complete result set with per-task duration and failure information.

This is the logical SOUL Super GPU boundary:

`TASK → DECOMPOSITION → ROUTING → PARALLEL EXECUTION → AGGREGATION → VALIDATION → RESULT`.

It complements, rather than replaces, N04's execution-worker specialization and N01's registry/device/cognitive resources.

## Verification

The repository includes executable tests for protocol serialization/trace propagation, neural forward/learning paths, prefrontal policy, compute execution/batching, N07 orchestration, federation scoring/retry, arbitrary payload delegation and parallel aggregation.

CI remains the authority for repository-wide build/test/race/security status. A branch or documentation statement is never treated as proof of operational readiness without evidence from the exact revision.

## Hardware boundary

A universal Go binary cannot truthfully claim to execute CUDA/HIP kernels on every device. N07 therefore uses a strict `Backend` interface. Native NVIDIA/AMD backends can be attached where their vendor runtime and drivers are installed; otherwise the runtime reports the available execution device and uses the concrete CPU backend. This keeps the core real and testable instead of embedding fake GPU results.
