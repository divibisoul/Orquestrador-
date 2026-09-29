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


## Live audit addendum — 2026-09-27

This addendum records the current repository/PR evidence without replacing the historical architecture map above.

### Exact current revisions

| Nucleus | main revision observed | Relevant PR | PR state at audit |
|---|---|---|---|
| N01 | 462fc7d4fc95bdff3307649a3b0c9abd9cffe98d | none in this front | FROZEN |
| N02 | f3dc0bd2630d6d5c77873edc0986262820c3b93a | #21 / fbfde114cef56ccfaa65de956b0c707ee55f3731 | OPEN / CI GREEN |
| N03 | 726a0f880d72d9fa7d6624742a69da568c6ad21b | #17 / d8c4fb956f37c325bbdba09250e852e2ed0590a9 | MERGED / REVALIDATE MAIN |
| N04 | 258f0a9dc3c6237fdab138f3fbeda929305cc126 | #23 / 03c1f8d60e3c17f09ff810b73863f87aca19fa6d | OPEN / CI GREEN |
| N05 | aa99fda429463759512ceb649f590ab505a58778 | #21 / 2a9bc1a7fcdb2e8c116260820afcb62baa0ec578 | OPEN / CI GREEN / LOCK REVIEW |
| N06 | e987f607d73f11751ab1a1a4b9a75950012dd200 | #17 / f850de6bd17dd89ea3a3a2a761eabf4d7f77044d | OPEN / CI GREEN |
| N07 | eb01a35400400321b9156433b681d40c29abd2b2 | #36 / bee3247b80c2c0263fa112f39be90050151b69d4 | OPEN / non-main base |

### Current N07 evidence boundary

PR #36 is based on feat/jev-soul-integration, not main. Its exact delta is 14 commits, 12 files, +429/-36. The PR head had a successful N07 Orquestrador CI run, but this cannot be used as proof that current main contains the same complete state.

Current main has the OpenAI-compatible handler at api/openai_compat.go and the main branch was subsequently normalized by a formatting commit. The latest pre-normalization N07 CI failed because api/openai_compat.go required gofmt. The latest N07 E2E commissioning also exposed two independent environment/code issues: missing staging secrets and an undefined writeJSON symbol in api/openai_compat.go on the main revision tested. These are tracked as F2 blockers until a new exact-head run proves closure.

### Canonical routing rule

OpenAI-compatible ingress remains N07-only. The request path is:

client -> N07 /v1/chat/completions -> canonical Soul Mesh -> CallBestDynamic -> N02 capability -> configured Gemini/Ollama provider.

No provider is a nucleus. No second public /v1 ingress is introduced. SARA remains transversal.

### OSS boundary

OSS integrations are deferred until the minimum F2 closure. LangGraph is mapped to N07 orchestration, LiteLLM remains conditional behind N02/private sidecar, and NeMo Agent Toolkit remains an optional N07 bridge. CrewAI/LlamaIndex/Letta/Agent Framework/AgentScope/CAMEL/Dify remain research candidates until the earlier gates are green.

See docs/SOUL_EXTERNAL_FUSION_MATRIX.md for the complete mapping and exact audit snapshot.

## Post-fusion audit — 2026-09-27

The live chain has now advanced without replacing the native nuclei:

N01 (frozen Clareira) → N02 conversation/providers → N03 perception → N04 tools/docs → N05 dispatch → N06 cognition/session → N07 orchestration/ingress.

N07 remains the only public OpenAI-compatible ingress. Gemini and Ollama are provider implementations behind N02. CallBestDynamic remains the routing authority; providers do not become nuclei. SuperGPU remains a logical control-plane abstraction until a real device backend is verified. SARA remains transversal.

Current post-merge main evidence:
- N02 exact main 993ad528...: Soul Mesh CI SUCCESS; validation diagnostics SUCCESS; Runner Forensics SUCCESS.
- N04 exact main 8c457291...: Soul Mesh CI SUCCESS; N03 diagnostics SUCCESS; N04 CI SUCCESS.
- N05 exact main a405dd02...: Bridge CI SUCCESS; Soul Mesh CI SUCCESS; N04 diagnostics SUCCESS.
- N06 exact main 5c11eaf3...: Soul Mesh CI SUCCESS; Channel Contract SUCCESS; diagnostics SUCCESS.
- N07 main a7f68f5a...: principal CI SUCCESS and production image inspection SUCCESS. N07 staging commissioning remains blocked only by missing staging secrets, not by the core build/test chain.
- N07 PR #36 has been retargeted to main and synchronized through a real merge commit; current exact-head PR CI is validating the complete federation/OpenAI ingress delta.

No claim of online runtime readiness is made from static source or structural CI alone.

## F4 post-merge architecture — 2026-09-27

The current additive topology is:

Open WebUI / OpenAI-compatible client
→ N07 sole /v1/chat/completions ingress
→ existing CallBestDynamic
→ Soul Mesh / N02-N06
→ provider/tool/perception/cognition owners

