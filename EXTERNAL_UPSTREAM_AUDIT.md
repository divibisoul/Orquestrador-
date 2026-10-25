# SOUL — Public Upstream / OSS Integration Audit
**Snapshot:** 2026-10-25  
**Audit base:** N07 `main` SHA `c4582f178a6bc82ad7a7a2ebbe4332cd7982439d`; other nucleus SHAs are recorded in each `SOUL_UPSTREAM_STATUS.md`.  
**This file is on branch:** `audit/ai-nuclei-status-2026-10-25` — no upstream submodule was changed, and no PR was opened.

## Executive result

The eight SOUL repositories contain **50 submodule references covering 26 unique public repositories** in the sampled main trees. N07 has a canonical registry of 16 providers. N02 has an additional registry with 13 mappings, of which 3 refer to projects already in the N07 registry and 10 introduce unique upstream sources.

**Being present as a Git submodule is not the same as a functional integration.** The N07 authority manifest explicitly says every provider is `adapter-planned`, sets runtime execution disabled by default, and requires real authenticated E2E evidence before `adapter-live`. N02’s 21 secondary capability mappings are all `TRIAGED`; its unit test checks catalog mapping and valid owner names, not calls into the external runtimes.

## Non-negotiable integration contract

1. N01–N07 remain the seven SOUL nuclei; SARA is transversal, not N08. Public upstreams are implementation augmenters, not new SOUL authorities.
2. Preserve each native owner and invoke through the existing Mesh / explicit adapter boundary. Do not replace a nucleus with an external project or create a second routing plane.
3. A declaration, submodule SHA, package import, probe, synthetic echo, HTTP 200 or “accepted” envelope does not prove a real capability transaction.
4. Runtime state may progress from `SOURCE_PINNED` to `ADAPTER_BOUND` to `RUNTIME_VERIFIED` only on exact source SHA, isolated provider version, genuine execution and authenticated/correlated returned evidence.
5. Do not update all submodules in bulk. Upstream projects have different release rhythms, package APIs, languages, dependencies and safety models. Reconcile conflicts one adapter at a time.
6. Unverified or restrictive license terms block packaging/distribution, but the source reference is preserved pending review. No silent deletion.

## The canonical 16-provider manifest on N07

Main source paths:
- `integrations/external-capabilities.json`
- `integrations/external-adapters.json`
- `integrations/external-capability-ownership.json`
- `scripts/external-capability-adapter.py`
- `orchestrator/external_adapters.go`
- `.github/workflows/external-adapter-contract.yml`

Observed manifest settings:
- Registry, adapter list and ownership map all describe 16 provider IDs.
- `canonical_control_plane=N07`; provider-to-owner mapping exists.
- `runtime_defaults.enabled=false`, `probe_without_execution=true`, `fail_closed=true`.
- All 16 ownership statuses are `adapter-planned`, not `adapter-live`.
- The adapter contract workflow is `workflow_dispatch` only and validates manifests / Python syntax / registry tests. It does not run a real E2E execution for every upstream.
- `package_available` in the generic Python runner checks whether a package can be imported in the active interpreter. That is **not** proof it is the exact code represented by the Git submodule SHA. Several generic provider handlers import from the active interpreter without first importing from the pinned submodule root. Tag this path `PINNED_SOURCE_VS_RUNTIME_PACKAGE_UNVERIFIED` until a reproducible, isolated environment and version check are provided.

## Main upstream inventory

The “ahead by” values are from GitHub’s compare API at this audit time: how many commits the upstream default branch contains after the SOUL-pinned commit. This is a **drift indicator**, not a recommendation to update automatically. A value of zero means that compare returned the same commit for that pin and current upstream default branch.

