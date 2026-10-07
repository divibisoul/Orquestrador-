# N07 → Frente 2 Handoff

Data: 2026-09-02

## Canonical transport

Protocol: `soul-mesh/1`
Contract: `1.1.0`
Endpoint path: `/api/soul-mesh`

## N07 runtime

N07 is operational in the published staging runtime according to the repository's commissioning evidence. The exact public deployment URL is not hard-coded in source and must be copied from the active deployment provider into peer configuration.

N07 Mesh cryptographic traffic uses `SOUL_MESH_HMAC_SECRET` when configured. Never commit the value to this repository or to `.env` files.

## Peer variables

- N01: `VITE_SOUL_MESH_N07_URL` and `VITE_SOUL_MESH_TOKEN` where the N01 build/runtime requires bearer compatibility.
- N02: `SOUL_MESH_N07_URL`; use the shared Mesh secret/token according to the active N02 gateway security configuration.
- N03: `SOUL_MESH_N07_URL`; use `SOUL_MESH_HMAC_SECRET`.
- N04: `SOUL_MESH_N07_URL`; use the existing Mesh credential variables.
- N05: `SOUL_N07_URL`; use `SOUL_MESH_HMAC_SECRET`.
- N06: `SOUL_MESH_N07_URL`; use `SOUL_MESH_HMAC_SECRET` or the runtime's documented compatibility alias.

## N05 certification gate

N05 is not certified until `.github/workflows/n05-n07-bridge-fallback.yml` executes successfully and `scripts/manual-hmac-validator.mjs` reports PASS against the real N07 endpoint. The original N05 PR #13 runner failure occurred before any workflow step (`runner_id=0`, `steps=[]`), so fallback runner validation is intentionally separate.

## N07 configured staging gate

PR #21 contains the credentialed staging gate. It requires these GitHub Actions secrets in N07:

- `N07_APP_TOKEN`
- `WEB3_STORAGE_TOKEN`
- `SUPABASE_URL`
- `SUPABASE_SERVICE_ROLE_KEY`

A missing secret is a BLOCKED state, not a PASS.

## Baseline

Use the N07 health and E2E evidence only as baseline until a fresh authenticated run against the current deployment proves live cross-runtime connectivity. A green repository build alone is not an ONLINE certificate.
