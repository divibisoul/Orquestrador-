# SOUL — N07 Orchestrator

N07 is the orchestration and federated-compute control plane for the SOUL multi-nucleus system. N01–N06 remain independently deployable runtimes connected through the canonical Mesh contract.

## Runtime surfaces

- `GET /v1/capabilities` — authenticated capability inventory.
- `GET /v1/health` — authenticated runtime health.
- `POST /v1/execute` — authenticated operation execution.
- `POST /v1/intent` — authenticated intent-to-operation mapping.
- `POST /v1/storage/upload` — authenticated artifact upload through Web3 Storage-compatible adapter.
- `GET /v1/storage/status/{cid}` — authenticated storage status.
- `GET /v1/storage/object/{cid}` — authenticated gateway object URL resolution.

## Federation

N07 provides canonical Mesh ingress, discovery, executable-capability routing, delegation, neural federation and bounded parallel SuperGPU orchestration. Peer runtimes keep ownership of their native capabilities and tools.

## Backend

Supabase is used for durable run/artifact metadata through server-side credentials. Web3 Storage is used for content uploads; the returned CID can be resolved through the configured IPFS gateway. Server secrets are environment configuration and must never be embedded in the Android APK.

## Android handoff

The downstream Android application should treat N07 as an HTTPS service boundary. Configure endpoint and application token at deployment time, not in source control. The client should call health, discovery/capabilities, then execute/intent endpoints as required by the app UX.

See `docs/ANDROID_INTEGRATION.md` for the current request/response contract and startup sequence.

## Release

The v1 release is finite. The repository tracks structural, integrated and online states in `RELEASE_GATES_V1.md` and `SOUL_EXECUTION_QUEUE.md`. CI is required evidence; online completion additionally requires live peer E2E.


## SARA regenerative service

N07 integrates the SARA regenerative core without creating a new SOUL nucleus or
duplicating ARA/ETR/ITR.

Server-side configuration:
- `SARA_SERVICE_URL`
- `SARA_SERVICE_TOKEN`
- `SARA_REQUEST_TIMEOUT` (optional, default 30s)

When configured, N07 registers:
`sara.cycle@1.0.0`, `sara.audit@1.0.0`,
`sara.regenerate@1.0.0`, `sara.state@1.0.0`,
`sara.capabilities@1.0.0`.

SARA integration details: `docs/SARA_INTEGRATION.md`.
