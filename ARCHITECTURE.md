
# SOUL — Complete Architecture Blueprint v1

**Scope:** N01–N07 + SARA.  
**Status:** SPECIFICATION / IMPLEMENTATION BASELINE.  
**Evidence rule:** repository structure is never treated as runtime proof.

## 0. Non-negotiable invariants

1. Exactly one canonical Soul Mesh application protocol exists.
2. N01–N06 remain independent runtimes and retain native ownership.
3. N07 is the canonical federation/control plane for discovery, routing, composition and SuperCompute/SuperGPU.
4. SARA remains authoritative for regenerative/governance/provenance operations.
5. Existing functionality, endpoints, modules and coordination documents are additive inputs; none are removed by this blueprint.
6. Production Mesh traffic requires mTLS and HMAC-SHA256 unless a formally reviewed compatibility adapter is used.
7. Runtime truth outranks repository presence.
8. No PASS/ONLINE state may be asserted without the evidence ladder.

## 1. Logical planes

- **Canonical Mesh plane:** synchronous capability calls.
- **Event plane:** NATS JetStream for durable asynchronous events, not a second request Mesh.
- **Trust plane:** Kubernetes network policy + mTLS + application HMAC.
- **Data/provenance plane:** PostgreSQL/Supabase, object storage and SARA trace stores.
- **Compute plane:** N07 scheduler/router plus bounded workers.

## 2. Six-core topology

N01–N06 create C(6,2)=15 bidirectional pairs and 30 directed logical links.

| Pair | Primary relationship |
|---|---|
| N01↔N02 | gateway/state ↔ multi-agent runtime |
| N01↔N03 | gateway/context ↔ audio/perception |
| N01↔N04 | coordination ↔ tools/documents |
| N01↔N05 | gateway ↔ inference/conversation |
| N01↔N06 | gateway ↔ cognition |
| N02↔N03 | agent context ↔ audio |
| N02↔N04 | agents ↔ tools/documents |
| N02↔N05 | agent execution ↔ inference/conversation |
| N02↔N06 | agents ↔ cognition |
| N03↔N04 | perception ↔ tools/documents |
| N03↔N05 | audio context ↔ conversation/inference |
| N03↔N06 | multimodal perception ↔ cognition |
| N04↔N05 | tools/documents ↔ conversation/inference |
| N04↔N06 | tools ↔ cognition |
| N05↔N06 | inference/conversation ↔ cognition |

~~~mermaid
graph TD
  N01[N01 Gateway]
  N02[N02 Multi-Agent]
  N03[N03 Audio]
  N04[N04 Tools-Docs]
  N05[N05 Conversation-Inference]
  N06[N06 Cognition]
  N01 <--> N02
  N01 <--> N03
  N01 <--> N04
  N01 <--> N05
  N01 <--> N06
  N02 <--> N03
  N02 <--> N04
  N02 <--> N05
  N02 <--> N06
  N03 <--> N04
  N03 <--> N05
  N03 <--> N06
  N04 <--> N05
  N04 <--> N06
  N05 <--> N06
~~~

## 3. N07 and SARA federation

N07 connects bidirectionally to N01–N06 using the same canonical Mesh semantics. These are federation/control links and do not change the six-core 30-link count.

~~~mermaid
graph LR
  N01 <--> N07
  N02 <--> N07
  N03 <--> N07
  N04 <--> N07
  N05 <--> N07
  N06 <--> N07
  N07 <--> SARA
~~~

Preferred production regenerative path: Nucleus → N07 federation → SARA. Direct Nucleus↔SARA compatibility clients are allowed where already implemented but do not become a second authority.

## 4. Canonical request profile

Transport: HTTPS/REST + JSON.

Existing compatibility surfaces:
- N01: POST /mesh/in
- N01 discovery: GET /mesh/discovery
- N01 health: GET /mesh/health
- N04/N05/N06 current application adapters: POST /api/soul-mesh
- legacy routes remain until a verified compatibility migration is complete

Semantic envelope fields:
version, contractVersion, messageId, correlationId, nonce, timestamp, source, target, type, capability, payload, metadata, operation, ttl, hmac.

For the existing 1.1.0 N07 wire form, capability is semantically mandatory and is serialized as payload.capability; business data is payload.payload. This avoids breaking the current N07 codec while making capability unambiguous.

## 5. All 30 core directed links

Common profile P1 unless noted:
- HTTPS/REST JSON
- mTLS + HMAC-SHA256 in production
- capability-based authorization
- timeout 30s
- 2 retries with exponential backoff 250ms / 750ms
- retry only idempotent operations or explicit idempotency key
- circuit breaker 5 eligible failures → open 60s → one half-open probe
- correlationId preserved
- fresh messageId + nonce per hop
- provenance references carried forward
- payload limit 2 MiB

