# Recovery Status

This document distinguishes source-level implementation from production capability.

| Component | Current state | Notes |
|---|---|---|
| Go module / entrypoint | Implemented | Requires actual `go build ./...` verification after recovery changes. |
| Orchestrator | Implemented primitives | Workflow execution, retry and resilience primitives exist. Canonical domain package: `core/orchestrator`. |
| Neo-Cortex / Prefrontal | Implemented primitives | Context, fusion, reasoning, planning, policy feedback. |
| Neural Fabric | Partial | Routing/encoding primitives exist; persistence and adaptive learning remain incomplete. |
| Super AGI | Partial | Provider-neutral interfaces and memory/verification boundaries exist; concrete model/training runtimes are not complete. |
| Mesh | Partial | Registration, discovery, heartbeat and stale marking exist; distributed transport is not complete. |
| State | Partial | Versioned/cache boundaries exist; real distributed consensus is not complete. |
| Security | Partial | Basic policy validation and explicit auth/audit boundaries exist; production identity, mTLS and durable audit are pending. |
| Protobuf | Contract | `.proto` contracts exist; generated bindings/live gRPC runtime require separate verification/implementation. |
| Gemini | Experimental | Kept isolated from canonical Go runtime until security and lifecycle hardening is complete. |
| GPU/NPU | Boundary | Device abstractions exist; hardware backends are not claimed as implemented. |
| CI | Configured | Gosec action is pinned; actual green status must be confirmed from workflow runs. |

## Recovery rule

No stub is allowed to return successful output that implies work happened when the operation is not implemented. Future stub conversions should return explicit `not implemented` errors and have tests.
