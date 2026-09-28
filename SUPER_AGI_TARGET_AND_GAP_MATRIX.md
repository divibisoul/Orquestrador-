# SOUL / AETERNUM — Super AGI Target & Gap Matrix

## 1. Architectural meaning

SOUL is being engineered as a federated AI operating system. The current architecture contains seven SOUL nuclei (N01–N07) plus SARA as a transversal regenerative/governance authority. SARA is not N08.

A "Super AGI" is treated here as the target capability envelope, not as a claim about current scientific status. A component is only promoted from STRUCTURAL/DECLARED to operational after reproducible execution evidence.

The target is:

IDENTITY → PERCEPTION → MEMORY → REASONING → GOAL → PLAN → TOOL USE → ACTION → OBSERVATION → VALIDATION → LEARNING → SELF-CORRECTION → GOVERNANCE → CONTINUOUS OPERATION

No layer may bypass the canonical SOUL Mesh or replace a native nucleus owner.

## 2. What a target Super AGI needs

| Capability layer | Target function | Native SOUL owner / source | Current evidence | Gap to close |
|---|---|---|---|---|
| Identity | stable nucleus identity, source/target, correlation and provenance | N01–N07 + SARA | REAL/STRUCTURAL by subsystem | close cross-runtime operational identity proof |
| Communication | authenticated, correlated, low-latency federation | N07 Mesh + peer runtimes | REAL locally; external E2E BLOCKED | commission live edges; measure latency/error |
| Discovery | discover only executable capabilities | N07 + peer discovery | IMPLEMENTED/STRUCTURAL | seven-core runtime proof |
| Routing | choose executable owner using health/latency/failure evidence | N07 CallBestDynamic | IMPLEMENTED/STRUCTURAL | live multi-peer measurements |
| Goal management | persistent objective, criteria, TTL, risk/cost/urgency/impact | N07 cognitive layer | STRUCTURAL on PR #42 | register in runtime and validate live |
| Planning | decomposition into executable steps | N07 cognitive planner + existing orchestration | STRUCTURAL | connect to production path; add dependency-aware parallel waves |
| Execution | invoke native capabilities without duplication | N01–N06 native owners via Mesh | REAL/STRUCTURAL | close real capability transaction matrix |
| Agent delegation | specialists/handoffs/agents-as-tools | N02/N04/N05/N06 + future OpenAI Agents adapter | PARTIAL | explicit inter-agent handoff contract + evidence |
| Tool use | functions, MCP, external services, document/tool execution | N04/N06/N07 | PARTIAL/STRUCTURAL | expose only executable tools and validate schemas |
| Short-term memory | bounded working context | N06 + N07 cognitive working memory | STRUCTURAL | runtime recall policy |
| Long-term memory | durable episodic/state memory and recall | SARA + existing Supabase/vector surfaces | STRUCTURAL/AVAILABLE | automatic retrieval into planning context |
| Perception — text | language/input processing | N02/N05 | REAL/STRUCTURAL | cross-nucleus proof |
| Perception — audio | transcription, synthesis and audio analysis | N03 | PARTIAL; several adapters pending | finish real adapters where dependencies are available |
| Perception — vision | image understanding | provider/runtime surfaces | NOT MEASURED | add a native owner boundary and real execution evidence |
| Realtime | low-latency streaming voice/multimodal interaction | N03 + realtime transport surfaces | STRUCTURAL/NOT MEASURED | live session proof and transport metrics |
| World state | structured representation of current external/internal state | N07 + SARA + memory owners | GAP | introduce an explicit state model without duplicating memory authority |
| Reasoning | model-backed analysis and synthesis | N02/N05/N06 | REAL/STRUCTURAL | multi-step evidence and evaluation |
| Neural learning | forward/backprop/learn | N07 neural + N01/N02 federation | IMPLEMENTED/STRUCTURAL | close learning loop with observed outcomes |
| Evaluation | compare expected vs observed outcome | N07/SARA/prefrontal | PARTIAL | formal outcome evaluator and regression evidence |
| Metacognition | uncertainty, confidence, self-monitoring | N07 prefrontal + observability | PARTIAL | explicit uncertainty/outcome feedback contract |
| Self-improvement | failure → finding → correction → test | SARA ARA/ETR/ITR + RGO governance | STRUCTURAL | complete multi-cycle evidence-backed learning loop |
| Safety/guardrails | admission, policy, risk, irreversible-action review | N07 prefrontal + SARA ETR | REAL/STRUCTURAL | use on every action path that can mutate state |
| Provenance | trace each decision/action back to source | SARA + N07 observability | STRUCTURAL | one cross-nucleus trace format |
| Rollback/recovery | emergency rollback and controlled retry | SARA + N07 resilience | IMPLEMENTED/PARTIAL | real failure injection and recovery evidence |
| Distributed compute | bounded parallel work and aggregation | N07 SuperGPU/Octacore | STRUCTURAL | external multi-nucleus execution proof |
| Sandbox | isolated workspace/code execution | future agent/sandbox boundary | GAP | add only through an owned, isolated adapter |
| Web/research | grounded web retrieval | provider/agent tool layer | NOT MEASURED in SOUL runtime | connect through owned tool boundary; never bypass governance |
| Human-in-the-loop | explicit human approval for sensitive operations | N07/SARA policy boundary | PARTIAL | define approval state/trace contract |
| Observability | metrics, traces, correlation, failure classification | N07 + SARA + provider tracing | REAL/STRUCTURAL | unify cross-runtime evidence |
| Continuous operation | restart/recovery, session continuity, state restoration | N07 + SARA + runtimes | PARTIAL | persistent checkpoints and resume semantics |

