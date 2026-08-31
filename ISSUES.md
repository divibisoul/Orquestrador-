# Recovery Issues

Baseline: `foundation/orchestrator-nexus-v1`, recovered into `recovery/consolidation`.

## RESOLVED

- **Cognitive import mismatch** — `core/cognitive/controller.go` imported the nonexistent root package `neuralfabric`; implementation lives under `core/neuralfabric`. Fixed to use the canonical package.
- **Prefrontal API mismatch in integration test** — test called `prefrontal.NewCortex()` and nonexistent `UpdatePolicies`. Test now uses `prefrontal.New()` and the existing `OptimizePolicy` operation.
- **Mesh JSON contract corruption** — `internal/mesh.Message` assigned `json:"id"` to five distinct fields. Tags are now unique (`id`, `type`, `trace_id`, `source`, `target`). CPU/GPU/NPU tags are also distinct.
- **Mesh stale-node stub** — `mesh.Registry.MarkStale` always returned zero. Nodes now track `LastHeartbeat`, and stale status is applied when the configured age is exceeded.
- **Security boundary too permissive/implicit** — policy validation now rejects invalid cost/confidence ranges and exposes explicit authentication/authorization and audit-record extension points.
- **CI supply-chain pinning** — `securego/gosec@master` replaced by the pinned `v2.22.8` action reference.

## OPEN / REQUIRES EXECUTION VERIFICATION

- Run `go list ./...`, `go build ./...`, `go test ./... -race` in a real Go toolchain. GitHub connector edits cannot themselves execute the repository's compiler.
- Resolve and document the architectural relationship between `core/orchestrator` and `internal/nexus`.
- Resolve and document the architectural relationship between `internal/mesh` and `mesh`.
- Resolve and document the architectural relationship between `internal/state` and `state`.
- Protobuf contracts exist, but generated Go/gRPC bindings and a live gRPC runtime still need validation/implementation if required by the final architecture.
- Raft remains a boundary rather than a distributed consensus implementation.
- Neural Fabric `Update`, `Save`, and `Load` remain incomplete and must not be represented as production learning/persistence.
- Super AGI training/fine-tuning boundaries remain non-operational and need explicit `not implemented` errors before being advertised as executable features.
- Gemini adapter changes from `feature/gemini-provider` have not been promoted automatically; they require separate review because the current adapter initializes credentials eagerly and its HTTP boundary requires hardening.
