# Recovery Status

This document distinguishes source-level implementation from production capability.

| Component | Current state | Notes |
|---|---|---|
| Go module / entrypoint | Implemented | Recovery CI validates formatting, vet, build and race tests. |
| Orchestrator | Implemented primitives | Workflow execution, retry, parallel/distributed execution and resilience primitives exist. Canonical domain package: `core/orchestrator`. |
| Neo-Cortex / Prefrontal | Implemented primitives | Single canonical Cortex implementation with context, fusion, reasoning, planning, policy feedback and compatibility API. |
| Neural Fabric | Partial | Routing/encoding/feedback primitives exist; persistence and adaptive weight updates remain incomplete. |
| Super AGI | Partial | Provider-neutral interfaces and memory/verification boundaries exist; concrete model/training runtimes are not complete. Learning/fine-tuning/replay boundaries now return explicit `not implemented` errors. |
| Mesh | Partial | Canonical `mesh` package provides registration, discovery, heartbeat and stale marking; distributed transport remains incomplete. |
| State | Partial | Canonical `state` package provides versioned storage/CAS; real distributed consensus is not complete. |
| Security | Partial | Basic policy validation and explicit auth/audit boundaries exist; production identity, mTLS and durable audit are pending. |
| Protobuf | Contract | `.proto` contracts exist; generated bindings/live gRPC runtime remain a future integration layer. |
| Gemini | Experimental | Isolated provider adapter; kept outside the canonical Go control plane until deliberately promoted. |
| GPU/NPU | Boundary | Device abstractions and local/simulated execution exist; physical accelerator backends are not claimed as implemented. |
| CI | Operational | CI, security and benchmark workflows target the self-hosted runner and support manual dispatch where appropriate. |
| Legacy `internal/*` | Removed from recovery runtime | `internal/nexus`, `internal/mesh` and `internal/state` were removed after consumer-graph verification; their source remains recoverable in Git history and the original foundation branch. |
| E2E | Implemented test | Six-plane test covers Cortex → Orchestrator → Mesh → Neural Fabric → Super AGI → State. |

## Self-hosted execution

The repository expects a self-hosted GitHub Actions runner for recovery validation. See `docs/SELF_HOSTED_RUNNER.md`.

## Recovery rule

No stub is allowed to return successful output that implies work happened when the operation is not implemented. Boundary methods must either perform the real operation or return an explicit `not implemented` error, with tests.
