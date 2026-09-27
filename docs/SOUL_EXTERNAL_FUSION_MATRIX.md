# SOUL External Fusion Matrix

Audit revision: 2026-09-27 (GitHub live state)
Authoritative implementation: repository code + exact GitHub commit/CI evidence.
This document is architectural mapping only; it does not promote structural evidence to runtime/online readiness.

## Non-negotiable boundaries
- N01-N07 remain the seven SOUL nuclei. SARA is transversal and is not N08.
- One canonical Soul Mesh only: soul-mesh/1, contract target 1.1.0.
- N07 remains the sole public OpenAI-compatible ingress.
- Providers (Gemini, Ollama, optional future providers) do not become new nuclei.
- SuperGPU is a logical orchestration/control-plane model until a real hardware backend is verified.
- No OSS component may introduce a second public ingress, second Mesh authority, duplicate SARA core, or replacement of an existing nucleus.
- Additions must be behind the existing ownership/capability contracts and be proven by CI and smoke/E2E evidence.

## Current N01-N07 / SARA state
| Unit | Repository | Current main/head evidence | Role | Current change | State |
|---|---|---|---|---|---|
| N01 | divibisoul/aeternum-core-29 | main 462fc7d4fc95bdff3307649a3b0c9abd9cffe98d | gateway / Clareira / EventBus / ChatEngine | frozen in this front; no Aeternum modal expansion | OBSERVED / FROZEN |
| N02 | divibisoul/Eternium- | main f3dc0bd...; PR #21 head fbfde114... | conversation + provider bridge | Ollama bridge, canonical Mesh updates, HMAC envelope response | IN_PR / CI GREEN |
| N03 | divibisoul/nexus-aeternum-fusion | main 726a0f880d72d9fa7d6624742a69da568c6ad21b; PR #17 merged | perception / multimodal / Gemini Live | Gemini Live + N07 integration already merged; current-main validation remains required | MERGED / REVALIDATE |
| N04 | divibisoul/nextjs-ai-chatbots | main 258f0a9dc3c6237fdab138f3fbeda929305cc126; PR #23 head 03c1f8d60... | tools / documents / artifacts | Gemini Skills real handlers + Mesh HMAC reference implementation | IN_PR / CI GREEN |
| N05 | divibisoul/nextjs-ai-chatbot | main aa99fda429463759512ceb649f590ab505a58778; PR #21 head 2a9bc1a... | dispatch / inference | N07 peer topology + adaptive transport updates + dependency/lock refresh | IN_PR / CI GREEN; LOCK RECONCILIATION REVIEW |
| N06 | divibisoul/nextjs-ai-chatbot-2000 | main e987f607d73f11751ab1a1a4b9a75950012dd200; PR #17 head f850de6d... | cognition / session / synthesis | federation configuration and dependency alignment | IN_PR / CI GREEN |
| N07 | divibisoul/Orquestrador- | main eb01a35400400321b9156433b681d40c29abd2b2; PR #36 head bee3247b... | orchestration / federation / SuperGPU control plane | OpenAI-compatible ingress + dynamic routing additions | IN_PR on non-main base; MAIN REVALIDATION REQUIRED |
| SARA | divibisoul/SARA | current main tree inspected live | transversal regenerative authority | ARA / ETR / ITR / memory / governance remain owned here | TRANSVERSAL / DO NOT DUPLICATE |

## PR audit
| PR | Base → Head | Exact observed delta | CI evidence at audit | Action |
|---|---|---|---|---|
| N02 #21 | main → feat/soul-ollama-provider | 11 files, +313/-18, 13 commits | Runner Forensics SUCCESS; validation diagnostics SUCCESS | keep IN_PR; validate bridge against legacy |
| N03 #17 | main → feat/soul-n03-n07-live | 9 files, +342/-26, 12 commits | historical PR had failures, but PR is already MERGED | no merge action; validate current main |
| N04 #23 | main → feat/soul-gemini-skills | 15 files, +1526/-568, 39 commits | Dependency Review SUCCESS; Soul Mesh CI SUCCESS; N03 diagnostics SUCCESS; N04 CI SUCCESS | candidate for merge only after dependency/current-main checks |
| N05 #21 | main → feat/soul-n05-n07-activation | 9 files, +764/-535, 24 commits | Dependency Review SUCCESS; N05/N07 Bridge CI SUCCESS; N04 diagnostics SUCCESS | keep IN_PR; lock/manifest reconciliation required |
| N06 #17 | main → feat/soul-n06-federation-config | 6 files, +837/-600, 22 commits | Soul Mesh CI SUCCESS; Channel Contract SUCCESS; Dependency Review SUCCESS; N06 diagnostics SUCCESS | candidate after exact-head review |
| N07 #36 | feat/jev-soul-integration → feat/soul-federation-expansion | 12 files, +429/-36, 14 commits | PR head N07 CI SUCCESS; branch is not based on main | do not merge blindly; reconcile with current main |

