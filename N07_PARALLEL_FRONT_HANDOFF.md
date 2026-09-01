# N07 live cross-front handoff

## WHAT_CHANGED
- Versioned operation identity is canonicalized as `name@semver` in the N07 registry and router.
- Explicit version requests route to the exact registration; unversioned requests resolve to the newest registered version.
- Route cache uses the same canonical identity, preventing stale or unreachable versioned routes.
- N01–N06 neural adapters target the same N07 neural capability surface and use `correlationId`, contract `1.1.0`, finite-number validation and bounded execution time.
- Neural federation now has an explicit `Target` routing field, finite-payload validation, deterministic empty-task behavior, deadline enforcement, and backward compatibility for the historical `Source` routing hint.
- Regression tests cover routing, correlation propagation, non-finite payload rejection, explicit/legacy routing, empty parallel batches and deadlines.

## WHAT_WAS_FOUND
- Previous N07 routing stored registered operations under bare names while incoming versioned operations carried `@semver`, making some built-ins unreachable through the router.
- The live N07 gateway and protocol currently accept N01–N07 identities and use the canonical Mesh contract.
- The neural federation task model lacked a canonical destination field and silently relied on `Source` as a routing hint.
- Cross-front work is still changing the repositories, so `main` must be reread before any subsequent modification.

## WHAT_REMAINS
- Obtain CI evidence for the new N07 HEAD; no status is currently returned for the tested HEADs.
- Complete live bidirectional commissioning of all six peer adapters.
- Unify response-signature verification semantics across all six TypeScript neural adapters against the N07 canonical response envelope.
- Complete distributed Super GPU runtime tests when all peer runtimes are concurrently reachable.
- Reconcile N01–N06 capability/tool changes continuously before final fusion.

## WHAT_NEXT_AGENT_SHOULD_DO
- Preserve current operation ownership and do not fork another N07 neural runtime.
- Read current N07 `main` and current peer SHAs before changing files.
- Reconcile any concurrent changes with this handoff rather than overwriting them.
- Treat CI as evidence: green code changes still require integrated runtime commissioning before claiming completion.

## CURRENT_SHA
- Neural federation hardening: `bb11e49cc1619f406310b841b048e672cc7bf147`
- Regression tests: `c6b1bd689500d6218d94f1b502f2794733b372d0`
- Handoff update: this commit.
