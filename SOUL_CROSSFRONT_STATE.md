# SOUL Crossfront State — 2026-10-06

## Authority

This branch is a non-destructive SOUL v1 restructuring line. Git history and existing mainline implementations remain preserved.

Evidence vocabulary:
- REAL = runtime execution observed by an executable test or E2E transaction.
- PROJECTED = architecture/manifest/adapter exists but runtime execution is not proven.
- BLOCKED = required runtime/configuration is absent or failed.
- UNMEASURABLE = evidence cannot be evaluated under current conditions.

## Native SOUL graph

| Node | Repository | Role |
|---|---|---|
| N01 | divibisoul/aeternum-core-29 | coordination, state, identity, memory |
| N02 | divibisoul/Eternium- | multi-agent generation and collaboration |
| N03 | divibisoul/nexus-aeternum-fusion | audio/perception/speech |
| N04 | divibisoul/nextjs-ai-chatbots | conversational/tool interaction |
| N05 | divibisoul/nextjs-ai-chatbot | chat, knowledge, retrieval |
| N06 | divibisoul/nextjs-ai-chatbot-2000 | cognition, reasoning, agents |
| N07 | divibisoul/Orquestrador- | orchestration, federation, control plane |
| SARA | divibisoul/SARA | regeneration, governance, resilience |
| JEV | divibisoul/jev-api | typed decisions and guardrails |

## Foundation status

### FASE 1

- N01 dual Mesh contract reconciliation: IMPLEMENTED / PROJECTED until CI executes.
- N01 → N07 real federation E2E workflow: IMPLEMENTED / PROJECTED until CI evidence completes.
- N07 GRCE real SARA boundary: IMPLEMENTED; runtime REAL only under integration environment.
- N02/N03/N05 foundation CI gates: implemented on isolated branches.
- N03 package-lock is present in current main; historical notes claiming it is absent are stale.

### FASE 2

GRCE core:
- six hooks: detect, characterize, regenerate, validate, freeze, trace.
- canonical GoldenRuleParticipant contract exists.
- real Go → HTTP SARA participant exists.
- external providers have bounded participant adapters and explicit configuration gates.

SOUL-28 provider boundaries:
- bijux-core → GRCE characterize.
- ouro-loop → GRCE validate.
- Recuris → GRCE trace.

No provider is promoted to ONLINE by manifest presence alone.

### FASE 3

SOUL-29 learning conduits are registered:
- FedML → federated learning provider boundary.
- Hivemind → decentralized learning implementation dependency.
- persistence boundary: N01 HortaCore.
- strategy boundary: N07 Prefrontal.
- transport boundary: SARA NervoVago.

No federated-training runtime is declared REAL until a real configured transaction is observed.

### FASE 4

NervoVago is the canonical SARA-side logical name for the existing VagusNerveBus implementation. VagusBus remains a compatibility alias; no second bus implementation is introduced.

N07 release gate:
- ONLINE: NO.
- Reason: complete runtime proof for all required gates is not yet present.
- CI queued/running: YES.

## Current graph arithmetic

- Native SOUL nodes: 9
- External capability providers: 20
- SOUL-29 graph nodes: 29
- Hivemind: implementation dependency, graph_node=false
- Canonical Mesh protocol: soul-mesh/1
- Contract: 1.1.0

## Non-elimination rule

No files, branches, upstream source history, previous contracts, or existing runtime authorities are deleted by this restructuring line.
