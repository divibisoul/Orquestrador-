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

## Secrets and restricted mode

N07 does not invent missing credentials. When one or more external credentials are absent, the service remains available for local diagnostics and execution but reports `status: degraded` and `secret_mode: restricted` on `/health`. External storage and Supabase persistence are not attempted by simulated staging.

See `docs/SECRETS.md` for the four required deployment secrets, their purpose, and the credentialed-staging procedure.

## Release

The v1 release is finite. The repository tracks structural, integrated and online states in `RELEASE_GATES_V1.md` and `SOUL_EXECUTION_QUEUE.md`. CI is required evidence; online completion additionally requires live peer E2E.
