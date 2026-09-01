# N07 system state

## Current phase
RELEASE CLOSURE / ACTIVE CROSS-FRONT INTEGRATION

## Ownership
N07 owns orchestration, neural federation, routing, capability composition and SuperGPU control-plane surfaces. N01-N06 remain independent runtimes and expose adapters into the common Mesh.

## Completed structural areas
- Canonical Mesh ingress/egress and N01..N07 identity handling.
- Discovery, executable capability routing, semantic-version routing and TTL cache.
- Delegation with correlation, HMAC, replay protection, timeout/deadline and circuit handling.
- Neural Federation N07 -> N01..N06 with bounded parallel admission and canonical Target routing.
- Capability/tool/agent composition and SuperGPU parallel orchestration.
- Observability for peers, latency, failures, in-flight work and correlation.
- Supabase + Web3 Storage backend adapters and isolated regression coverage.
- Android integration contract documented for downstream app/APK work.

## Current evidence
- Main contains the finite 40-function queue and release gates.
- Backend/storage regression branch passed Format, Vet, Test, Race Test and Build in GitHub Actions run #310/verify; follow-up normalization run #326 is active/queued for the updated workflow.
- N07 CI uses exact-head evidence and does not claim live runtime success without E2E proof.

## Next finite gates
1. Retarget/reconcile backend/storage PR against current main and validate merge-base CI.
2. Reconcile latest N01-N06 response signatures and health semantics.
3. Run six-peer live E2E where runtimes are concurrently reachable.
4. Run release smoke/deploy and hand off to Android/APK stage.

## Rule
A failure in one area triggers correction and revalidation; it does not cancel remaining release work. Do not expand v1 scope with optional features before these blockers are closed.
