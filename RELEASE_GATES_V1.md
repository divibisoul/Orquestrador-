# SOUL v1 — Release Gates

Release is finite. New features outside these gates do not extend v1.

## Mandatory gates

1. N01–N07 current heads reconciled against the canonical SOUL contract.
2. N07 current runtime and APIs compile and pass unit/regression tests.
3. Mesh ingress/egress, discovery, delegation, authorization and correlation are validated.
4. Neural Federation N07↔N01..N06 uses the canonical `soul-mesh/1` / `1.1.0` message contract or an explicitly tested compatibility adapter.
5. Capability, agent and tool ownership is explicit; declared capability is not treated as executable without proof.
6. SuperGPU/parallel execution is bounded, cancellable, dependency-aware and aggregatable.
7. Resilience controls are active: timeout, retry/backoff, breaker, rate limit, payload bounds and recovery.
8. Observability exposes routing, peer health, latency, failures, in-flight work and trace/correlation state.
9. Backend integrations (Supabase + current Storacha path, with explicit legacy compatibility) have automated regression coverage and exact-head CI proof.
10. CI on the release candidate is green: format, vet, test, race, build and production-container validation.
11. A seven-nucleus E2E smoke suite executes at least one non-ping native capability through the orchestrated path and verifies source, target and correlation on the returned result.
12. Android/device integration contract is stable and documented for the downstream APK/app project.
13. Release candidate starts with required configuration and fails closed when authentication or required storage credentials are missing.
14. Production deployment publishes an immutable image digest and keeps the Storacha agent state persistent.
15. Supabase schema is reproducible from repository migrations and service-role credentials are never shipped to client code.

## Completion states

STRUCTURAL = implementation and repository tests exist.

VALIDATED = exact release candidate passes automated verification and production image build.

INTEGRATED = deployed nucleus endpoints exchange a real native capability transaction with preserved correlation.

ONLINE = production deployment is reachable, health/ready checks pass, storage round-trip succeeds and the seven-nucleus E2E smoke suite passes.

Only ONLINE is final v1 completion.
