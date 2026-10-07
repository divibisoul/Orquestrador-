
# SOUL Infrastructure Blueprint v1

## 1. Canonical platform

| Layer | Concrete choice |
|---|---|
| Containers | OCI/Docker |
| Production | Kubernetes |
| Discovery | Kubernetes Services + DNS |
| Application Mesh | canonical Soul Mesh only |
| Event bus | NATS JetStream |
| Secrets | HashiCorp Vault + External Secrets Operator |
| Network policy | Cilium NetworkPolicy |
| mTLS | workload identity/certificate management; sidecar or ambient implementation |
| Telemetry | OpenTelemetry Collector |
| Metrics | Prometheus |
| Dashboards | Grafana |
| Logs | structured JSON + centralized log sink |
| Traces | Tempo or equivalent OTLP trace backend |
| DB | PostgreSQL/Supabase |
| Object storage | Storacha current path + explicit legacy Web3 compatibility |
| Progressive delivery | Argo Rollouts |

Kubernetes service discovery is used instead of introducing Consul/etcd. A logical capability registry still exists as data; it is not a second transport or infrastructure Mesh.

## 2. Service catalogue

| Service | Repository | Language | Container port | Local host | Health | Readiness | Metrics |
|---|---|---|---:|---:|---|---|---|
| soul-n01 | divibisoul/aeternum-core-29 | TS/Node/Vite | 8080 | 8080 | /mesh/health | /mesh/health + deps | /metrics |
| soul-n02 | divibisoul/Eternium- | TS/Node | 3020* | 3020 | /mesh/health* | /ready* | /metrics* |
| soul-n03 | divibisoul/nexus-aeternum-fusion | TS/Node | 3030* | 3030 | /mesh/health* | /ready* | /metrics* |
| soul-n04 | divibisoul/nextjs-ai-chatbots | TS/Next | 3000 | 3040 | /api/health* | /api/ready* | /api/metrics* |
| soul-n05 | divibisoul/nextjs-ai-chatbot | TS/Next | 3000 | 3050 | /api/health* | /api/ready* | /api/metrics* |
| soul-n06 | divibisoul/nextjs-ai-chatbot-2000 | TS/Next | 3000 | 3060 | /api/health* | /api/ready* | /api/metrics* |
| soul-n07 | divibisoul/Orquestrador- | Go | 8080 | 8087 | /health | /ready | /metrics |
| sara | divibisoul/SARA | Python | 8090 | 8090 | /health | /health + deps | /metrics* |

Asterisk = target normalization surface. Adding a standard health/readiness route must not remove existing endpoints.

## 3. Canonical configuration

Version-controlled:
config/soul-config.v1.yaml
validated by:
contracts/soul-config.schema.json

The config contains:
- peer service URLs
- contract version
- timeouts
- payload limits
- retry and breaker values
- worker/bulkhead limits
- event bus/telemetry endpoints
- feature flags
- secret references

Secret values are never committed. Vault or an equivalent secret store injects them at runtime.

## 4. Canonical environment names

Common:
SOUL_MESH_SECRET
SOUL_MESH_HMAC_SECRET
SOUL_MESH_GATEWAY_HMAC_REQUIRED
SOUL_MESH_TOKEN
SOUL_MESH_N01_URL ... SOUL_MESH_N07_URL
SARA_SERVICE_URL
SARA_SERVICE_TOKEN
N07_APP_TOKEN
N07_HTTP_ADDR
N07_MAX_REQUEST_BYTES
N07_BACKEND_TIMEOUT

Observability:
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_SERVICE_NAME
OTEL_RESOURCE_ATTRIBUTES

Data:
DATABASE_URL
SUPABASE_URL
SUPABASE_SERVICE_ROLE_KEY
STORACHA_TOKEN

External providers remain nucleus-owned and retain their existing variables.

## 5. Secrets management

Vault is the authoritative production store.

Principles:
- one Kubernetes ServiceAccount identity per nucleus;
- least-privilege paths;
- short-lived credentials where provider supports it;
- dual-key overlap for Mesh-secret rotation;
- automatic renewal;
- no secrets in images, source or browser bundles.

Rotation procedure:
1. add new version;
2. deploy dual verification;
3. verify all peers;
4. switch signing;
5. verify E2E;
6. revoke old version.

Supabase service-role keys are server-only.

## 6. Network security

Default deny pod ingress/egress.

Allow-list:
- nucleus → declared nucleus peer Mesh
- nucleus → N07 federation
- N07/SARA → approved data stores
- all services → OTel collector
- Prometheus → metrics
- ingress/WAF → public applications

mTLS workload identity:
spiffe://soul/<namespace>/<service>

Capability authorization evaluates:
subject × source × capability × target × operation version.

Rate limit:
100 Mesh requests/min/source-target, burst 20, subject to capability overrides.

Payload:
2 MiB hard limit before JSON processing.

## 7. Resilience

Timeout: 30s.
Retry: 2 attempts after initial request; backoff 250ms and 750ms.
Breaker: 5 failures → open 60s → half-open probe.

Retry exclusions:
- destructive operations
- non-idempotent operations without idempotency key
- auth failures
- schema failures
- replay failures

Bulkhead:
- N07 global workers 32
- per-peer default 8
- per-capability default 8
- audio/provider pools isolated from control-plane pool

## 8. Event bus

NATS subjects:
soul.mesh.events.v1
soul.capability.events.v1
soul.execution.events.v1
soul.provenance.events.v1
soul.sara.events.v1
soul.alerts.v1

Use JetStream for durable replayable event streams. Consumers must be idempotent by eventId + correlationId.

The event bus is asynchronous infrastructure and never replaces the canonical Soul Mesh request/response contract.

## 9. Persistence

PostgreSQL/Supabase:
runs, artifacts, capability snapshots, deployment records, evidence indexes.

Object storage:
large artifacts, logs, release evidence, SBOM/attestation bundles.

SARA:
TemporalVectorDB, DecisionTrace and ProvenanceTracker remain source-of-truth audit records for regenerative execution.

Snapshots:
- database PITR/WAL + daily full backup
- object storage versioning
- SARA snapshot before regenerative/migration boundaries
- N07 execution metadata persisted before final acknowledgement

## 10. Observability

~~~mermaid
flowchart LR
  N[N01..N07 + SARA] --> OT[OTel Collector]
  N --> JS[NATS JetStream]
  OT --> P[Prometheus]
  OT --> G[Grafana]
  OT --> T[Tempo]
  OT --> L[Loki-or-log-sink]
  JS --> A[Audit Consumers]
  A --> DB[(PostgreSQL)]
  A --> S[SARA Trace Stores]
~~~

## 11. Kubernetes probes

Every workload has startup, liveness and readiness semantics.
Readiness must include required local dependencies. Liveness must detect deadlock, not normal dependency outage. Startup prevents premature liveness failure.

## 12. Production release

Every release must publish:
immutable OCI digest, SBOM, source revision, contract report, Mesh E2E evidence, failure injection report and provenance evidence.

Argo Rollouts:
10% → 25% → 50% → 100%.
Automatic rollback on contract/auth/readiness/E2E/error-rate/latency/provenance failures.
