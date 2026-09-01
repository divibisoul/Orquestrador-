# N07 live cross-front handoff

## WHAT_CHANGED
- Versioned operation identity is canonicalized as `name@semver` in the N07 registry and router.
- Explicit version requests route to the exact registration; unversioned requests resolve to the newest registered version.
- Route cache uses the same canonical identity, preventing stale or unreachable versioned routes.
- N01–N06 neural adapters target the same N07 neural capability surface and preserve `correlationId`, Mesh contract `1.1.0`, finite-number validation and bounded execution time.
- Neural federation uses an explicit target route, deadline enforcement and deterministic validation.
- Federated SuperGPU execution is now bounded to 32 concurrent tasks, rejects oversized task sets, derives per-task child correlations, and enforces a 15-second default task timeout with cancellation propagation.
- Regression coverage includes route/version behavior, neural cache behavior, non-finite input rejection, Mesh capability discovery and parallel-execution invariants.

## WHAT_WAS_FOUND
- Previous N07 routing could make versioned registrations unreachable through the router.
- The neural federation model previously relied on loose routing semantics and needed canonical target handling.
- The enhanced federated gateway previously started an unbounded goroutine for every submitted task. That could exhaust process resources under a large request even though the request body itself was bounded.
- Cross-front branches continue to evolve independently; current `main` must be reread before each subsequent mutation.

## WHAT_REMAINS
- Re-run and obtain CI evidence for the latest HEAD after this hardening.
- Complete live bidirectional commissioning of all six peer adapters.
- Unify response-signature verification semantics across all six TypeScript neural adapters against the canonical N07 response envelope.
- Complete distributed SuperGPU integration tests while all peer runtimes are concurrently reachable.
- Reconcile N01–N06 capability/tool changes continuously before final fusion.

## WHAT_NEXT_AGENT_SHOULD_DO
- Preserve current operation ownership; do not fork another N07 neural runtime.
- Read current N07 `main` and current peer SHAs before changing files.
- Preserve concurrent front work and resolve SHA conflicts by rereading the changed file.
- Treat CI as evidence, not as a declaration.

## CURRENT_SHA
- Bounded parallel gateway: `bf54ed894ddf62e460628317363055cacf589926`
- Regression test update: `4fef034ee1ae0b0cd1f380afcb53e64f5a349f3b`
- Handoff: this commit.
