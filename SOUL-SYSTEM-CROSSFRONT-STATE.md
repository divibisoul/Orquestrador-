# SOUL — Crossfront State (Generated)

> Generated from the declarative SOUL matrix. Missing repository evidence is surfaced as DEGRADED; runtime availability is never fabricated.

| Nucleus | Repository | Role | Contract | Mesh | Neural/Federation | Evidence state | Source |
|---|---|---|---:|---|---|---|---|
| N01 | divibisoul/aeternum-core-29 | host-reference-gateway | not-detected | present | not detected | OBSERVED | scripts/soul-mesh-server-entry.mjs |
| N02 | divibisoul/Eternium- | conversation-interaction | not-detected | present | not detected | OBSERVED | api/soul-mesh.ts |
| N03 | divibisoul/nexus-aeternum-fusion | perception-voice-multimodal-context | not-detected | present | not detected | OBSERVED | api/soul-mesh.ts |
| N04 | divibisoul/nextjs-ai-chatbots | tools-documents-artifacts | 1.1.0 | present | not detected | OBSERVED | lib/soul-mesh/SoulMeshProtocol.ts |
| N05 | divibisoul/nextjs-ai-chatbot | orchestration-dispatch-execution | not-detected | not detected | not detected | OBSERVED | app/api/soul-mesh/route.ts |
| N06 | divibisoul/nextjs-ai-chatbot-2000 | cognition-synthesis-audit-governance | not-detected | present | not detected | OBSERVED | app/api/soul-mesh/route.ts |
| N07 | divibisoul/Orquestrador- | super-agi-master-orchestration-federation-supergpu-control-plane | not-detected | not detected | not detected | DEGRADED | not-found |

## Governance

- Canonical Mesh contract target: **soul-mesh/1 / 1.1.0**.
- SOUL is one system; N01–N07 are specialized nuclei.
- The matrix is the source of truth for nucleus identity, role and topology.
- A legacy implementation may remain only as a compatibility adapter pointing at a canonical authority.
- Repository evidence and live runtime evidence are kept separate.


## Current audit snapshot — 2026-09-27

The generated table above is retained as historical evidence. The following section supersedes its stale N07/PR state for current engineering decisions.

### Exact repository revisions

| Unit | Repository | main revision | Current state |
|---|---|---|---|
| N01 | divibisoul/aeternum-core-29 | 462fc7d4fc95bdff3307649a3b0c9abd9cffe98d | OBSERVED / FROZEN in this front |
| N02 | divibisoul/Eternium- | f3dc0bd2630d6d5c77873edc0986262820c3b93a | IN_PR #21; PR head CI green |
| N03 | divibisoul/nexus-aeternum-fusion | 726a0f880d72d9fa7d6624742a69da568c6ad21b | PR #17 already merged; current main revalidation required |
| N04 | divibisoul/nextjs-ai-chatbots | 258f0a9dc3c6237fdab138f3fbeda929305cc126 | IN_PR #23; PR head CI green |
| N05 | divibisoul/nextjs-ai-chatbot | aa99fda429463759512ceb649f590ab505a58778 | IN_PR #21; PR CI green; manifest/lock reconciliation remains |
| N06 | divibisoul/nextjs-ai-chatbot-2000 | e987f607d73f11751ab1a1a4b9a75950012dd200 | IN_PR #17; PR head CI green |
| N07 | divibisoul/Orquestrador- | current main advanced by documentation commits after eb01a354 | main revalidation required; PR #36 is based on feat/jev-soul-integration |
| SARA | divibisoul/SARA | 8d290a2d5a59d714df4f434a5ad1553eadc93e08 | transversal authority; no N08 mapping |

### Current change ledger

