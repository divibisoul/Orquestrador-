# SOUL Mesh Topology

Canonical protocol: `soul-mesh/1`.
Canonical contract: `1.1.0`.

N01, N02, N03, N04, N05 and N06 are independent runtime nuclei. N07 is the orchestration, federation and prefrontal control-plane runtime. All communication uses the existing Soul Mesh; no parallel transport is introduced.

## First-class topology

| Nucleus | Runtime role | N07 endpoint variable |
|---|---|---|
| N01 | Android / gateway peer | `VITE_SOUL_MESH_N07_URL` |
| N02 | Cognition / inference | `SOUL_MESH_N07_URL` |
| N03 | Audio / speech | `SOUL_MESH_N07_URL` |
| N04 | Documents / tools | `SOUL_MESH_N07_URL` |
| N05 | Conversation | `SOUL_N07_URL` |
| N06 | Cognitive support | `SOUL_MESH_N07_URL` |
| N07 | Federation / orchestration / SuperGPU control plane | runtime-owned |

N02 is a first-class node of the mesh. Its canonical gateway was merged into `main` in N02 PR #13 with merge commit `28ab385589ea664c92c0adbeeeb3e26752a645e4`.

## Authentication

Production Mesh traffic uses the shared `SOUL_MESH_HMAC_SECRET` where the runtime bridge requires HMAC. Application control-plane endpoints additionally use the N07 application bearer token where configured.

## Readiness semantics

`READY_FOR_CONNECTION` means the endpoint variable is present and syntactically valid. `CONNECTED` is reserved for authenticated runtime traffic with HTTP success, canonical contract validation and correlationId preservation. `NOT_CONFIGURED`, `MISSING_SECRET` and `DEGRADED` are not success states.