An optional LangGraph planner may sit outside that control path. It delegates to N07; it does not become an eighth nucleus, a second Mesh, a provider authority, or a second ingress. LangGraph 1.2.12 was selected from the current PyPI release and its StateGraph/START/END/compile model matches the required minimal graph. Official references: https://pypi.org/project/langgraph/1.2.12/ and https://github.com/langchain-ai/docs/blob/main/src/oss/langgraph/graph-api.mdx

An optional NeMo Agent Toolkit environment is also external to N07's Go runtime. Official NVIDIA documentation describes nvidia-nat as framework-agnostic and provides a LangChain/LangGraph integration; the SOUL adapter uses only the stable 1.8.x package line and does not rewrite SARA. Official reference: https://docs.nvidia.com/nemo/agent-toolkit/latest/quick-start/installing.html

No production-online claim is inferred from these OSS integration checks.


## Primordial essence + composition continuity — 2026-09-29

The system is now tracked on two distinct levels:

1. **Primordial essence:** the native responsibility that exists before pairwise fusion. It remains owned by the original nucleus/runtime and is never replaced by the composition layer.
2. **Composition:** the new executable workflow made possible when two or more native responsibilities exchange compatible capabilities, context, tools, agents and results.

This separation prevents a recurring failure mode: a newer “super” layer can accidentally hide or abandon the native module that gave it the underlying capability.

### Current mainline source state observed

| Part | Exact mainline revision observed | Native essence | Current continuity state |
|---|---|---|---|
| N01 | `99fb9d22ddea3439c8b5be396ccae9e11ae3f66e` | core/gateway/runtime + intent/AGI | MERGED recovery + HortaCore continuity + NeuralManagementCore |
| N02 | `f2fcbf9381c90a27d944794f1275b4bafa2f3516` | language/conversation/context | MERGED BNCv2 + CSAE + DCRS + VagusBus; further executor-evidence corrections present |
| N03 | `9951485529c8480ff3bf4d7305538d152b6acdfe` | perception/audio/speech/multimodal | MERGED composition contract repair |
| N04 | `ed4e2a75b772d1efaed2457c392f52bddaffbfa8` | tools/documents/artifacts/execution | MERGED tool.execute compatibility boundary |
| N05 | `e0c8778bebaba4a4b821bc6923c2a51dcfc431ca` | inference/conversation/dispatch | MERGED executable capability catalog |
| N06 | `bdcb099166701453b6cb62672fa6a642addbd9b0` | cognition/synthesis/planning/validation | MERGED composition-target repair |
| N07 | `abf01b43611a582bf0205be49818c5925fccde0f` | orchestration/fusion/SuperGPU control | MAIN composition authority + current recovery/authority fronts |
| SARA | `20d550029c5320bf559fe356af02ed7aa49cb4ad` | regeneration/audit/governance/memory/Bayesian | MERGED Chimera/Vagus/OctaCore recovery + Bayesian runtime |

### What the connections reveal

The first useful pattern is not “more modules = more AI”. It is **functional closure**: one nucleus produces something that the next nucleus can actually consume without taking ownership away from the producer.

The strongest currently evidenced chain is:

`N03 perception → N02 reasoning/context → N04 tool/artifact execution`

N03 already contains the dual-fusion path that combines N02 cognition and N04 action. N02's restored BNCv2/CSAE/DCRS pipeline supplies an additional internal processing chain before provider-backed generation. N04 provides the execution boundary needed to turn a reasoning result into an artifact/tool operation.

A second closure is:

`N05 inference → N06 planning/validation → N04 execution`

That is materially different from simply connecting three endpoints: inference produces an intermediate semantic result, N06 transforms it into a constrained plan and validation stage, and N04 owns the actual tool/document/artifact execution.

A third closure is transversal:

`N01 context/gateway → N07 routing/composition → N02..N06 native capability → SARA audit/regeneration/governance`

Here SARA is not another copy of a nucleus. Its value comes from observing and governing the composed workflow while the specialized runtime continues to own execution.

### Engineering invariant derived from this audit

Every new optimization must leave a bridge in both directions:

`ORIGIN → CONSUMER → RESULT → NEXT CONSUMER → EVIDENCE`

The origin module remains addressable. The consumer is explicit. The produced result is retained with correlation/provenance. The next consumer is explicit. The evidence level is recorded independently from the architecture claim.

This is the continuity mechanism designed to prevent future fronts from burying a repaired module when they optimize another layer.

### Evidence boundary

The composition entries above are **STRUCTURAL** unless a real cross-runtime execution trace promotes them to `EXECUTED`, `VERIFIED` or `FUSED`. CI can prove code/contract behavior; it does not prove that every remote service is simultaneously online.

### Historical preservation rule

Older architecture notes in this file remain historical records. This section is the current reconciliation layer and must be updated additively when mainline revisions change; historical evidence is not to be deleted merely because the implementation evolved.
