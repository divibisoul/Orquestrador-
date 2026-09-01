# SOUL — 30-task cumulative execution checklist

This is the finite execution queue distilled from the latest cumulative SOUL directives. A task is complete only with executable implementation and evidence; structural completion is not runtime proof. Re-read current peer heads before each mutation.

| # | Task | Closure test | Current state |
|---:|---|---|---|
| 1 | Audit current GitHub state | current HEAD/branches/PRs inspected | DONE |
| 2 | Preserve concurrent work | no blind overwrite; SHA-checked writes | DONE |
| 3 | Reconcile N01–N06 current fronts | latest neural/Mesh commits mapped | DONE |
| 4 | Keep N07 active in parallel | N07 main/front remains active | DONE |
| 5 | Canonical Mesh contract | soul-mesh/1 + contract 1.1.0 aligned | DONE-STRUCTURAL |
| 6 | N07 six-peer topology | N01..N06 identity + channels exposed | DONE-STRUCTURAL |
| 7 | Capability discovery | live discovery + version normalization | DONE-STRUCTURAL |
| 8 | Intelligent routing | capability + health + latency + failures | DONE-STRUCTURAL |
| 9 | Delegation | local-first, remote fallback, correlation | DONE-STRUCTURAL |
| 10 | Hybrid transports | in-process/HTTP/realtime/event adapter boundary | DONE-STRUCTURAL |
| 11 | Resilience | timeout/retry/backoff/circuit/fallback | DONE-STRUCTURAL |
| 12 | Security | HMAC/replay/skew/TTL/payload validation | DONE-STRUCTURAL |
| 13 | Neural federation | N07 neural surface consumed by N01..N06 | DONE-STRUCTURAL |
| 14 | Neural worker bound | fixed worker pool and cancellation | DONE-STRUCTURAL |
| 15 | Agent identity | N07 discovery/router/executor/composer/validator/observer roles | DONE-STRUCTURAL |
| 16 | Capability composition | component provenance + dependency-aware execution | DONE-STRUCTURAL |
| 17 | Federated SuperGPU | decomposition/routing/parallel/aggregate/validation path | DONE-STRUCTURAL |
| 18 | Observability | peer health, latency, load, failures, traces, metrics | DONE-STRUCTURAL |
| 19 | Distributed discovery cache | bounded TTL + failure invalidation | DONE-STRUCTURAL |
| 20 | Runtime lifecycle | context-aware startup/shutdown/cancellation | DONE-STRUCTURAL |
| 21 | 40 public N07 functions | executable implementations + regression coverage | DONE-STRUCTURAL |
| 22 | Function/tool/agent fusion | ownership-safe composition across peers | INTEGRATION |
| 23 | N01×N06×N07 fusion | live ownership + gateway + composed path | INTEGRATION |
| 24 | Neural bridge response symmetry | all six adapters verify canonical response | INTEGRATION |
| 25 | Exact-head CI | format/vet/test/race/build green | PENDING-EVIDENCE |
| 26 | Multi-runtime Mesh E2E | N01..N06 concurrently reachable | PENDING-ENVIRONMENT |
| 27 | Neural E2E | forward/learn/deadline/error across peers | PENDING-ENVIRONMENT |
| 28 | SuperGPU E2E | concurrency/partial failure/recovery metrics | PENDING-ENVIRONMENT |
| 29 | Android integration handoff | stable N07 contract consumed by Android/Google Studio | READY-FOR-INTEGRATION |
| 30 | v1 release | final smoke + package/deploy prerequisites documented | BLOCKED-ONLY-BY-25..28 |

## Release rule

No new v1 feature is added after task 30 unless it is required to fix a release-blocking failure. Enhancements become v2 work after the v1 gate closes.