- N01 remains frozen for this external-fusion front; no Clareira/Aeternum modal expansion is being mixed into this work.
- N02 #21 is additive and introduces the Ollama provider bridge plus Mesh/HMAC changes. It is not merged.
- N03 #17 is already merged. Its historical PR checks included failures, so current main evidence is authoritative for present readiness.
- N04 #23 is additive and its observed PR checks are green; it is the reference candidate for Mesh HMAC/skills behavior, but remains unmerged.
- N05 #21 has green reported PR checks. Its main package manifest lacks dependencies that are present in the main lock importer, while the PR changes the manifest substantially and adds overrides; this is a real reconciliation task, not a reason to invent a green state.
- N06 #17 has green reported PR checks and remains open.
- N07 #36 adds the OpenAI-compatible adapter, dynamic routing path, tests and formatting repair, but its base is feat/jev-soul-integration rather than main. It must not be merged as though it were a main-based PR.
- N07 current-main CI evidence exposed missing writeJSON in the main executable tree during the E2E verify run, while the same run also lacked required staging secrets. These are separate blockers.
- SARA remains external/transversal. No OSS work may duplicate ARA, ETR or ITR.

### Status vocabulary

OBSERVED = repository evidence exists.
IN_PR = change exists on a PR branch and is not merged.
VALIDATED = exact revision passed the required automated checks.
BLOCKED = a reproducible failure or missing environment evidence prevents validation.
ONLINE = only after real deployed peer/runtime smoke evidence.

No state above uses simulated success.

## F2/F3 completion snapshot — 2026-09-27

### Current state ledger

| Nucleus | Main | State |
|---|---|---|
| N01 | 462fc7d4fc95bdff3307649a3b0c9abd9cffe98d | OBSERVED / FROZEN |
| N02 | 993ad528e5257b33f7fa6283a24a57a7d50f67ec | VALIDATED |
| N03 | 726a0f880d72d9fa7d6624742a69da568c6ad21b | MERGED / REVALIDATION NOT TRIGGERED |
| N04 | 8c457291181b3295b50615bf3abe01599fe0226e | VALIDATED |
| N05 | a405dd02598fa8f74d2209f59b447dab63e821c4 | VALIDATED |
| N06 | 5c11eaf3065bcc38c8d905d4806fbe17b3c11b5b | VALIDATED |
| N07 | a7f68f5a8673120e680d264070815b7a62fcee3c | VALIDATED / STAGING BLOCKED |
| N07 PR #36 | cea7e7641010351afb237fcaa27ba3b5ce0741bf | OPEN / CI IN PROGRESS |
| SARA | current main inspected live | TRANSVERSAL; not N08 |

### Hard boundaries verified

1. One canonical Soul Mesh only.
2. N07 is the sole public OpenAI-compatible ingress.
3. N02 owns Gemini/Ollama provider bridges; no provider became a nucleus.
4. CallBestDynamic remains capability/discovery/health/latency/failure-aware routing in N07.
5. N01/Clareira is frozen in this front.
6. SARA is not duplicated or promoted to N08.
7. No physical SuperGPU claim was introduced.
8. Random values were not introduced into health/readiness logic in this fusion pass.

### Remaining gates

- N03 current-main CI was not independently rerun in this pass because its main revision was unchanged; its prior merge remains historical evidence.
- N07 PR #36 must not merge until its exact-head PR CI finishes green.
- N07 staging remains BLOCKED until the environment supplies the existing required secrets.
- OSS insertion remains behind the canonical boundaries. LiteLLM is still conditional; LangGraph/NeMo are additive candidates after #36 and its post-merge validation.

## F4 completion ledger — 2026-09-27

| Layer | State |
|---|---|
| N01 Clareira/Aeternum | FROZEN in this front |
| N02 Gemini + Ollama provider owner | VALIDATED |
| N03 Gemini Live / multimodal | MERGED; current-main independent rerun still not triggered |
| N04 Skills / documents | VALIDATED |
| N05 Dispatch / topology | VALIDATED |
| N06 Cognition / session | VALIDATED |
| N07 federation / sole OpenAI ingress | CORE VALIDATED; post-F4 main CI running |
| SARA | TRANSVERSAL; unchanged as authority |
| LangGraph | MERGED / VALIDATED |
| NeMo Agent Toolkit | MERGED / exact-PR validated; post-merge main validation pending |
| LiteLLM | CONDITIONAL / not installed |
| CrewAI / LlamaIndex | PENDING |

### Explicit non-regressions

One canonical Soul Mesh is preserved. N07 remains the only public OpenAI-compatible ingress. SARA is not duplicated or promoted to N08. SuperGPU remains logical. No existing nucleus was replaced or deleted.
