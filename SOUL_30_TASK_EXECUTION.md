# SOUL — 30-task cumulative execution checklist

This queue is the finite v1 engineering scope distilled from the cumulative SOUL directives. A task is complete only with executable implementation and evidence. Re-read current peer heads before every mutation because N01–N06 and N07 evolve concurrently.

| # | Task | Closure test | Current state |
|---:|---|---|---|
| 1 | Audit current GitHub state | current HEAD/branches/PRs inspected | DONE |
| 2 | Preserve concurrent work | SHA-checked reads/writes; no blind overwrite | DONE |
| 3 | Reconcile N01–N06 current fronts | latest Mesh/neural commits mapped | DONE |
| 4 | Keep N07 active in parallel | N07 main/front remains active | DONE |
| 5 | Canonical Mesh contract | soul-mesh/1 + contract 1.1.0 alignment | DONE-STRUCTURAL |
| 6 | N07 six-peer topology | N01..N06 identity + bidirectional channels | DONE-STRUCTURAL |
| 7 | Capability discovery | canonical live discovery + version handling | DONE-STRUCTURAL |
| 8 | Intelligent routing | executable capability + health/latency/failure evidence | DONE-STRUCTURAL |
| 9 | Delegation | local-first + remote fallback + correlation | DONE-STRUCTURAL |
| 10 | Hybrid transports | adapter boundary for local/HTTP/realtime/event modes | DONE-STRUCTURAL |
| 11 | Resilience | timeout/retry/backoff/circuit/fallback | DONE-STRUCTURAL |
| 12 | Security | HMAC/replay/skew/TTL/payload controls | DONE-STRUCTURAL |
| 13 | Neural federation | single N07 neural surface consumed by N01..N06 | DONE-CODED |
| 14 | Neural worker bound | fixed worker-pool ceiling + cancellation | DONE-CODED |
| 15 | Agent identity | discovery/router/executor/composer/validator/observer roles | DONE-STRUCTURAL |
| 16 | Capability composition | dependency-aware composition + provenance | DONE-STRUCTURAL |
| 17 | Federated SuperGPU | bounded parallel fan-out + aggregation + required failures | DONE-STRUCTURAL |
| 18 | Observability | latency/load/failure/trace/metrics surfaces | DONE-STRUCTURAL |
| 19 | Discovery cache | TTL cache + invalidation on peer failure | DONE-CODED |
| 20 | Runtime lifecycle | context-aware shutdown/cancel | DONE-STRUCTURAL |
| 21 | 40 public N07 functions | implementations + regression coverage | DONE-STRUCTURAL |
| 22 | Function/tool/agent fusion | ownership-safe composed execution path | DONE-STRUCTURAL; LIVE E2E PENDING |
| 23 | N01×N06×N07 fusion | ownership + gateway + composed route | DONE-STRUCTURAL; LIVE E2E PENDING |
| 24 | Neural response symmetry | all six adapters align request/response authentication | DONE-CODED; LIVE E2E PENDING |
| 25 | Exact-head CI | format/vet/test/race/build green on exact release HEAD | FORMAT GREEN; VET NEXT GATE |
| 26 | Multi-runtime Mesh E2E | N01..N06 reachable/configured concurrently | BLOCKED-ENVIRONMENT |
| 27 | Neural E2E | forward/learn/deadline/error through all peers | BLOCKED-ENVIRONMENT |
| 28 | SuperGPU E2E | concurrency/partial-failure/recovery measured on live peers | BLOCKED-ENVIRONMENT |
| 29 | Android integration handoff | documented stable HTTP/Mesh contract for Android/Studio | READY-FOR-INTEGRATION |
| 30 | v1 release | final smoke + package/deploy prerequisites | BLOCKED-25..28 |

## Evidence ledger

- `bb11e49cc1619f406310b841b048e672cc7bf147`: N07 neural federation hardening; explicit Target, finite-value validation, deadlines and bounded parallelism.
- `c6b1bd689500d6218d94f1b502f2794733b372d0`: neural federation regression coverage.
- `d42c52cdb9e4e313f11b175e8d65cece3a6d92dc`: neural network formatting normalization.
- `42e34572ad1a7636dff5b2fa461d76774a247479`: Release Doctor HMAC readiness fix.
- `63dd1e6427aa818bf694fff3f1a73d95ba479261`: peer helper collision fix; followed by automatic formatting commit `e764d2c3553983b00e4a7f64109c0edac49d8075`.
- `5f28c81d3f5194a9f602171a66cb496cf9802960`: Android integration contract.
- `cmd/release-doctor`: six-peer readiness probe.
- Current CI run `33522430756` demonstrated `Format check = success`, then exposed executable integration defects at `Vet`; fixes were applied and formatting normalized again.

## Release rule

No new v1 feature is added after task 30 unless required to resolve a release-blocking failure. Post-release improvements become v2 work.
