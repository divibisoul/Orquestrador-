
# Soul Mesh Contract 1.1.0

## 1. Wire identity

Protocol: soul-mesh/1  
Contract: 1.1.0  
Encoding: UTF-8 JSON  
Transport: HTTPS/REST

## 2. Canonical envelope

~~~json
{
  "version": "1.0",
  "contractVersion": "1.1.0",
  "messageId": "uuid",
  "source": "N05",
  "target": "N06",
  "timestamp": 1790000000000,
  "nonce": "uuid",
  "correlationId": "uuid",
  "type": "CAPABILITY_REQUEST",
  "ttl": 1,
  "payload": {
    "capability": "inference.generate",
    "payload": {},
    "metadata": {}
  },
  "operation": "inference.generate@1.0.0",
  "metadata": {
    "traceparent": "00-..."
  },
  "hmac": "64hex"
}
~~~

## 3. Mandatory invariants

- correlationId stays constant across all hops of the logical request;
- messageId is fresh for every hop;
- nonce is fresh per signed message;
- source and target are explicit;
- timestamp is inside accepted clock skew;
- capability is present;
- payload <= 2 MiB;
- authorization happens before handler execution;
- HMAC is verified before trust;
- mTLS identity agrees with envelope source;
- replay is rejected.

## 4. Message types

PING, HEALTH, DISCOVERY_REQUEST, DISCOVERY_RESPONSE, CAPABILITY_REQUEST, TASK, TASK_RESULT, ERROR, COMBO_REQUEST, COMBO_RESULT, EVENT.

## 5. Result envelope

A result is another Mesh envelope with:
- fresh response messageId
- same correlationId
- source = executor
- target = requester
- capability = executed capability
- status = ok / partial / error / degraded / cancelled
- result or structured error
- evidenceState
- provenance references

## 6. Authentication

Production:
mTLS + HMAC-SHA256.
Local:
mTLS may be relaxed only for local profile, while CI uses authenticated mode.

HMAC source input is the canonical unsigned envelope representation defined by the N07 implementation. Any non-Go adapter must reproduce the exact field order/content or invoke a shared conformance test.

Minimum production secret: 32 bytes.

## 7. Discovery

GET /mesh/discovery:
- nucleus identity
- protocol and contract
- health/readiness
- transport
- agent registrations
- capability registrations
- evidence state
- revision
- TTL

POST /mesh/register:
- validates ownership and registration schema
- refreshes capability/agent state
- never transfers ownership

### 7.1 Complete peer reachability

The canonical peer set is N01, N02, N03, N04, N05, N06 and N07. Every nucleus declares all six other nuclei as peers; each source-to-target route is represented in both directions at the network level. A nucleus must never route to itself.

The `adjacentPeers` chain is a preferred composition/fusion order, not an access-control or reachability restriction. Cross-nucleus calls may route directly to any peer when that peer advertises the requested executable capability. Capability ownership remains with the destination nucleus; complete connectivity does not transfer execution authority.

The N07 topology endpoint publishes direct inbound and outbound channels for N01–N06 and reports the full graph as 42 logical directed routes. Each channel exposes only the shared bootstrap operation `mesh.describe`; N07 tries `mesh.describe` first and uses `mesh.discovery` only as a legacy fallback. Executable operations are discovered from the destination at runtime. N07's own declared operation catalog is published separately and is not copied onto peer channels. The `mesh.describe` payload exposes `declaredCapabilities` and `executableCapabilities` from registered handlers, while retaining `operations` as a compatibility alias; callers may execute only entries in `executableCapabilities`. For transition compatibility, N07 accepts a service-base URL or a complete `/api/soul-mesh` URL in `SOUL_MESH_N01_URL` … `SOUL_MESH_N06_URL` and appends the path only when absent. A channel claims only HTTP/HTTPS; other transports must not be inferred from the topology. This is a structural declaration only: live connectivity remains unverified until authenticated real-runtime probes succeed.

