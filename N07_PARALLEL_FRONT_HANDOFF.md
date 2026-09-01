# N07 live cross-front handoff

## WHAT_CHANGED
- Versioned operation identity is canonicalized as `name@semver` in the N07 registry and router.
- Explicit version requests route to the exact registration; unversioned requests resolve to the newest registered version.
- Route cache uses the same canonical identity, preventing stale or unreachable versioned routes.
- N01–N06 neural adapters target the same N07 neural capability surface and preserve `correlationId`, Mesh contract `1.1.0`, finite-number validation and bounded execution time.
- Neural federation now has context-aware `ParallelContext` and `BroadcastContext` paths with bounded admission (`32`) and deterministic result ordering; legacy APIs delegate to these paths.
- Federated execution preserves parent cancellation through the peer invocation boundary instead of creating detached background work.
- The cumulative cross-front execution queue is now recorded in `SOUL_EXECUTION_QUEUE.md` so concurrent fronts share the same 40-function closure order and evidence gates.

## WHAT_WAS_FOUND
- Previous N07 routing could make versioned registrations unreachable through the router.
- The neural federation model previously relied on loose routing semantics and needed canonical target handling.
- Parallel neural fan-out admitted an unbounded number of goroutines; the new admission bound prevents uncontrolled resource amplification.
- Legacy federation methods did not accept a caller context, which could detach cancellation from the parent orchestration request.
- Cross-front branches continue to evolve independently; current `main` must be reread before each subsequent mutation.

## WHAT_REMAINS
- Obtain exact-head CI evidence for the latest HEAD.
- Complete live bidirectional commissioning of all six peer adapters.
- Unify response-signature verification semantics across all six TypeScript neural adapters against the canonical N07 response envelope.
- Complete distributed SuperGPU integration tests while all peer runtimes are concurrently reachable.
- Reconcile N01–N06 capability/tool changes continuously before final fusion.

## WHAT_NEXT_AGENT_SHOULD_DO
- Preserve current operation ownership; do not fork another N07 neural runtime.
- Read current N07 `main` and current peer SHAs before changing files.
- Preserve concurrent front work and resolve SHA conflicts by rereading the changed file.
- Treat CI as evidence, not as a declaration.
- Use `SOUL_EXECUTION_QUEUE.md` as the cumulative task order and update it when executable evidence changes.

## CURRENT_SHA
- Neural federation hardening: `d5d6a8ecedbea189bb9b9ce8ca0e0ec2abd96b32`
- Neural regression coverage: `450a47711cf448ee01b8dad06f2bbf2333ffff99`
- Cross-front execution queue: `f9b230a562885b3c4b8a98ddfb6fc79530bdc039`
