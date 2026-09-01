# N07 — Orquestrador

N07 is the SOUL Orchestrator nucleus. Its responsibility is to coordinate the other nuclei through one typed, traceable protocol and to connect three execution layers:

```text
N01..N06
   │
   ▼
N07 Orquestrador ──► Neural Network ──► Prefrontal Cortex
        │                    │                  │
        └────────────────────┴──────────────────┘
                         │
                         ▼
                    SuperGPU Runtime
```

## Runtime contract

- Language: Go 1.25.
- Protocol version: `N07.v1`.
- Every message carries `trace_id`, source, target, operation, sequence and optional deadline/metadata.
- No silent fallback: unavailable routes and invalid inputs return explicit errors.
- The compute layer is hardware-aware. It discovers NVIDIA/AMD driver tools when present and exposes a backend interface for native GPU implementations. The built-in CPU backend is an execution backend, not a GPU simulation.
- Shutdown is context-aware and releases compute reservations.

## 40 implemented functions

### Orquestrador — 10
`New`, `Register`, `Route`, `Submit`, `Execute`, `Cancel`, `Status`, `Health`, `Stats`, `Shutdown`, plus built-in route registration.

### Neural Network — 10
`New`, `AddEdge`, `RemoveEdge`, `Activate`, `Forward`, `Learn`, `Normalize`, `Attention`, `Backprop`, `Health`.

### Neocórtex Pré-frontal — 10
`New`, `Evaluate`, `Plan`, `Prioritize`, `Inhibit`, `Select`, `ValidateAction`, `Commit`, `Recall`, `Health`.

### SuperGPU — 10
`New`, `Discover`, `Select`, `Reserve`, `Release`, `Execute`, `Batch`, `MemoryStats`, `Health`, `Shutdown`.

## Verification

The repository includes executable Go tests for protocol serialization/trace propagation, neural forward/learning paths, prefrontal decision policy, compute execution/batching, and the end-to-end N07 orchestration path.

## Hardware boundary

A universal Go binary cannot truthfully claim to execute CUDA/HIP kernels on every device. N07 therefore uses a strict `Backend` interface. Native NVIDIA/AMD backends can be attached where their vendor runtime and drivers are installed; otherwise the runtime reports the available execution device and uses the concrete CPU backend. This keeps the core real and testable instead of embedding fake GPU results.
