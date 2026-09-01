# N07 system state

## Current phase
FINAL RELEASE RECONCILIATION

## Ownership
N07 owns orchestration, neural federation, routing, capability composition and the SuperGPU control plane. N01-N06 remain independent runtimes and expose adapters into the common Mesh.

## Completed structural and integrated areas
- Canonical Mesh ingress/egress and N01..N07 identity handling.
- Discovery, executable capability routing, semantic-version routing and TTL cache.
- Delegation with correlation, HMAC, replay protection, timeout/deadline and circuit handling.
- Neural Federation N07 -> N01..N06 with bounded parallel admission and canonical Target routing.
- Capability/tool/agent composition and SuperGPU parallel orchestration.
- Observability for peers, latency, failures, in-flight work and correlation.
- Unified authenticated backend with Supabase run/artifact persistence and Web3 Storage upload/status/object access.
- SuperGPU backend operation registration and production release migration.
- Backend regression suite previously passed Format, Vet, Test, Race Test and Build on integration head `7441c8b00204133ffcf5c16163d2e349f2714354`.
- Backend integration PR #19 was merged into `main` as `8124ec399e669347450d710e3d44e586b9df863a`.
- N07 multi-peer federation E2E coverage is present in `mesh/n01_n07_federation_e2e_test.go` and is executed against N04/N05/N06 federation paths with canonical correlation/HMAC assertions.
- Android integration contract remains documented for downstream HTTPS/APK integration.

## Current cross-front evidence
- Current N07 `main` at release-reconciliation commit `00cb3f88da7a2f20a66f6a493b763243c92585d5` immediately before this state commit; the prior auto-format action normalized the new federation E2E test.
- N02 Mesh endpoint contract uses `soul-mesh/1`, contract `1.1.0`, seven-nucleus identity, correlation validation, HMAC/Bearer authentication, retry and capability discovery.
- N04 Mesh CI completed successfully with dependency installation, Mesh typecheck and 17/17 contract/runtime tests passing.
- N06 Channel Contract CI completed successfully on the seven-nucleus canonical manifest.
- N01 Mesh regression run at `831658443c06c80893220a8ba63ad0c0db61050e` failed without persisted job steps/log blob; a retry reproduced the same failure mode.
- N02 Mesh CI run at `e0d9fce60daec2cc046e32a15056f144c52f31b6` failed without persisted job steps/log blob; a retry reproduced the same failure mode.
- N03 Mesh CI is blocked by the current `main` dependency state: the workflow requires `npm ci`, while the current `main` did not contain `package-lock.json`. PR #9 was opened to generate a lockfile from the exact package manifest, but its Actions job also failed before step logs were persisted.
- N05 Mesh CI is currently failing on the latest main revisions; the latest retry has not produced persisted job-step logs, so no unsupported code-level root cause is asserted.

## Live E2E closure boundary
- The repository contains real federation execution and E2E validation code, but no proof is recorded here of a simultaneously reachable, externally deployed N01+N02+N03+N04+N05+N06 runtime set. The live deployment endpoints/secrets are not present in repository configuration and cannot be inferred safely.
- Structural/integrated E2E is therefore validated; external six-runtime commissioning remains environment-gated and must fail closed rather than be represented as simulated success.

## Release decision
- N07 structural release surfaces are complete and the finite E2E test harness is committed.
- Overall SOUL v1 is not marked ONLINE while N01/N02/N03/N05 CI and the external six-runtime commissioning gate remain unresolved.
- No optional feature scope is opened while these finite blockers remain.

## Rule
A detected failure is corrected and revalidated when the environment provides the required evidence. Missing external runtime configuration is never replaced by mocks, placeholders or fabricated green status.
