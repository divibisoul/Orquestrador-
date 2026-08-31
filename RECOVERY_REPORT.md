# Recovery Report — Orquestrador-Nexus

## Scope

Repository: `divibisoul/Orquestrador-`

Recovery branch: `recovery/consolidation`

Source baseline: `foundation/orchestrator-nexus-v1`

## Actions completed

1. Created `recovery/consolidation` directly from the foundation branch, preserving the foundation history.
2. Corrected the broken `core/cognitive` import to the canonical `core/neuralfabric` package.
3. Repaired the malformed compute runtime syntax.
4. Consolidated duplicate prefrontal implementations into one canonical `core/prefrontal/Cortex`, preserving useful compatibility APIs.
5. Updated the integration tests to target the canonical APIs.
6. Fixed malformed Mesh JSON tags and made heartbeat/stale tracking operational in canonical `mesh`.
7. Refactored `cmd/nexus` to consume canonical `core/*` and `mesh` packages instead of the legacy `internal` runtime.
8. Strengthened the security policy boundary and added focused security tests.
9. Pinned Gosec to `securego/gosec@v2.22.8`.
10. Hardened the experimental Gemini adapter with lazy credentials, request authentication and generic provider errors.
11. Added explicit `not implemented` errors to Super AGI learning/fine-tuning/replay boundaries and tests proving they cannot report false success.
12. Added `tests/e2e_six_planes_test.go`, exercising Cortex → Orchestrator → Mesh → Neural Fabric → Super AGI → State.
13. Moved CI, security scanning and benchmarks to the configured self-hosted runner and added manual dispatch for CI/security.
14. Modernized GitHub Actions to current `checkout@v7` and `setup-go@v7` majors; the configured runner version 2.336.0 satisfies the documented runner requirements for these action generations.
15. Added `docs/SELF_HOSTED_RUNNER.md` describing runner prerequisites, registration hygiene and verification commands.
16. Updated README, ARCHITECTURE, STATUS, SECURITY and CONTRIBUTING documentation.

## Execution evidence

The recovery branch previously passed GitHub Actions CI and Security validation on hosted runners. The successful validation included `go vet ./...`, `go build ./...`, `go test ./... -count=1 -race`, Python adapter syntax validation, and a Gosec scan with zero reported issues.

After switching the workflows to `runs-on: self-hosted` and modern action majors, fresh CI/Security runs are required and are tracked against the current recovery head. They are an explicit acceptance gate rather than an assumption.

## Canonical architecture

- `core/orchestrator`: canonical workflow/execution engine and resilience primitives.
- `core/prefrontal`: canonical executive reasoning/planning/decision layer.
- `core/superagi`: canonical provider-neutral model and memory boundary.
- `core/neuralfabric`: canonical routing/prediction/feedback layer.
- `mesh`: canonical capability registry/discovery/health layer.
- `state`: canonical versioned state/CAS layer.
- `internal/nexus`, `internal/mesh`, `internal/state`: legacy implementations retained until a complete consumer-graph migration proves they are safe to remove.

## Honest capability boundaries

The recovery does not claim production functionality that is not present. Real distributed Raft, generated/live gRPC bindings, production mTLS/SPIFFE, durable audit storage, physical GPU/NPU backends and concrete external model runtimes remain separate implementation layers.

Neural Fabric `Update`, `Save` and `Load` remain explicit boundary operations. Super AGI training, LoRA and replay boundaries now fail explicitly with `not implemented` instead of returning false success.

## Self-hosted runner

Recovery validation, security scanning and benchmarks now run on the repository's self-hosted Actions runner. The runner must provide Linux/x64 execution, GitHub Actions runner compatibility, Go 1.23 availability through setup-go, Python 3 for the adapter checks and Docker for the Gosec container action.

A single self-hosted runner serializes jobs when it is busy; multiple registered runners are required for parallel job execution.

## Acceptance status

| Criterion | Status |
|---|---|
| Recovery branch created | DONE |
| Import/API inconsistencies identified and corrected | DONE for identified findings |
| Canonical Prefrontal implementation | DONE |
| Canonical Orchestrator executable path | DONE |
| Canonical Mesh heartbeat/stale behavior | DONE |
| False-success Super AGI learning boundaries removed | DONE |
| Six-plane E2E test added | DONE |
| Self-hosted CI configuration | DONE |
| Self-hosted security workflow | DONE |
| Self-hosted benchmark workflow | DONE |
| `go build ./...` on current recovery head | PENDING fresh self-hosted run |
| `go test ./... -race` on current recovery head | PENDING fresh self-hosted run |
| `go vet ./...` on current recovery head | PENDING fresh self-hosted run |
| Security scan on current recovery head | PENDING fresh self-hosted run |
| Remaining legacy consumer migration | PENDING |
| gRPC runtime | PENDING |
| Production distributed consensus | PENDING |
| `main` update | PENDING explicit approval |
| `v0.1.0-recovery` tag | PENDING explicit approval |

## Merge policy

`main` remains intentionally untouched until the current self-hosted CI and security gates are green and the remaining legacy consumer graph has been reviewed. No destructive deletion is performed merely to make the tree look cleaner.
