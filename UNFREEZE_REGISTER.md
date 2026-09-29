# SOUL — Unfreeze Register

## Purpose

This register separates implementation eligibility from runtime/release promotion.

A phase can be UNFROZEN FOR IMPLEMENTATION while its production gate remains NOT OPEN. This is the state required to finish pending foundations without pretending that an uncommissioned runtime is online.

## Current state

| Area | Implementation state | Runtime gate | Required closure |
|---|---|---|---|
| Phase 1 — Mesh | ACTIVE | PARTIAL | live peer URLs + HMAC + real transactions |
| Phase 2 — Federation | UNFROZEN FOR IMPLEMENTATION | NOT OPEN | six/seven-core smoke evidence |
| Phase 3 — Memory/Handoff | UNFROZEN FOR IMPLEMENTATION | NOT OPEN | durable recall + cross-runtime handoff proof |
| Phase 4 — Planner / Goal→Plan→Act | UNFROZEN FOR IMPLEMENTATION | NOT OPEN | production registration + real execution |
| Phase 5 — Learning / RGO | UNFROZEN FOR IMPLEMENTATION | NOT OPEN | failure→finding→correction→test cycles |
| Phase 6 — Super AGI gates | UNFROZEN FOR FOUNDATION WORK | NOT OPEN | all previous gates + world state + metacognition + governance evidence |
| OpenAI agent integration | UNFROZEN FOR ADAPTER WORK | NOT OPEN | server-side integration + exact-head CI + runtime proof |
| Realtime / multimodal | UNFROZEN FOR ADAPTER WORK | NOT OPEN | N03/native provider proof |
| Vision | UNFROZEN FOR ADAPTER WORK | NOT OPEN | native owner + executable adapter + evidence |
| Sandbox | UNFROZEN FOR DESIGN/IMPLEMENTATION | NOT OPEN | isolated production boundary + audit |
| 72-nodule inventory | UNFROZEN FOR AUDIT | NOT OPEN | source enumeration before node-level verification |
| Clareira | UNFROZEN FOR CONTINUOUS EVOLUTION | PROTECTED | evolve only additively and with source/runtime evidence |

## Rules

- UNFROZEN does not mean ONLINE.
- NOT OPEN does not mean abandoned.
- BLOCKED means a dependency must be supplied or repaired.
- NOT MEASURED means the transaction has not been observed.
- No phase may close by documentation alone.
- No native module may be removed to make a later phase pass.
- New functionality must use the canonical Mesh/authority path.

## Work order

FOUNDATION → FEDERATION → MEMORY → PLANNING → LEARNING → SUPER AGI GATES

Within every phase:

PRESERVE → AUDIT → CORRECT → CONNECT → TEST → VALIDATE → DOCUMENT → ADVANCE

## Current unfreeze decision

The downstream phases are no longer treated as frozen for engineering work. They remain gated for production activation until their prerequisites have actual evidence.