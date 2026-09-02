# SOUL — FRENTE 2 REPORT

## Scope
Continuation of the existing N07 PR #20 hardening work on `fix/n07-storacha-runtime-hardening-2026-09-01`. Current GitHub state is authoritative; concurrent front changes are preserved.

## CI gate history

- Requested baseline: run `33568818527`.
- Subsequent real runs exposed new root causes; no green state was declared prematurely.
- Run `33621586782`: failed at Format because `backend/backend.go` was not gofmt-formatted.
- Run `33621786798` (run #414): Format passed, Vet failed because `backend/backend_test.go` still called the pre-split `Status(context.Context, cid)` API.
- Run `33621855794` (run #415): Format and Vet passed. Test exposed five real regressions:
  1. modern Storacha unit fixtures lacked configured credential state;
  2. degraded-storage tests still expected the old `degraded` status;
  3. degraded-storage tests used a CID rejected by the stricter validator;
  4. N01→N07 federation returned `type is required` because lowercase wire `type=request` was not normalized to a canonical kind;
  5. the circuit classification test did not recognize `INVALID_MESH_CONTRACT` as a non-connectivity failure.
- Run #420 was dispatched after the latest contract/test corrections and was queued at report creation time. Its result must be checked before any claim of green CI.

## Permanent hardening applied

- Peer response verification now preserves the received HMAC in the reconstructed `MeshEnvelope` before cryptographic verification.
- Peer responses pass through `protocol.ValidateMeshWireResponse` before HMAC verification.
- Structured contract errors use `INVALID_MESH_CONTRACT` and identify the invalid field.
- Circuit opening is restricted to connectivity/transport-style failures; HMAC, authentication, replay, correlation and structured contract errors do not increment the circuit failure counter.
- Lowercase wire request/response/error types are normalized consistently in the canonical Mesh gateway.
- Storage without authorized credential material reports `not-configured`, never a fabricated healthy state.
- `STORACHA_UCAN` is accepted as a protected credential input alongside `WEB3_STORAGE_TOKEN`; no credential is committed to source.
- Storage commissioning workflow explicitly distinguishes `not-configured` from real upload success and logs a deterministic SKIPPED message when credentials are absent.

## Storage commissioning

Real Storacha upload commissioning is **BLOCKED/NOT-CONFIGURED** in this environment because no real UCAN/token is available. The workflow does not invent one and does not mark upload commissioning as PASS.

## Files added/changed by this front

- `protocol/validate.go`
- `protocol/validate_test.go`
- `mesh/peers.go`
- `mesh/peers_circuit_test.go`
- `mesh/http.go`
- `backend/storage.go`
- `backend/backend.go`
- `backend/backend_test.go`
- `backend/storage_state_test.go`
- `backend/storage_degraded_test.go`
- `.github/workflows/n07-e2e-commissioning.yml`
- `Manual-Guide.md`
- `SOUL-FRENTE-2-REPORT.md`

## Merge gate

PR #20 must remain open until the current branch reaches a completed CI run with Test, Race, Build, Backend regression and federation E2E green. Storage may remain `not-configured` when no real credential is provisioned, but that state must remain explicit and must not be confused with successful real-upload commissioning.
