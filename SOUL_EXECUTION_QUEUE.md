# SOUL cumulative execution queue

This queue is finite for v1. A task is complete only with executable code and validation evidence. It is a closure queue, not an open-ended backlog.

## Current priority
1. Re-read current N01–N07 heads before every mutation.
2. Preserve peer ownership; N07 owns orchestration/neural federation and peers expose adapters.
3. Reconcile protocol, correlation, authentication, timeout/cancellation, capability ownership and tool/agent contracts before fusion.
4. Close one complete task before starting the next; a local failure triggers correction and revalidation, never an indefinite expansion.

## Finite closure gates

| # | Task | Completion evidence | Status |
|---:|---|---|---|
| 1 | Core 40-function runtime | code paths + package tests | VERIFIED STRUCTURAL |
| 2 | Mesh contracts | canonical envelope + discovery/delegation tests | VERIFIED STRUCTURAL |
| 3 | Neural federation | N01..N06 adapters + N07 federation tests | VERIFIED STRUCTURAL |
| 4 | SuperGPU | bounded parallel + cancellation + aggregation | VERIFIED STRUCTURAL |
| 5 | Resilience/security | HMAC + TTL + retry/breaker + bounds | VERIFIED STRUCTURAL |
| 6 | Observability | peer/routing/trace metrics | VERIFIED STRUCTURAL |
| 7 | Backend storage | Supabase + Web3 Storage regression suite | CI GREEN ON PR #18 BRANCH |
| 8 | Backend integration | retarget to release main + exact-head CI | IN EXECUTION |
| 9 | Cross-nucleus contract reconciliation | current N01..N06 response signatures/health | PENDING |
| 10 | Six-peer E2E | all six peers concurrently reachable | PENDING ENVIRONMENT |
| 11 | Android integration handoff | stable N07 API contract | VERIFIED STRUCTURAL |
| 12 | Release smoke/deploy | configured startup + online smoke | PENDING ENVIRONMENT |

## Closure rule
A failed subtask produces a correction and a new validation cycle; it never resets completed gates and never expands v1 indefinitely. When a gate has executable evidence, mark it closed and move immediately to the next finite gate.

## Current evidence
- Backend/storage regression branch passed Format, Vet, Test, Race Test and Build in GitHub Actions run #310 / verify job.
- Backend/storage branch is intentionally separate from main until current-main integration is validated.
- N07 remains the single orchestration/neural-federation runtime; peers remain independent runtimes.
- Structural completion is distinct from online completion.
