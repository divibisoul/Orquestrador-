# SOUL Frente 1 — Readiness Handoff

Status is evidence-based and must not be promoted to ONLINE without live authenticated runtime evidence.

## PR20 CI baseline

The previously inspected PR20 run `33568818527` executed on a GitHub-hosted runner. The verify job reached checkout, Go setup, format check and `go vet`; its `Test` step failed in the N01→N07 federation test during peer discovery with an HMAC validation error. Subsequent race/build/backend/E2E steps were correctly skipped by GitHub. The runner itself was provisioned; the failure was in the test path, not `steps: null` provisioning.

## Storage readiness

When the Storacha/UCAN credential is absent, N07 remains startable. Storage reports `status=degraded`, `reason=missing_ucan`, `code=STORAGE_NOT_CONFIGURED`, and a `cureSuggestion`. Storage upload, CID status and object endpoints return HTTP 503 rather than pretending that storage is available.

Required production configuration for real Storacha use includes `WEB3_STORAGE_TOKEN` and `STORACHA_SPACE`.

## N05 local HMAC validation

N05 retains its external endpoint fallback. Its additional local fallback workflow can pull `ghcr.io/divibisoul/soul-n07-orquestrador:latest`, start N07 on the same GitHub-hosted runner, wait for `/health`, and run `scripts/manual-hmac-validator.mjs` against `http://localhost:8080` with an ephemeral test secret. This proves the real HTTP/HMAC path without a public N07 deployment URL.

The N05 bridge still uses its existing `SOUL_N07_URL` and `SOUL_MESH_HMAC_SECRET` inputs; no runtime replacement is introduced.

## Environment contract N01–N06

| Nucleus | N07 endpoint variable | Mesh HMAC variable |
|---|---|---|
| N01 | `VITE_SOUL_MESH_N07_URL` | `VITE_SOUL_MESH_TOKEN` / existing Mesh secret path |
| N02 | `SOUL_MESH_N07_URL` | `SOUL_MESH_HMAC_SECRET` |
| N03 | `SOUL_MESH_N07_URL` | `SOUL_MESH_HMAC_SECRET` |
| N04 | `SOUL_MESH_N07_URL` | `SOUL_MESH_HMAC_SECRET` |
| N05 | `SOUL_N07_URL` | `SOUL_MESH_HMAC_SECRET` |
| N06 | `SOUL_MESH_N07_URL` | `SOUL_MESH_HMAC_SECRET` |

Do not commit real secret values or production endpoints to source control.

## Contract prover

`node scripts/contract-15-pairs.test.js` validates all 15 unordered N01–N06 pairs in both directions using a test-only HMAC key and emits `.diagnostics/contract-15-pairs-valid.json`. It performs no network calls.

`node scripts/topology-validator.js` checks endpoint variable presence/format only and emits `.diagnostics/topology-readiness.json`. It performs no network calls.

## State semantics

`READY_FOR_CONNECTION` is the maximum state before endpoint and credential injection plus live traffic. `CONNECTED` requires real authenticated execution, HTTP success, contract 1.1.0 validation and preserved correlationId. Missing secrets are `MISSING_SECRET`/`DEGRADED`, never PASS.
