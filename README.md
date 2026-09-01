# N07 — Orquestrador

N07 is the SOUL Orchestrator nucleus. Its responsibility is to coordinate the other nuclei through one typed, traceable protocol and connect the cognitive, federation and compute layers:

```text
N01..N06
   │
   ▼
N07 Orquestrador
   │
   ├──► Neural Network ──► Prefrontal Cortex ──► Compute Backend
   │
   ├──► Soul Mesh Gateway
   │      └──► Peer Federation: discovery → scoring → delegation → retry
   │
   └──► SuperGPU Runtime
```

The `cognitive.execute` route is the concrete local end-to-end path: neural forward pass → prefrontal risk/utility decision → decision commit → compute execution, with the same `trace_id` throughout.

## Runtime contract

- Language: Go 1.25.
- Protocol version: `N07.v1`; canonical Soul Mesh contract: `1.1.0`.
- Every message carries trace/correlation identity, source, target, operation and sequence data.
- No silent fallback: unavailable routes and invalid inputs return explicit errors.
- The compute layer is hardware-aware. It discovers NVIDIA/AMD driver tools when present and exposes a backend interface for native GPU implementations. The built-in CPU backend is an execution backend, not a GPU simulation.
- Shutdown is context-aware and releases compute reservations.
- `orchestrator.Federation` adds peer capability discovery through the existing `mesh.describe` capability, dynamic candidate scoring and bounded delegation. It does not create a parallel transport protocol.

## 40 public runtime functions

### Orquestrador — 10
`New`, `Register`, `Route`, `Submit`, `Execute`, `Cancel`, `Status`, `Health`, `Stats`, `Shutdown`.

### Neural Network — 10
`New`, `AddEdge`, `RemoveEdge`, `Activate`, `Forward`, `Learn`, `Normalize`, `Attention`, `Backprop`, `Health`.

### Neocórtex Pré-frontal — 10
`New`, `Evaluate`, `Plan`, `Prioritize`, `Inhibit`, `Select`, `ValidateAction`, `Commit`, `Recall`, `Health`.

### SuperGPU — 10
`New`, `Discover`, `Select`, `Reserve`, `Release`, `Execute`, `Batch`, `MemoryStats`, `Health`, `Shutdown`.

## Runnable control plane

The repository contains an executable N07 server at `cmd/orchestrator` exposing:

- `POST /api/soul-mesh` — canonical Mesh ingress;
- `GET /health` — runtime health;
- `GET /stats` — counters and latency telemetry;
- `GET /metrics` — Prometheus-compatible scalar metrics.

Run:

```bash
go run ./cmd/orchestrator
```

The address defaults to `:8080` and can be overridden with `N07_ADDR`.

## Federation usage

Register only real N01–N06 peers with their Mesh endpoint and known capability evidence. Then call `DiscoverLive` to refresh capability evidence from each peer's `mesh.describe` response. `Delegate` selects an eligible peer using health, success rate, latency and in-flight load, then performs bounded retries with context-aware backoff.

This layer intentionally preserves N07's existing local registration/routing path. It is a cooperative extension for work that is not locally implemented, not a replacement for the canonical Mesh.

## Verification

The repository includes executable Go tests for protocol serialization/trace propagation, neural forward/learning paths, prefrontal decision policy, compute execution/batching, N07 cognitive orchestration and the federation routing/retry primitives.

CI remains the authority for repository-wide build/test/race/security status. A branch or documentation statement is never treated as proof of operational readiness without evidence from the exact revision.

## Hardware boundary

A universal Go binary cannot truthfully claim to execute CUDA/HIP kernels on every device. N07 therefore uses a strict `Backend` interface. Native NVIDIA/AMD backends can be attached where their vendor runtime and drivers are installed; otherwise the runtime reports the available execution device and uses the concrete CPU backend. This keeps the core real and testable instead of embedding fake GPU results.
