# N07 live cross-front handoff

## CURRENT CONTRACT
N07 owns the canonical orchestration, neural-fabric, prefrontal and SuperGPU runtime surfaces. Peer nuclei expose adapters and remain independently deployable. Do not create a second N07 runtime or copy peer implementations into N07.

## WHAT_CHANGED
- Versioned operation identity is canonicalized as `name@semver` in the N07 registry and router.
- Explicit version requests route to the exact registration; unversioned requests resolve to the newest version.
- Route cache uses the same canonical identity.
- N01–N06 neural adapters target the same N07 neural capability surface and preserve correlation, Mesh contract, finite-number validation and bounded execution time.
- Neural federation exposes context-aware `ParallelContext` and `BroadcastContext` with deterministic result ordering and a fixed 32-worker ceiling.
- Neural parallel admission uses a fixed worker pool rather than one goroutine per submitted task.
- Parent cancellation propagates through federated peer invocation.
- SuperGPU parallel Mesh execution is bounded to 32 tasks and uses capability-aware dynamic peer routing.
- Dynamic routing evaluates executable capability availability across configured peers concurrently and scores candidates using health, latency and recent failures before invocation.
- Explicit neural `Target` remains authoritative; the historical `Source` routing hint remains compatible when `Target` is omitted.
- Storage is now provider-neutral at the N07 backend boundary. Current production mode is Storacha/w3up/UCAN through the Guppy Go client; the legacy Web3.Storage HTTP path is retained only as explicit compatibility mode.
- Storacha uploads use a bounded, private temporary staging directory, sanitize the supplied filename, preserve the configured space, and validate the returned root CID before exposing it.
- Storage HTTP requests now use the configured backend timeout, with an explicit `STORACHA_TIMEOUT` override.
- Storage tests cover the modern mode, upload size enforcement, CID validation, legacy HTTP behavior and cancellation.

## COORDINATION RULE
Before every mutation, reread current `main` and the exact target-file SHA. Concurrent fronts may advance `main` between any two operations. Never force-reset or overwrite a newer front's work.

## VALIDATION RULE
CI is evidence. A green format/vet/test/race/build result is required before declaring the current executable state validated. Structural readiness is not equivalent to live six-runtime commissioning or live provider authentication.

## STORAGE COMMISSIONING RULE
A real Storacha commissioning result requires an authorized Guppy agent with a persistent `STORACHA_DATA_DIR` and an assigned `STORACHA_SPACE`. No front may fabricate a CID, token, UCAN proof or provider success. When those deployment credentials are available, run one controlled N07 upload and verify the resulting CID through the N07 status/gateway path.

## CURRENT_EVIDENCE
- Base `main` was rechecked before this front's mutation and was at `4b946cd3c4acbdab4c7af90432b56d91a1beaf17`.
- Current storage implementation lives in `backend/storage.go`; the historical `storage/web3storage` client was removed by prior storage fronts and must not be recreated as a duplicate implementation.
- Current Docker image installs `github.com/storacha/guppy` and persists `/var/lib/n07/storacha` for provider state.
- Current PR: `#20` (`fix(n07): harden Storacha storage adapter runtime`).
- Live provider upload remains a deployment-level evidence gate because no authorized Storacha agent state is available through the GitHub source-control interface.
