# SOUL Live Architecture Matrix

This file is the durable architectural memory shared by the six parallel engineering fronts.

## Topology rule

Base AI nuclei: **N01, N02, N03, N04, N05, N06**.

Orchestration/fusion layer: **N07**.

N01-N06 each require five bidirectional base-peer relationships. N07 is an additional Mesh participant and orchestration endpoint connected to all six base nuclei.

```text
                N07
        Orchestration / Fusion
       /  /  /  |  \  \  \
     N01 N02 N03 N04 N05 N06
      \_____________________/
        bidirectional base Mesh
```

## Required capability model

Every nucleus must expose, where applicable:

- identity;
- agents;
- capabilities;
- tools;
- providers;
- context;
- memory interfaces;
- execution;
- inputs;
- outputs;
- discovery;
- delegation;
- response/correlation;
- observability;
- security;
- resilience;
- tests and CI.

N07 additionally owns orchestration concerns: routing, distributed execution coordination, composition, result aggregation, validation and SuperGPU scheduling.

## Connection matrix

| A | B | Transport | Discovery | Delegation | Capability invocation | Agents | Tools | Parallelism | Validation | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| N01 | N02 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N01 | N03 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N01 | N04 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N01 | N05 | HTTP/Mesh candidate | Partial | Partial | Partial | Partial | Partial | Candidate | Required | IN PROGRESS |
| N01 | N06 | Mesh | Partial | Partial | Partial | Partial | Partial | Candidate | Required | IN PROGRESS |
| N01 | N07 | HTTP/Mesh | Validated structurally | Prepared | Canonical route | Prepared | Prepared | N07 orchestration | In progress | IN PROGRESS |
| N02 | N03 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N02 | N04 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N02 | N05 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N02 | N06 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N02 | N07 | Mesh candidate | Prepared structurally | Prepared | Canonical route | Prepared | Prepared | N07 scheduler | Required | OPEN |
| N03 | N04 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N03 | N05 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N03 | N06 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N03 | N07 | Mesh candidate | Prepared structurally | Prepared | Canonical route | Prepared | Prepared | N07 scheduler | Required | OPEN |
| N04 | N05 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N04 | N06 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N04 | N07 | Mesh candidate | Prepared structurally | Prepared | Canonical route | Prepared | Prepared | SuperGPU candidate | Required | IN PROGRESS |
| N05 | N06 | TBD from live code | Audit required | Audit required | Audit required | Audit required | Audit required | Audit required | Required | OPEN |
| N05 | N07 | Mesh candidate | Prepared structurally | Prepared | Canonical route | Prepared | Prepared | SuperGPU candidate | Required | IN PROGRESS |
| N06 | N07 | Mesh | Structural | Prepared | Canonical route | N06 runtime | N06 tools | SuperGPU candidate | Required | IN PROGRESS |

The table intentionally distinguishes **structural preparation** from **operational validation**. No row may be promoted to operational without evidence from the exact revision being claimed.

## Pairwise synergy ledger

| Pair | Strongest discovered complement | Candidate emergent function | Existing implementation | Missing piece | Evidence | Status |
|---|---|---|---|---|---|---|
| N01 x N02 | To be derived from live capability inventories | Not yet named | Audit in progress | Capability inventories | Pending | OPEN |
| N01 x N03 | To be derived from live capability inventories | Not yet named | Audit in progress | Capability inventories | Pending | OPEN |
| N01 x N04 | Routing + execution opportunity | Not yet named | Partial Mesh | Cross-runtime proof | Pending | OPEN |
| N01 x N05 | Input/service + inference | Inference delegation candidate | N05 inference runtime | End-to-end proof | N05 CI green | IN PROGRESS |
| N01 x N06 | User/context + cognitive/tool support | Cognitive delegation candidate | N06 runtime | End-to-end proof | N06 structural fixes | IN PROGRESS |
| N01 x N07 | Request/response + orchestration | Orchestrated remote execution | N07 HTTP gateway | Full route proof | N01/N07 tests | IN PROGRESS |
| N04 x N07 | Execution backend + orchestration | Distributed execution scheduling | N04 execution; N07 SuperGPU | Cross-runtime scheduler | Pending | IN PROGRESS |
| N05 x N07 | Inference + orchestration | Distributed inference scheduling | N05 inference pool; N07 router | Cross-runtime scheduling | N05 CI green | IN PROGRESS |
| N06 x N07 | Cognitive support + orchestration | Cognitive-plan-to-compute pipeline | N06 runtime; N07 cognitive route | End-to-end proof | N07 CI green | IN PROGRESS |

Do not create an emergent capability until its contract, owner, input/output model, dependencies, execution path and tests are defined.

## SuperGPU flow

```text
TASK
  -> decomposition
  -> capability discovery
  -> routing
  -> resource/capacity selection
  -> parallel inter-nucleus execution
  -> parallel intra-nucleus execution
  -> result aggregation
  -> validation
  -> retry/fallback where justified
  -> final result
```

SuperGPU is an architectural execution model. It must never claim physical GPU acceleration unless a real backend and device execution path are present and verified.

## Front synchronization contract

Every parallel front must record, in a commit or update to this file or another linked artifact:

`WHAT_CHANGED / WHAT_WAS_FOUND / WHAT_REMAINS / EXACT_REVISION / CI_EVIDENCE / NEXT_ACTION`

This document is a map, not a substitute for tests. The repositories and their exact revisions remain the source of executable truth.
