# N07 Android integration contract

This document defines the boundary between the N07 Go runtime and the Android application layer. It does not embed provider secrets or require the Android app to reimplement the Mesh runtime.

## Runtime boundary

```text
Android app
   |
   | HTTPS + application bearer token
   v
N07 /execute
   |
   +--> N07 local capabilities
   +--> discovery / routing
   +--> N01..N06 peer delegation
   +--> neural federation / SuperGPU
```

The canonical inter-nucleus Mesh remains available through `/api/soul-mesh` and is protected by the Mesh HMAC contract. The Android application should use `/execute` for its external application boundary unless it is itself part of the trusted Mesh.

## Required deployment configuration

- `N07_HTTP_ADDR`: bind address for the N07 service, default `:8080`.
- `N07_APP_TOKEN`: application bearer token required by `/execute` and `/status`; never embed the value in source control.
- `SOUL_N07_BASE_URL`: HTTPS base URL of the deployed N07 runtime.
- `SOUL_MESH_HMAC_SECRET`: trusted Mesh secret for N01..N06 federation; never hard-code in source or persist in logs.
- `SOUL_MESH_CONTRACT_VERSION`: `1.1.0` for canonical Mesh interoperability.

For a mobile client, the preferred production topology is Android -> authenticated HTTPS ingress -> N07. A shared long-lived Mesh secret should not be placed directly in an APK. A trusted backend/edge can exchange an application credential for the internal Mesh credential boundary.

## Startup sequence

1. `GET /health`.
2. Verify N07 reports an expected operational state.
3. Establish the authenticated application session/token.
4. Establish a correlation ID for each user request at the application layer when tracing is available.
5. Invoke `/execute` with the operation and bounded payload.
6. Use returned operation/trace metadata for diagnostics.
7. Surface explicit errors; never fabricate fallback success.

## Android request contract

`POST /execute` requires:

- `Authorization: Bearer <application-token>`;
- `Content-Type: application/json`;
- `operation` as a registered N07 operation;
- optional numeric `payload` matching the operation input contract;
- optional `metadata` for trace/correlation context.

For trusted Mesh calls, use the canonical envelope and HMAC path instead of the application bearer boundary.

## Operational behavior

The Android layer treats N07 as the orchestration endpoint. It does not duplicate N01-N06 capability ownership, neural models, or SuperGPU scheduling. N07 discovers and delegates to the appropriate nucleus.

## APK handoff

N07 is a Go backend/orchestration service and is not itself an Android APK. The APK belongs to the Android client/runtime repository. The N07 deliverable is the stable network contract, runnable service, health/discovery surface, authentication boundary and integration documentation required for the Android/Google Studio stage.

## Release smoke test

Before APK integration, verify:

```text
GET /health            -> 200
POST /execute          -> 401 without bearer
POST /execute          -> 401 with wrong bearer
POST /execute + bearer -> authenticated operation path
GET /status + bearer  -> runtime status
GET /metrics          -> metrics output
```

## Readiness boundary

N07 is ready for Android integration when exact-head CI is green and the deployed endpoint passes health, authenticated execution, discovery, correlation and error-path smoke tests. Full seven-runtime E2E remains an integration-environment gate and must be reported separately from structural completion.
