
# SOUL Observability Blueprint

## 1. Trace root

SOUL business trace root = correlationId.
OpenTelemetry spans use process/span context underneath it.

Required span attributes:
soul.correlation_id
soul.message_id
soul.parent_message_id
soul.source
soul.target
soul.capability
soul.operation
soul.contract_version
soul.evidence_state
soul.route_revision
soul.provider_revision
soul.retry_count
soul.breaker_state

Never put secrets or uncontrolled user payloads into span attributes.

## 2. Required metrics

Counters:
soul_mesh_requests_total
soul_mesh_failures_total
soul_mesh_retries_total
soul_mesh_circuit_open_total
soul_provenance_events_total
soul_sara_cycles_total

Histograms:
soul_mesh_request_duration_seconds
soul_combo_duration_seconds
soul_sara_cycle_duration_seconds

Gauges:
soul_mesh_inflight
soul_capability_available
soul_capability_route_score
soul_worker_pool_inflight
soul_evidence_state

High-cardinality IDs stay in traces/logs, not metric labels.

## 3. Structured logs

Each log event contains:
timestamp, level, service, event, correlationId, messageId, source, target, capability, status, durationMs, evidenceState and revision.

Payload text, credentials and raw prompts are redacted unless explicitly classified safe.

## 4. Dashboards

1. SOUL Overview
2. Mesh Topology
3. Capability Router
4. SuperGPU/SuperCompute
5. SARA Cycle
6. Provenance/Evidence
7. Release Gates
8. Storage Health
9. Security/Auth/Replays
10. Capacity/Cost

## 5. Alerts

Critical:
- Mesh auth failure spike
- replay surge
- circuit open > 60s
- error rate > 5% for 5m
- p95 latency regression > 5s sustained
- worker saturation > 90% for 5m
- provenance missing on executable result
- SARA trace persistence failure
- canary Mesh smoke failure
- storage regression for ONLINE candidate

Warning:
- readiness flapping
- capability TTL churn
- high retry rate
- provider degraded state
- route-score oscillation

## 6. Evidence query

correlationId → trace spans → Mesh hop IDs → SARA trace references → source revisions → CI runs → deployment digest.

This reconstructs a complete evidence chain without fabricating missing spans.

## 7. SARA trace linkage

Every SARA cycle stores:
cycle_id, correlationId, phase, DecisionTrace reference/hash, TemporalVectorDB references, ProvenanceTracker references, rollback flag, source revision, output fingerprint and evidence state.
