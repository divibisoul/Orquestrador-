# N07 live cross-front handoff

## CURRENT CONTRACT
N07 owns the canonical orchestration, neural-fabric, prefrontal and SuperGPU runtime surfaces. Peer nuclei expose adapters and remain independently deployable. Do not create a second N07 runtime or copy peer implementations into N07.

## WHAT_CHANGED
- Versioned operation identity is canonicalized as `name@semver` in the N07 registry and router.
- Explicit version requests route to the exact registration; unversioned requests resolve to the newest registered version.
- Route cache uses the same canonical identity.
- N01–N06 neural adapters target the same N07 neural capability surface and preserve correlation, Mesh contract, finite-number validation and bounded execution time.
- Neural federation exposes context-aware `ParallelContext` and `BroadcastContext` with deterministic result ordering and a fixed 32-worker ceiling.
- Neural parallel admission uses a fixed worker pool rather than one goroutine per submitted task.
- Parent cancellation propagates through federated peer invocation.
- SuperGPU parallel Mesh execution is bounded to 32 tasks and now uses capability-aware dynamic peer routing.
- Dynamic routing evaluates executable capability availability across configured peers concurrently and scores candidates using health, latency and recent failures before invocation.
- Explicit neural `Target` remains authoritative; the historical `Source` routing hint remains compatible when `Target` is omitted.

## COORDINATION RULE
Before every mutation, reread current `main` and the exact target-file SHA. Concurrent fronts may advance `main` between any two operations. Never force-reset or overwrite a newer front's work.

## VALIDATION RULE
CI is evidence. A green format/vet/test/race/build result is required before declaring the current executable state validated. Structural readiness is not equivalent to live six-runtime commissioning.

## REMAINING CROSS-FRONT WORK
- Commission live bidirectional Mesh with N01–N06 concurrently reachable.
- Reconcile the six peer response-signature implementations against the canonical N07 response envelope.
- Reconcile current N01–N06 capability/tool/agent heads before final fusion.
- Validate composed capabilities and distributed SuperGPU execution end-to-end with all peer runtimes available.

## CURRENT_EVIDENCE
- Current `main` is `70ec375553167349564b868e9826d1102e3353c6`.
- The latest N07 CI run for that HEAD completed format check, vet, tests, race test and build successfully.
- The preceding CI failures were diagnosed and corrected rather than bypassed: a Mesh test namespace collision, an over-strict built-in-operation test expectation, and an unintended federation target behavior change were addressed while preserving compatibility.