| # | Directed link | Target service port | Capability authority |
|---:|---|---:|---|
| 1 | N01→N02 | 3020 | N02 |
| 2 | N02→N01 | 8080 | N01 |
| 3 | N01→N03 | 3030 | N03 |
| 4 | N03→N01 | 8080 | N01 |
| 5 | N01→N04 | 3040 host / 3000 svc | N04 |
| 6 | N04→N01 | 8080 | N01 |
| 7 | N01→N05 | 3050 host / 3000 svc | N05 |
| 8 | N05→N01 | 8080 | N01 |
| 9 | N01→N06 | 3060 host / 3000 svc | N06 |
| 10 | N06→N01 | 8080 | N01 |
| 11 | N02→N03 | 3030 | N03 |
| 12 | N03→N02 | 3020 | N02 |
| 13 | N02→N04 | 3040 host / 3000 svc | N04 |
| 14 | N04→N02 | 3020 | N02 |
| 15 | N02→N05 | 3050 host / 3000 svc | N05 |
| 16 | N05→N02 | 3020 | N02 |
| 17 | N02→N06 | 3060 host / 3000 svc | N06 |
| 18 | N06→N02 | 3020 | N02 |
| 19 | N03→N04 | 3040 host / 3000 svc | N04 |
| 20 | N04→N03 | 3030 | N03 |
| 21 | N03→N05 | 3050 host / 3000 svc | N05 |
| 22 | N05→N03 | 3030 | N03 |
| 23 | N03→N06 | 3060 host / 3000 svc | N06 |
| 24 | N06→N03 | 3030 | N03 |
| 25 | N04→N05 | 3050 host / 3000 svc | N05 |
| 26 | N05→N04 | 3040 host / 3000 svc | N04 |
| 27 | N04→N06 | 3060 host / 3000 svc | N06 |
| 28 | N06→N04 | 3040 host / 3000 svc | N04 |
| 29 | N05→N06 | 3060 host / 3000 svc | N06 |
| 30 | N06→N05 | 3050 host / 3000 svc | N05 |

**Port status:** N01 and N07 currently advertise :8080 in their configuration. N04–N06 use Next.js defaults in their repositories; local host ports 3040/3050/3060 map to service port 3000. N02/N03 target ports above are blueprint deployment values and must be verified against their actual listener before launch-command changes.

## 6. Capability ownership

| Family | Owner |
|---|---|
| inference.* | N05 |
| conversation.* | N05 |
| document.* | N04 |
| tool.* | N04 |
| audio.* | N03 |
| cognition.* | N06 |
| agent.* | N02/N06 according to exact capability |
| discovery / routing / composition / supergpu.* | N07 |
| sara.* | SARA |
| Android/native host functions | N01 |

SOUL-25 upstream repositories are providers/adapters, not owners.

## 7. Capability discovery

A capability registration contains:
id, version, owner, status, input schema, output schema, transport, evidence state, limits, dependencies and provenance.

Status lifecycle:
DECLARED → CONFIGURED → AVAILABLE → EXECUTABLE, with DEGRADED or BLOCKED as terminal non-PASS conditions.

A record is routable as executable only when owner resolution, handler/provider health and evidence rules all pass.

## 8. Agent discovery

An agent registration contains:
id, nucleus, role, lifecycle, capabilities, tools, resource limits, evidence state, revision and provenance.

Agent registration never transfers ownership between nuclei.

Defaults:
- capability TTL 60s
- agent heartbeat 15s
- stale after 3 missed heartbeats
- route cache 30s
- negative cache 5s

## 9. mesh.combo

Version 1 is an explicit ordered list, not an implicit graph.

Execution:
1. validate all steps before execution;
2. preserve originating correlationId;
3. mint fresh messageId and nonce for each hop;
4. resolve step input from prior result;
5. record provenance per hop;
6. apply timeout/retry/breaker;
7. stop, continue or compensate according to policy;
8. execute compensations in reverse order for atomic sections;
9. aggregate deterministically;
10. validate final output;
11. persist trace.

Policies:
- fail-fast
- best-effort
- compensate
- quorum

A failed or degraded hop is never converted into PASS.

## 10. SuperCompute / SuperGPU

~~~mermaid
flowchart LR
  A[TASK] --> B[DECOMPOSITION]
  B --> C[SCHEDULER]
  C --> D[CAPABILITY ROUTER]
  D --> E[PARALLEL EXECUTION]
  E --> F[AGGREGATION]
  F --> G[VALIDATION]
  G --> H[PROVENANCE]
  H --> I[RESULT]
~~~

N07 is the control-plane scheduler. Native capability execution remains in the owner nucleus. Global parallel admission defaults to 32 tasks; per-peer and per-capability bulkheads are smaller. Every task is cancellable and dependency-aware.

Router hard filters:
authorization → contract compatibility → provider availability → breaker → deadline feasibility.

Score factors:
health 40%, latency 25%, recent failure inverse 20%, capacity headroom 10%, priority fit 5%.

## 11. Correlation and provenance

Root request creates or adopts correlationId.
Every hop:
- preserves correlationId;
- creates fresh messageId;
- creates fresh nonce;
- records parentMessageId;
- carries capability version and provenance references.

W3C trace context may be bridged to OpenTelemetry. External trace/correlation headers are never trusted before authentication.

Provenance fields:
sourceRevision, adapterId, capabilityVersion, evidenceId, parentEventId.

## 12. Current evidence snapshot

Inspected 2026-10-03:
- N01 current head has successful recent Mesh/validation workflows.
- N02 current head has successful recent Mesh/integration workflows.
- N04, N05 and N06 current heads have successful recent Mesh/integration workflows.
- SARA current head has successful validation/integration workflows.
- N03 currently has a real provider-boundary test failure and an npm EOVERRIDE build failure.
- N07 currently has successful core integration/capability/SuperAGI checks but failing Pages and external-ownership workflows.
- No simultaneous seven-runtime live E2E proof was found.

Therefore the release state remains below ONLINE.

## 13. Commissioned definition

Commissioned means all of:
discovery + capability registration + authenticated delegation + native execution + correlation integrity + injected failure + recovery + composition + provenance.

A 200 response, health endpoint, compile or test suite alone is not commissioning evidence.
