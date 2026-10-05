
# SOUL E2E Commissioning Plan

## 1. Evidence ladder

L1 live E2E
L2 integration test
L3 unit/build
L4 contract/schema
L5 static inspection

Higher gates cannot be satisfied by lower evidence.

## 2. Exact commissioning order

### Stage 1 — N06↔N05
Discovery, capability registration, authenticated transaction in both directions, correlation preservation, fresh message IDs, native capability execution, induced timeout, breaker recovery, combo and provenance.

### Stage 2 — N05↔N04
Add tool/document authorization denial, payload limit and idempotent retry.

### Stage 3 — N04↔N03
Add real provider boundary checks and degraded-provider behavior.

### Stage 4 — N03↔N02
Add multimodal context transfer and bounded payload test.

### Stage 5 — N02↔N01
Add registration/discovery, source/target authorization and replay attack.

### Stage 6 — N01↔N06↔N07 + SARA
Require N01↔N07, N06↔N07 and N07↔SARA, seven-nucleus discovery, native non-ping capability, SuperGPU bounded parallel execution, deterministic aggregation, cancellation, partial failure and SARA 12-phase trace.

## 3. Failure injection

| Injection | Expected evidence |
|---|---|
| 35s downstream delay | timeout, retry, explicit failure |
| 5 eligible failures | breaker OPEN |
| 61s after open | half-open probe |
| duplicate nonce | replay reject |
| old timestamp | skew reject |
| >2 MiB body | reject before execution |
| unauthorized capability | semantic forbidden |
| provider absent | DEGRADED/BLOCKED |
| partial combo step | partial result according to policy |
| cancellation | downstream cancellation |
| lost trace persistence | FAIL/BLOCKED, never converged |

## 4. Combo E2E

~~~mermaid
sequenceDiagram
  participant C as Caller
  participant N07 as N07
  participant N05 as N05
  participant N04 as N04
  participant N03 as N03
  participant S as SARA
  C->>N07: combo correlation C1
  N07->>N05: inference hop M1 C1
  N05-->>N07: result M2 C1
  N07->>N04: tool hop M3 C1
  N04-->>N07: result M4 C1
  N07->>N03: audio hop M5 C1
  N03-->>N07: result M6 C1
  N07->>S: trace hop M7 C1
  S-->>N07: trace M8 C1
  N07-->>C: validated result M9 C1
~~~

## 5. Handoff artifact

Each stage emits a JSON handoff containing:
stage, exact revisions, endpoints, capabilities exercised, test results, failure injections, evidence state, blockers and rollback point.

## 6. Commissioned

A link is commissioned only when discovery + capability + authenticated delegation + correlation + native execution + failure injection + recovery + composition + provenance all succeed.

## 7. Effort budgets

Stage 1: 2h  
Stage 2: 2h  
Stage 3: 2h  
Stage 4: 2h  
Stage 5: 2h  
Stage 6: 4–6h  
Full seven-runtime soak: 2h+

These are engineering allocation budgets, not asynchronous promises.
