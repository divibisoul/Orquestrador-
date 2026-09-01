# N07 — Orquestrador

N07 is the SOUL Orchestrator nucleus. It remains an independent AI/runtime and an active parallel engineering front for N01–N06. Its responsibility is to coordinate the nuclei through one typed, traceable protocol and compose cognitive, federation and compute capabilities without copying peer runtimes.

```text
N01 ─┐
N02 ─┤
N03 ─┤
N04 ─┼──► N07 Orquestrador ──► Neural ──► Prefrontal ──► Compute
N05 ─┤          │
N06 ─┘          ├──► Soul Mesh / discovery / routing / delegation
               ├──► capability composition / fusion
               └──► federated SuperGPU / parallel execution
```

## Runtime contract

- Language: Go 1.25.
- Protocol: `N07.v1`; canonical Soul Mesh contract: `1.1.0`.
- Every message carries source, target, message identity, correlation identity, nonce, timestamp and type.
- HMAC authentication, replay protection, timestamp skew and optional TTL are enforced at the protocol boundary.
- Payloads are bounded by the HTTP gateway and validated before execution.
- Local execution is preferred; unavailable capabilities may be discovered and delegated through the canonical Mesh.
- No physical GPU claim is made without a real provider/backend. The logical SuperGPU is distributed parallel execution across nuclei.

## Public runtime surface

### Core Orchestrator
`New`, `Register`, `Route`, `Submit`, `Execute`, `Cancel`, `Status`, `Health`, `Stats`, `Shutdown`.

### Neural
`New`, `AddEdge`, `RemoveEdge`, `Activate`, `Forward`, `Learn`, `Normalize`, `Attention`, `Backprop`, `Health`.

### Prefrontal
`New`, `Evaluate`, `Plan`, `Prioritize`, `Inhibit`, `Select`, `ValidateAction`, `Commit`, `Recall`, `Health`.

### SuperGPU
`New`, `Discover`, `Select`, `Reserve`, `Release`, `Execute`, `Batch`, `MemoryStats`, `Health`, `Shutdown`.

### Federation
`RegisterPeer`, `RemovePeer`, `PeersSnapshot`, `DiscoverPeers`, `Discover`, `Call`, `CallWithCorrelation`, `CallBest`, retry/circuit protection, TTL discovery cache and peer telemetry.

### Composition and agent federation
N07 exposes explicit agent identity through `N07Agents`, capability composition through the canonical fusion gateway, and parallel federated task execution with parent/child correlation and per-task duration.

## Mesh inputs and outputs

N07 exposes one canonical ingress: `POST /api/soul-mesh`. It accepts requests from N01–N06 and returns correlated responses or explicit errors. Structural peer links exist for all six base nuclei, with IN/OUT channel metadata exposed through `GET /topology`.

The transport resolution model is adapter-based: `IN_PROCESS`, `LOOPBACK_HTTP`, `HTTP`, `REALTIME` and `EVENT`. Only transports with concrete implementation/configuration are counted as operational.

## Control plane

The runnable N07 control plane exposes:

- `POST /api/soul-mesh` — canonical Mesh ingress;
- `GET /health` — local health;
- `GET /status` — runtime statistics;
- `GET /topology` — N07↔N01..N06 topology;
- `GET /federation/peers` — configured peer state;
- `GET /federation/discovery` — concurrent live peer discovery;
- `GET /metrics` — Prometheus-compatible telemetry.

## Distributed execution

The canonical federation can discover a capability, score peers using health/latency/failure evidence, delegate with correlation, retry transient failures and trip a circuit after repeated failures.

`mesh.supergpu.parallel` fans out independent tasks with unique child correlation IDs, records per-task latency, preserves partial failures and returns an aggregated result. `mesh.fusion.execute` composes multiple capabilities while retaining component provenance.

The logical pipeline is:

`TASK → DISCOVERY → CAPABILITY ROUTING → DECOMPOSITION → PARALLEL EXECUTION → AGGREGATION → VALIDATION → RESPONSE`.

## Verification

Tests cover protocol validation/HMAC/replay/TTL, N07 runtime behavior, peer loading, routing, retry/circuit behavior, discovery caching, topology and federated gateway paths. GitHub Actions on the exact revision remains the authority for final build/test/race/security readiness.

## Hardware boundary

N07 does not claim native CUDA/HIP execution merely from device discovery. A real accelerator backend must be attached to the compute boundary. This keeps the logical SuperGPU truthful and portable while allowing native providers to be added without replacing the orchestration core.