| Upstream repository | Native fit / intended use | SOUL pin SHA | Observed upstream drift | State and risk |
|---|---|---|---:|---|
| [AgentScope](https://github.com/agentscope-ai/agentscope) | N03 multimodal/voice/agent boundaries; N07 optional coordination | `72f3f6fa0b2fc38b8517f408ab616f0f2bd229e6` | 107 | `adapter-planned`; a live adapter call is not evidenced |
| [Browser Use](https://github.com/browser-use/browser-use) | N04 browser automation; N05 retrieval/browser augmentation | `302d8fcb245a7a63fb7531a4734c9ce3c7792779` | 113 | Adapter source and runner exist in N04; E2E still needed |
| [CrewAI — N07 pin](https://github.com/crewAIInc/crewAI) | N07 role/team execution | `8078f9130c35a47be95d4a55bf1d73b3fd44fc88` | 122 | Pin conflict with N02; do not mix API assumptions |
| [CrewAI — N02 pin](https://github.com/crewAIInc/crewAI) | N02 proposed specialist handlers | `fbcf2de39d3f8252500b74a0751b92c5774c751b` | 124 | Pin conflict with N07 (2 commits apart); reconcile through one compatibility contract |
| [DSPy](https://github.com/stanfordnlp/dspy) | N06 reasoning/LM-programming; other nuclei by explicit delegation | `ba3f9198efe5d125c7c1a2b40b1f1e6166209bd2` | 52 | `adapter-planned`; check model configuration and actual optimization/evaluation path |
| [Kokoro](https://github.com/hexgrad/kokoro) | N03 local speech synthesis | `dfb907a02bba8152ca444717ca5d78747ccb4bec` | 0 | Pin matches default branch; N03 adapter itself still reports execution unproven until real synthesis succeeds |
| [LangGraph](https://github.com/langchain-ai/langgraph) | N07 stateful planning/routing workflow; N01 context orchestration | `157a06dda988d85afeb8751ff27b35ab3f4f8bf4` | 96 | Real local bridge exists. The generic adapter had a false-positive echo path; focused correction/test is in isolated fix branch, pending new CI |
| [Letta Code](https://github.com/letta-ai/letta-code) | N01 persistent identity/memory augmentation | `1fcc9666817ab852bc2532a3a989f712e1fd6c19` | 190 | Adapter shells out to CLI; runtime package/source and persistence scope need proof |
| [LlamaIndex](https://github.com/run-llama/llama_index) | N05 indexing/retrieval/RAG | `962940ddc079cc21701d28d1237c84c82a7c5164` | 9 | Generic adapter constructs an index from supplied text; no persistence or production retrieval E2E evidenced |
| [MetaGPT](https://github.com/FoundationAgents/MetaGPT) | N07/N02/N06 role-based collaboration | `11cdf466d042aece04fc6cfd13b28e1a70341b1f` | 0 | Pin matches default branch; runner exists, real online execution not evidenced |
| [Microsoft Agent Framework — N07 pin](https://github.com/microsoft/agent-framework) | N07 workflows / agent execution | `a2f4506c0ba30cea7c9bbe907fc158c0db2cc6a3` | 270 | Pin conflict with N02; do not share environment assumptions blindly |
| [Microsoft Agent Framework — N02 pin](https://github.com/microsoft/agent-framework) | N02 strategic-defense capability mapping | `f1fb145c4d80191aa86b3480d12112d828233763` | 287 | Pin conflict with N07 (17 commits apart); reconcile APIs and lock environments |
| [OpenHands](https://github.com/OpenHands/OpenHands) | N06 bounded software/tool execution | `2414d6ee5e31bede2e78211f72b58e9949575a75` | 368 | Adapter is a configured external session client; availability/configuration and execution need E2E proof |
| [Pydantic AI](https://github.com/pydantic/pydantic-ai) | N01 typed agents; SARA optional structured-agent augmentation | `6bc07cf18b0641ea92343d8c589cfb922108b802` | 908 | Large drift; SARA source boundary is projected and has no verified active runtime in this audit |
| [smolagents](https://github.com/huggingface/smolagents) | N06/N02 code-agent/tool augmentation | `c30b115286e000e98711fae5e85993547b73d826` | 1 | Near current; adapter path still needs live model/tool-result proof |
| [SuperAGI](https://github.com/TransformerOptimus/SuperAGI) | N07 autonomous agent facade / telemetry, delegated to native owners | `c3c1982e7bd6a11cfed53c5a193ea502f924b1b6` | 0 | Pin matches API default branch but repository has not seen a source push since Jan 2025; maintenance/security review before production use |
| [Superpowers](https://github.com/obra/superpowers) | Shared development methodology and skills, not a substitute for runtime inference | `8ca22dba9a94f28898bbce59f2537ff4d87c747d` | 3 | Markdown skill reader is a real bounded read path; skill use still must be audited per workflow |
| [Whisper](https://github.com/openai/whisper) | N03 local speech-to-text | `86098128c0b4f24f0e2aa2994de830614b474227` | 0 | Pin matches default branch; actual adapter requires enablement, source and a model weight path |

## N02’s ten additional unique sources

N02 records 13 sources. CrewAI, Microsoft Agent Framework and SuperAGI overlap the N07 canonical registry, leaving 10 unique projects below. The N02 triage manifest has 21 mappings, and every mapping is still `TRIAGED`; it does not prove an adapter transaction.

| Upstream repository | Assigned fit in N02 triage | Pinned SHA | Observed drift / legal/runtime note |
|---|---|---|---|
| [BRAIN](https://github.com/shivamtyagi18/BRAIN) | multimodal cortex (routed to N03) | `8d4201f72373156165ebc646c5064991fd3ad94a` | 0; small project, verify source content and tests before choosing a production role |
| [GENesis-AGI](https://github.com/WingedGuardian/GENesis-AGI) | autonomous embodiment | `a85f630a09c6cd38be1b5f8379ce309be41d0b72` | 361; autonomous runtime is not a substitute for SOUL’s approval, budget and Mesh policy gates |
| [nanobot](https://github.com/HKUDS/nanobot) | biomolecular-designer mapping (primary N06) | `d3d70f20655bce5d3c7a017bfbe803efe09a2438` | 450; personal-agent/workflow framework; domain-specific mapping is only triage until a typed handler and safety tests exist |
| [Syntra Kernel](https://github.com/gd2bk1ng/syntra_kernel) | reality synthesis | `755d8ceb0e81c3015329bfc0b01b01e606b44d7a` | 0; Rust cognitive OS, small project; runtime/language boundary and behavior evidence needed |
| [PraisonAI](https://github.com/MervinPraison/PraisonAI) | EUS | `bb470c786b63a6a8d5e5f3abcfc6080f1a8832f1` | 844; large API drift; no isolated runner and pinned dependency evidence observed |
| [Hermes Agent](https://github.com/NousResearch/hermes-agent) | UCI | `e8c97320ac8691d4de92af49f98459f9ef9ddb08` | Compare capped at 10,000 commits; reported ahead-by 11,871. Extremely high churn: isolate API and data migration before any update |
| [HASHIRU](https://github.com/HASHIRU-AI/HASHIRU) | strategic planning | `de2b514572dbad8625d974e5ff26587ab805502b` | 0; repository license metadata is missing. Runtime supports autonomous agent/tool administration and an option to bypass auth: block production execution until licensing, allowlists and mandatory auth are verified |
| [Codette Reasoning](https://github.com/Raiff1982/Codette-Reasoning) | ethical governance / Einstein reasoning | `1608092a2862ce8070210db37e57f04ac8717965` | 47; GitHub reports `NOASSERTION`, repository declares CSAL 1.0 non-commercial, not MIT; `BLOCKED_LICENSE` for commercial bundling pending review |
| [Chyren-Selin](https://github.com/Mega-Therion/chyren-selin) | existential safety / COT-AHRD | `dd5a633831e3779077736836c474a2dc75737a10` | GitHub API returned “Moved Permanently” for repo/compare; `BLOCKED_REF_RESOLUTION` pending canonical URL and pin revalidation. N02 registry claims Apache-2.0 but runtime/legal proof remains open |
| [Xun](https://github.com/MenxLi/xun) | UCI / emergent cognition | `129065cb94748d6a9e8cbe6b51d8f46d8bbeae80` | 168; repository license metadata is missing; `BLOCKED_LICENSE` for shipping/bundling until permission is verified |

## Cross-nucleus compatibility findings

### 1. Different pins of the same upstream
CrewAI and Microsoft Agent Framework appear in N07 and N02 at different commits. N07 and N02 must not run these projects on the assumption of a single API until the two pins are reconciled. This is a change-management conflict, not a reason to silently delete one pin. Preserve both current references during testing.

### 2. Catalog mismatches
- **N04 Browser Use:** the current main `integrations/capability-boundary.json` has an empty `upstream_capability_sources` array even though `.gitmodules` pins Browser Use, the adapter/runner/tests exist and the capability is named in N04 code. An additive change setting it to `["browser-use"]` was committed to N04’s isolated audit branch `audit/ai-nucleus-status-2026-10-25` at `db80711df855b3876d6f421ce008aa37200da817`; main is unchanged.
- **N07 LangGraph:** the generic adapter had a graph with one node that returned the input unchanged while emitting `PASS`. This could be mistaken for cognitive workflow success. The isolated fix branch now calls the real local `integrations/langgraph/graph.py` bridge, requires `N07_APP_TOKEN` and correlation, checks `langgraph==1.2.12`, and validates `metadata.correlation_id` returned by N07. New tests and a broader CI workflow have been added; expanded validation succeeded in GitHub Actions run `40191936673` at exact branch SHA `dc281511c15ddb19d16dd90f62b84a751aaba3c1`. This proves deterministic contracts and whole-repository Go tests, not a live provider or online multi-nucleus transaction.
- **N03 Whisper:** real adapter/runner code exists, but it requires explicit enablement, the pinned source root, a model weight path and device/dependencies. Do not claim speech-to-text is live before a real audio fixture succeeds.
- **N03 Kokoro:** adapter/runner exist, but `describeKokoroAdapter()` currently returns `DEGRADED / KOKORO_EXECUTION_NOT_YET_PROVEN` when source is present. Keep it not-live until model load and a real returned WAV/audio payload are verified.
- **N05 Piscina:** current `N05InferencePool.ts` is an in-process priority/concurrency queue and reports `piscinaEnabled: false`; `MESH_STATUS.md` mentions a Piscina path. This is a code/document mismatch to reconcile before calling external augmentation or worker-pool acceleration complete.
- **N02 additional sources:** registry and triage exist, but no generic N07 adapter is defined for these 10 unique sources. They remain `SOURCE_PINNED / TRIAGED`, not connected. Do not add them to the 16-source N07 canonical manifests without creating and testing adapters; the current validator requires the source, adapter and ownership IDs to match exactly.
- **SARA Pydantic AI:** the upstream is pinned and a boundary document exists, but the Python project has no verified dependency/import/adapter E2E for it. Keep `PROJECTED`; do not replace the ARA→ETR→ITR/regenerative cycle with an LLM wrapper.
- **N01 Letta Code and Pydantic AI:** source pins and an N07 adapter definition exist, but an imported Python/Node executable must be version-pinned and tested against the actual source SHA before counting the capability.

### 3. Dependency and runtime isolation
A gitlink SHA does not prove the runtime uses that code. Each adapter needs:
- explicit provider/package version or execution of the pinned source tree itself;
- isolated runtime/dependency lock (Python/Node/Cargo as appropriate);
- input/output schema and size bounds;
- explicit timeout/cancellation, retries and resource budgets;
- least-privilege capability allowlist and no implicit provider tool execution;
- mandatory authentication and redaction of secrets;
- deterministic tests with provider calls stubbed only for unit tests, plus a separate real-provider smoke test;
- evidence envelope identifying provider, source SHA, package version, owning nucleus, capability, result status and correlation ID.

### 4. License and supply-chain gates
Keep references in the source registry but block shipping execution until terms are verified for HASHIRU and Xun; Codette has a non-commercial CSAL 1.0; Chyren’s pinned source requires URL/commit resolution. For every upstream, check the license at the pinned SHA, transitive dependencies, model/weight licenses, subprocess privileges and network-egress permissions before bundling.

## Priority execution plan

**P0 — Freeze and record.** Do not change pinned submodule SHAs across the ecosystem yet. Re-read current default and feature branch heads before any later mutation. Preserve all 50 references and existing branches.

**P1 — Correct contract/evidence issues.** Run and finish validation for the isolated N07 LangGraph false-success correction; merge is not performed by this audit. Verify the N04 Browser Use catalog correction on its audit branch. Reconcile N05 pool/docs claims without removing existing provider paths.

**P2 — Native-owner source verification.** For the 16 canonical providers and 10 extra N02 sources, record exact SHA, actual imported/runtime package version, entrypoint, dependencies, license and intended native capability. Mark absent adapter as `BLOCKED_ADAPTER`, license as `BLOCKED_LICENSE`, unresolved source as `BLOCKED_REF_RESOLUTION`, and actual missing deployment configuration as `BLOCKED_ENV` only after a specific transaction establishes the dependency.

**P3 — Capability-by-capability proof.** Start with high-confidence, narrow augmenters that already have adapter code: N03 Whisper transcription / Kokoro synthesis, N04 Browser Use automation with domain allowlists, N05 retrieval with supplied documents, N01 typed agent/memory helper, N06 DSPy/smolagents/OpenHands in bounded contexts, N07 actual LangGraph and CrewAI workflows. Run one native capability per owner with real result validation, then verify its Mesh round-trip. Do not activate several providers at once.

**P4 — Resolve repeated framework pins.** Choose compatible pinned revisions only after checking adapters/tests for CrewAI and Microsoft Agent Framework across N02/N07. Record why a pin advances and what API/tests prove it. No broad batch update.

**P5 — Connect N02’s 21 triaged function mappings.** Each mapping must point to one primary native owner plus secondary collaborators; add a typed adapter or mark it explicitly blocked. Keep all mapping rows and provenance; do not turn a triage record into “connected” by status-only edits.

**P6 — Verify the full SOUL path.** For each successful augmentation, record source nucleus -> Mesh -> destination owner -> upstream adapter -> real result -> authenticated return -> unchanged correlation ID. Only that qualifies for `RUNTIME_VERIFIED`; upstream source availability or CI alone does not.

## Evidence and limitations
- All reported gitlinks were inspected from the exact repository main tree SHAs listed in the SOUL AI status records. Each reference preserves its own commit pin.
- GitHub compare was queried for the public upstreams; the drift numbers above are commit counts relative to each upstream’s default branch at audit time, not tested compatibility statements.
- No global submodule updates were made. No public upstream code was changed.
- No PR was opened or merged. Main branches were not modified by this master document.
- The N07 LangGraph corrective branch is under execution/CI; use its latest actual status rather than assuming its older Go-only run validates the new Python workflow.