## 3. Example target architecture from OpenAI

The official OpenAI Agents SDK for TypeScript is a useful reference implementation for the agentic layer: agents with tools, sandbox agents, realtime agents, delegation/handoffs, guardrails, sessions and tracing. It also supports function tools and MCP servers. This is a reference for missing orchestration primitives; it is not treated as proof that SOUL or any other system is a scientifically demonstrated Super AGI.

Reference repository:
https://github.com/openai/openai-agents-js

Reference capabilities to map into SOUL:

- Agent loop / multi-turn tool loop
- Function tools
- MCP tool calling
- Agents-as-tools
- Handoffs
- Guardrails
- Sessions / working context
- Realtime agents
- Sandbox agents
- Human-in-the-loop
- Tracing

The integration rule is:

OpenAI agent capabilities become adapters/providers inside native SOUL ownership; they do not create a new SOUL nucleus, a second Mesh, or a second public ingress.

## 4. Tool/function inventory for the target

The target tool plane is grouped by function rather than copied from one framework:

### Core cognitive functions
`goal.create` `goal.validate` `goal.plan` `goal.execute` `goal.observe` `goal.evaluate` `goal.replan`
`memory.store` `memory.recall` `state.read` `state.update` `reasoning.run` `reasoning.verify`

### Federation functions
`mesh.describe` `mesh.discovery` `mesh.delegate` `mesh.execute` `mesh.health` `mesh.route` `mesh.trace`

### Specialist functions
`ai.generate` `audio.transcribe` `speech.synthesize` `audio.analyze` `vision.analyze` `document.process` `artifact.process` `tool.execute`

### Compute functions
`supergpu.describe` `supergpu.execute` `supergpu.parallel` `neural.forward` `neural.learn` `neural.backprop`

### Governance functions
`sara.audit` `sara.cycle` `sara.regenerate` `sara.state` `sara.capabilities` `sara.trace` `policy.evaluate` `policy.approve` `policy.rollback`

These are target contracts. A name is not considered operational merely because it appears in a registry.

## 5. Communication requirement

The system must optimize for speed + precision, not raw throughput at the expense of integrity.

The canonical transport envelope must preserve:

`protocol | contractVersion | id | correlationId | source | target | capability | timestamp | nonce | HMAC | payload`

The existing N07 peer path already provides timeout, discovery caching, retries/backoff, circuit breaking, HMAC and correlation. The next proof is empirical: measure connection reuse, end-to-end latency, error rate, retries and recovery on real peers.

No second message bus is permitted.

## 6. What is still a true structural blocker

1. Real six/seven-runtime federation evidence is incomplete.
2. N07 cognitive runtime registration is not yet on the production main path.
3. Working-memory exists, but automatic long-term recall is not yet part of the goal loop.
4. Parallel cognitive execution is not yet connected to the bounded parallel scheduler.
5. Vision runtime is not yet evidenced.
6. Realtime multimodal execution is not yet evidenced.
7. A world-state model is not yet explicitly separated as a first-class contract.
8. Failure→Finding→Correction→Regression is implemented as a principle and in parts of SARA/N07, but not yet demonstrated as a closed multi-cycle learning loop.
9. Cross-nucleus provenance/trace is not yet unified end-to-end.
10. Human approval and sandbox boundaries need an explicit production contract.

## 7. Protected scope

The following are protected from deletion, burial or replacement:

- N01 Clareira and its native processing graph, channels, homeostasis and event path.
- All native N01–N06 agents, tools, registries, capabilities and runtimes.
- N07 orchestration, Mesh, routing, neural, prefrontal, SuperGPU, Octacore and observability surfaces.
- SARA ARA/ETR/ITR, memory, provenance, rollback and governance.
- The broader architectural claim of 72 nódulos. No node is marked VERIFIED until it is enumerated from source.

Integration raises system connectivity; it does not lower any native subsystem.

## 8. Exit condition for this stage

This target/gap stage is closed only when each gap has:

`OWNER → SOURCE → CONTRACT → IMPLEMENTATION → TEST → LIVE EVIDENCE (when applicable) → TRACE → NEXT GATE`

The presence of this matrix does not promote any blocked runtime gate to PASS.