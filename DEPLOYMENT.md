
# SOUL Production Deployment Blueprint

## 1. Release artifact

A production release contains:
- source revision
- OCI image digest
- SBOM
- vulnerability scan
- contract test report
- Mesh E2E report
- failure-injection report
- provenance/evidence bundle
- migration checksum

## 2. Per-service Kubernetes objects

- Argo Rollout or Deployment
- Service
- ServiceAccount
- ConfigMap
- ExternalSecret
- NetworkPolicy
- PodDisruptionBudget
- ServiceMonitor
- PrometheusRule

SARA:
- PVC for trace/regenerative state
- snapshot job

N07:
- persistent state only where required by current storage adapter
- isolated compute worker pool when resource pressure justifies it

## 3. Startup order

~~~mermaid
flowchart LR
  I[Infra dependencies] --> S[SARA persistence]
  S --> N7[N07]
  N7 --> N1[N01]
  N1 --> N2[N02]
  N2 --> N3[N03]
  N3 --> N4[N04]
  N4 --> N5[N05]
  N5 --> N6[N06]
  N6 --> E[E2E]
~~~

Deployment start order is distinct from the user-mandated pair commissioning order.

## 4. Progressive delivery

Argo Rollouts:
10% → 25% → 50% → 100%.

Promotion requires:
readiness, contract checks, Mesh capability transaction, correlation preservation and provenance persistence.

Automatic rollback for:
contract mismatch, auth failure, readiness collapse, E2E failure, excessive errors/latency or provenance loss.

## 5. Rollback

Application rollback returns to the last immutable digest whose release evidence is valid.

Database migrations use expand/contract where rollback is required. Destructive migrations require forward recovery, not blind image rollback.

SARA snapshots precede state-schema changes.

## 6. RELEASE_GATES_V1 operationalization

The existing 15 release gates are authoritative.
The pipeline maps:
1–3 → contract/discovery E2E
4–5 → federation/ownership validation
6–8 → compute/resilience/observability
9 → backend/storage regression
10 → CI + production image
11 → seven-nucleus E2E
12 → Android contract
13 → fail-closed config validation
14 → immutable image + persistence
15 → database/client secret boundary

Final state = ONLINE only when every mandatory gate is verified.
