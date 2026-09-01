# N07 Android integration contract

This document defines the boundary between the N07 Go runtime and the Android application layer. It does not embed provider secrets or require the Android app to reimplement the Mesh runtime.

## Runtime boundary

```text
Android app
   |
   | HTTPS / canonical Soul Mesh
   v
N07 /api/soul-mesh
   |
   +--> local N07 capabilities
   |
   +--> discovery / routing
   |
   +--> N01..N06 peer delegation
   |
   +--> neural federation / SuperGPU
```

## Required Android configuration

- `SOUL_N07_BASE_URL`: HTTPS base URL of the deployed N07 runtime.
- `SOUL_MESH_HMAC_SECRET`: provisioned through the secure application/backend configuration path; never hard-code in source or persist in logs.
- `SOUL_MESH_CONTRACT_VERSION`: `1.1.0` for canonical Mesh interoperability.

For a mobile client, the preferred production topology is Android -> authenticated HTTPS ingress -> N07. Directly exposing a shared long-lived Mesh secret in an APK is not considered a secure production design; use a trusted backend/edge credential exchange when possible.

## Startup sequence

1. `GET /health`
2. verify N07 reports the expected protocol/contract version;
3. query discovery through the canonical Mesh surface;
4. establish a correlation ID for each user request;
5. invoke a capability through `/api/soul-mesh`;
6. use returned route/correlation metadata for tracing;
7. surface explicit errors instead of fabricating fallback success.

## Android request contract

Every request should carry:

- source nucleus/client identity;
- target (`N07`);
- operation/capability;
- correlation ID;
- timestamp;
- bounded payload;
- authentication material supplied by the trusted ingress.

## Operational behavior

The Android layer must treat N07 as an orchestration endpoint. It should not duplicate N01-N06 capability ownership, neural models, or SuperGPU scheduling. N07 discovers and delegates to the appropriate nucleus.

## APK handoff

N07 is a Go backend/orchestration service and is not itself an Android APK. The Android APK belongs to the Android client/runtime repository. The N07 deliverable is the stable network contract, runnable service, health/discovery surface and integration documentation required for that APK to connect safely.

## Readiness boundary

The N07 repository is ready for Android integration when exact-head CI is green and the deployed endpoint passes health, discovery, authenticated request/response and correlation smoke tests. Full seven-runtime E2E remains an integration-environment gate, not an excuse to invent success inside N07.
