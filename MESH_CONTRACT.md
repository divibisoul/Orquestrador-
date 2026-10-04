
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
