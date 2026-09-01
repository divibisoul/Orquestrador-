# SOUL v1 — Release Gates

Release is finite. New features outside these gates do not extend v1.

## Mandatory gates

1. N01–N06 current heads reconciled against N07 contracts.
2. N07 current runtime and APIs compile and pass unit tests.
3. Mesh ingress/egress, discovery, delegation and correlation are validated.
4. Neural Federation N07↔N01..N06 uses one canonical payload/signature contract.
5. Capability, agent and tool ownership is explicit; declared capability is not treated as executable without proof.
6. SuperGPU/parallel execution is bounded, cancellable and aggregatable.
7. Resilience controls are active: timeout, retry/backoff, breaker, rate limit, payload bounds and recovery.
8. Observability exposes routing, peer health, latency, failures, in-flight and trace/correlation state.
9. Backend integrations (Supabase + Web3 Storage) have automated regression coverage and exact-head CI proof before merge.
10. CI on the release candidate is green: format, vet, test, race and build.
11. E2E N01↔N07, N02↔N07, N03↔N07, N04↔N07, N05↔N07, N06↔N07 are verified when runtimes are concurrently available.
12. Android integration contract is stable and documented for the downstream APK/app project.
13. Release candidate can start with required configuration and fail closed when secrets are missing.

## Completion states

STRUCTURAL = implementation and repository tests exist.

INTEGRATED = cross-front contracts and CI are green on the exact release candidate.

ONLINE = real runtimes are reachable together and the E2E smoke suite succeeds.

Only ONLINE is final v1 completion.