### 7.2 Public external repository fan-out

The N07 external-federation catalog reconciles 31 pinned public upstream repositories while preserving one canonical Mesh and one control plane:

- 22 adapter descriptors are in the current provider map (16 primary and 6 complementary). Their presence in the catalog does not by itself prove runtime readiness; adapter execution is disabled by default and separately gated.
- 9 additional sources are disclosed as `PROJECTED`: FedML, Hivemind, Temporal, Cognitive Workspace, RAVANA, Ray, Bijux DAG Runtime, Ouro Loop and Recurs.
- The canonical, provider-agnostic metadata routes are `external.federation.source.describe@1.0.0` and `external.federation.source.probe@1.0.0`; callers pass `{"provider":"ravana"}` as the ordinary Mesh payload. The gateway normalizes that into operation metadata, so consumers do not need a provider-specific adapter; legacy envelopes with `metadata.provider` remain supported. Both routes work for active and projected records and return the same identity/owner/pin/evidence shape. This is the recommended integration path for N01–N07.
- Existing per-source `external.<id>.describe@1.0.0` and `external.<id>.probe@1.0.0` aliases remain available for compatibility with projected records. They expose source-presence and exact-pin evidence through N07; they do not run external code or grant a runtime execution route.
- Federation discovery exposes the projected records and their pinned revision/evidence requirements, but the runtime route table excludes their native operations. A projected operation must fail closed rather than fall through to an unregistered runner.
- Every canonical nucleus N01–N07 is represented as an authorized consumer of the projected-source catalog, either as the declared primary owner or as a consumer. SARA remains a transversal authority, not an additional nucleus.
- Candidate repositories without a materialized/pinned submodule, and sources marked `BLOCKED`, are not counted as routeable Mesh peers.

Promotion gate for any projected source:
1. exact source URL, revision, Gitlink and worktree reconciliation;
2. native adapter contract and ownership/consumer validation;
3. deterministic adapter tests, including failure/timeout behavior;
4. authenticated Mesh conformance tests preserving correlation, provenance, HMAC/mTLS and destination ownership;
5. real runtime execution evidence on the exact CI head.

Catalog response status `PASS` means the catalog operation succeeded only. It must include `runtime_evidence: UNVERIFIED` until an independent runtime proof exists. A source is never marked executable merely because its public repository is reachable.

## 8. Error taxonomy

| Code | Meaning | Retry |
|---|---|---|
| MESH_CONTRACT_VERSION_MISMATCH | version conflict | no |
| MESH_AUTH_FAILED | authentication failure | no |
| MESH_REPLAY_DETECTED | nonce replay | no |
| MESH_PAYLOAD_TOO_LARGE | >2 MiB | no |
| CAPABILITY_FORBIDDEN | unauthorized consumer | no |
| CAPABILITY_UNAVAILABLE | owner unavailable | maybe |
| PEER_TIMEOUT | 30s exceeded | yes if idempotent |
| CIRCUIT_OPEN | protection active | later |
| CAPABILITY_EXECUTION_FAILED | native handler failed | policy |
| PARTIAL_COMPOSITION | combo partial result | policy |

## 9. Context and provenance

Handoff:
correlationId, trace context, messageId, parentMessageId, deadline, capability, capabilityVersion, provenanceRefs.

OpenTelemetry context is linked to Mesh correlation but does not replace business correlation.

## 10. mesh.combo

A combo contains:
comboId, correlationId, steps[], policy, input, aggregation, compensations[], provenancePolicy.

Linear v1 execution is deterministic.
A later DAG form must be versioned as a new contract.

## 11. Idempotency

Mutating operations accept:
payload.metadata.idempotencyKey

Scope:
source + target + capability + key.

## 12. Security ordering

deserialize size guard → transport identity → envelope decode → timestamp/replay → HMAC → source/target policy → capability authorization → breaker/deadline → handler → result schema → provenance → response.