## Canonical fusion placement
| Capability | Native owner | Integration point | Required proof |
|---|---|---|---|
| Gemini conversation generation | N02 | N02 provider bridge -> N07 CallBestDynamic | provider smoke + Mesh response/HMAC proof |
| Ollama local generation | N02 | N02 Ollama bridge; local/server runtime only | isolated provider test + Mesh smoke |
| Gemini Live perception/voice | N03 | N03 Live session <-> N07 | ephemeral-token flow + session smoke |
| Documents/tools | N04 | N04 real handlers + N07 delegation | tool.run/document.* integration tests |
| Dispatch/inference | N05 | N05 peer adapter -> N07 dynamic route | route selection + fallback/cancellation evidence |
| Cognition/session | N06 | N06 federation adapter -> N07 | capability discovery + session/correlation evidence |
| Regeneration | SARA | transversal service boundary | SARA service contract + regression tests |
| Orchestration | N07 | Mesh discovery / CallBestDynamic / aggregation | exact-head CI + E2E |
| SuperGPU | N07 | logical bounded parallel scheduler | deterministic parallel tests; no physical GPU claim without device evidence |

## OSS research inventory
Live GitHub repository metadata observed 2026-09-27. Star counts are a snapshot, not a quality score.
| Priority | OSS | Repository | Stars | License observed | Proposed SOUL placement | Gate |
|---|---|---|---:|---|---|---|
| P0 conditional | LiteLLM | BerriAI/litellm | 59,705 | MIT outside enterprise/ | N02 provider layer or private sidecar | only if Gemini+Ollama coverage is insufficient; never expose second public /v1 |
| P1 | LangGraph | langchain-ai/langgraph | 42,354 | MIT | N07 orchestration adapter | minimal graph: decomposition -> CallBestDynamic -> aggregate |
| P2 | NVIDIA NeMo Agent Toolkit | NVIDIA/NeMo-Agent-Toolkit | 2,646 | Apache-2.0 | N07 Python bridge / observation | optional; no SARA-core rewrite |
| P3 | CrewAI | crewAIInc/crewAI | 59,087 | MIT | thin N02/N06 adapter | wait for stable Mesh/HMAC |
| P4 | LlamaIndex | run-llama/llama_index | 52,330 | MIT | N04 RAG/document layer | after skills are stable; prefer current workflow APIs |
| Research | Letta | letta-ai/letta | 24,903 | Apache-2.0 | N06 memory/session adapter | only after N06 memory authority is mapped |
| Research | Microsoft Agent Framework | microsoft/agent-framework | 13,822 | MIT | N07 orchestration/workflow adapter | compare against LangGraph before selecting overlap |
| Research | AgentScope | agentscope-ai/agentscope | 32,450 | Apache-2.0 | N06/N07 multi-agent adapter | additive only |
| Research | CAMEL | camel-ai/camel | 17,782 | Apache-2.0 | N02/N06 role/agent adapter | later |
| Research | Dify | langgenius/dify | 157,321 | license requires package-level review | external/self-host workflow integration | not a core dependency yet |
| Deprecated path | AutoGen | microsoft/autogen | 61,186 | CC-BY-4.0 in repo metadata | do not add as new core layer | Microsoft documents Agent Framework as successor |

## OSS insertion law
An OSS repository is consumed as a capability/adaptor, not copied wholesale into a nucleus.
1. Audit upstream APIs, license, runtime, dependency footprint and overlap with current native modules.
2. Map upstream capabilities to an existing N01-N07 owner.
3. Add the smallest connector/adapter that preserves the native implementation.
4. Reuse canonical Mesh discovery, HMAC/correlation, timeout, retry and observability.
5. Add deterministic tests before considering the adapter implemented.
6. Promote to integrated only after exact-head CI is green.
7. Promote to online only after real peer/runtime smoke passes.

## Current gate
OSS integration is intentionally NOT started from this document alone. F2 must close the N07 current-main validation gap and N05 package/lock reconciliation first.