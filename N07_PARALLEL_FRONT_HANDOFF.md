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
- Web3 Storage is now a first-class N07 provider capability behind a transport-neutral adapter: `storage.web3.upload@1.0.0` and `storage.web3.status@1.0.0`.
- Web3 Storage credentials are runtime-only (`WEB3_STORAGE_TOKEN`); the provider is never considered connected merely because configuration exists.
- N07 exposes authenticated provider health, upload and CID-status endpoints and registers the storage capabilities with the Mesh.
- The provider adapter uses the documented raw `POST /upload` contract and enforces a bounded upload size.

## COORDINATION RULE
Before every mutation, reread current `main` and the exact target-file SHA. Concurrent fronts may advance `main` between any two operations. Never force-reset or overwrite a newer front's work.

## VALIDATION RULE
CI is evidence. A green format/vet/test/race/build result is required before declaring the current executable state validated. Structural readiness is not equivalent to live six-runtime commissioning or live provider authentication.

## REMAINING CROSS-FRONT WORK
- Commission live bidirectional Mesh with N01–N06 concurrently reachable.
- Reconcile the six peer response-signature implementations against the canonical N07 response envelope.
- Reconcile current N01–N06 capability/tool/agent heads before final fusion.
- Validate composed capabilities and distributed SuperGPU execution end-to-end with all peer runtimes available.
- Configure a real `WEB3_STORAGE_TOKEN` in the deployment secret store and run `/storage/web3/health` against the configured account.
- Perform one controlled upload and verify the returned CID through `/storage/web3/status` before declaring provider commissioning complete.
- Keep the Web3 Storage adapter replaceable because the provider has moved from the deprecated legacy API toward its newer UCAN/w3up architecture.

## CURRENT_EVIDENCE
- The repository is under concurrent mutation; current `main` must be reread before every subsequent change.
- Previous N07 CI evidence passed format check, vet, tests, race test and build after the federation fixes.
- Web3 Storage integration has contract coverage in `storage/web3storage/client_test.go`; live provider authentication and upload remain deployment-level evidence gates.
