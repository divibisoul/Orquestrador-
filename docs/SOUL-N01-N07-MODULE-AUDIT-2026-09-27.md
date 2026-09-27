# SOUL Module Audit — N01–N07 + SARA
Date: 2026-09-27
Authority: live GitHub tree, repository CI and exact commit evidence.

## Audit rule
Repository presence is OBSERVED only. VALIDATED requires an executed owning CI/test path. ONLINE requires real deployed runtime evidence.

| Unit | Files | Tests | Workflows | Mesh | Capability | State |
|---|---:|---:|---:|---:|---:|---|
| N01 | 391 | 18 | 8 | 96 | 16 | REMEDIATION VALIDATING |
| N02 | 155 | 1 | 4 | 50 | 9 | VALIDATED |
| N03 | 195 | 3 | 7 | 56 | 11 | MERGED; POST-MERGE REVALIDATION |
| N04 | 270 | 18 | 4 | 54 | 11 | VALIDATED |
| N05 | 268 | 14 | 5 | 66 | 4 | VALIDATED |
| N06 | 273 | 15 | 4 | 54 | 11 | VALIDATED |
| N07 | 140 | 30 | 7 | 19 | — | CORE VALIDATED |
| SARA | 125 | 20 | 2 | 0 | — | TRANSVERSAL |

## Functional ownership

### N01
Owns Clareira/Aeternum gateway, native neural processing and ChatEngine. Its Mesh layer covers health, telemetry, resilience and transports. soul-sentinel contains Android bootstrap, Mesh transport and instrumentation tests. Aeternum modal expansion remains frozen in this front.

### N02
Owns conversation generation and providers. Gemini and Ollama remain provider implementations, not nuclei. N02 owns provider bridges, peer fabric, discovery, transports and response-HMAC verification.

### N03
Owns multimodal/perception and Gemini Live. N03PairFusion is dimension-scoped so an identifier shared between different dimensions is not falsely counted as overlap.

### N04
Owns real tools, documents and artifact execution. Gemini Skills and canonical Mesh HMAC remain inside N04; its API routes are nucleus integration surfaces, not a second system Mesh.

### N05
Owns dispatch/inference-side execution, adaptive transport, inference pool/cache, resilience and N05/N06 synergy. N07 remains orchestration authority.

### N06
Owns cognition/session/tool runtime and N05/N06 interop. It does not become N07.

### N07
Owns federation/control plane. api/openai_compat.go is the sole public OpenAI-compatible ingress. mesh/dynamic_router.go owns capability/availability/latency/failure-aware routing. SuperGPU remains logical bounded scheduling.

### SARA
Owns ARA/ETR/ITR, contracts, governance, provenance, memory and federation. It is transversal and is not N08.

## Shared invariants
1. One canonical Soul Mesh contract.
2. N07 is the only public OpenAI-compatible ingress.
3. Gemini/Ollama remain N02 providers.
4. CallBestDynamic remains capability-driven.
5. SARA is transversal; no duplicated ARA/ETR/ITR.
6. No physical SuperGPU claim without device evidence.
7. HMAC protection uses nonce, timestamp, canonicalization and correlation at active boundaries.
8. Remediation is additive; existing modules are preserved.

## Integration map
Client → N07 ingress → CallBestDynamic → N02 generation.
N03 Live ↔ N07; N04 tools/docs ↔ Mesh; N05/N06 federation ↔ N07; SARA ↔ transversal service boundary.
LangGraph is an external planner bridge to N07, not a Mesh participant. NeMo is an external observation/evaluation environment, not a SARA replacement.

## Remaining gates
N01: finish Android exact validation, then merge PR #40 only with required green evidence.
N03: record post-merge main revalidation after current lock synchronization.
N07 staging: blocked by missing deployment secrets.
LiteLLM: conditional only. CrewAI/LlamaIndex/Letta/Agent Framework/AgentScope/CAMEL/Dify remain pending until a concrete non-overlapping capability is proven.


## Re-audit update — 2026-09-27

| Unit | Updated state |
|---|---|
| N01 | WEB/Mesh remediation VALIDATED; Android debug/release builds + artifact checks SUCCESS; instrumentation still running |
| N02 | VALIDATED |
| N03 | VALIDATED on post-merge main: Soul Mesh CI + N05 diagnostics SUCCESS |
| N04 | VALIDATED |
| N05 | VALIDATED |
| N06 | VALIDATED |
| N07 | CORE VALIDATED; staging secrets still BLOCKED |
| SARA | TRANSVERSAL |

### Current OSS research boundary

The following repositories were rechecked against GitHub metadata on 2026-09-27 and remain candidates without automatic insertion:
- Microsoft Agent Framework — MIT; orchestration/workflows; potential N07 adapter, but overlapping with LangGraph.
- Letta — Apache-2.0; agent memory; potential N06 adapter after memory authority is mapped.
- AgentScope — Apache-2.0; multi-modal/multi-agent/realtime capabilities; potential N06/N07 adapter.
- CAMEL — Apache-2.0; multi-agent society/role composition; potential later N02/N06 adapter.
- smolagents — Apache-2.0; lightweight agent/tool execution; useful as an executor experiment, but it overlaps existing N04/N06 tool ownership.
- Dify — repository reports NOASSERTION license metadata; requires package-level license review before any dependency decision.
- LiteLLM — repository reports NOASSERTION metadata and is a gateway/proxy; remains conditional because N02 already has Gemini + Ollama and N07 must remain the sole public ingress.

No candidate above is installed merely because it is popular. Similarity of function, ownership, dependency cost and non-duplication are required before insertion.

## HMAC audit update — N06

N06 had a concrete bidirectional-contract gap: request signing existed, but the active response path did not expose or verify top-level response nonce/HMAC consistently, and the N06 canonicalizer omitted contractVersion.

PR #18 now adds:
- contractVersion to N06 HMAC canonicalization
- top-level nonce/hmac response fields
- response signing in N06 route and shared endpoint
- response verification in peer-client and N06PeerAdapter
- deterministic HMAC tests in Soul Mesh CI

State: IN VALIDATION until exact-head CI is green.