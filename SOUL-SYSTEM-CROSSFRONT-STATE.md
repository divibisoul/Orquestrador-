# SOUL — Crossfront State (Generated Baseline)

> N07 governance baseline. Repository evidence is reported separately from live runtime commissioning. Canonical Soul Mesh contract target: **1.1.0**.

| Nucleus | Repository | Contract target | Mesh role | Neural layer status | Runtime commissioning |
|---|---|---:|---|---|---|
| N01 | `divibisoul/aeternum-core-29` | 1.1.0 | Gateway / Android peer | Local bridge + existing Mesh runtime | External runtime required |
| N02 | `divibisoul/Eternium-` | 1.1.0 | Cognition / inference | Synaptic bridge + Mesh runtime | External runtime required |
| N03 | `divibisoul/nexus-aeternum-fusion` | 1.1.0 | Audio / speech | Synaptic bridge + Mesh runtime | External runtime required |
| N04 | `divibisoul/nextjs-ai-chatbots` | 1.1.0 | Documents / tools | Synaptic bridge + Mesh runtime | External runtime required |
| N05 | `divibisoul/nextjs-ai-chatbot` | 1.1.0 | Conversation | Synaptic bridge + Mesh runtime | External runtime required |
| N06 | `divibisoul/nextjs-ai-chatbot-2000` | 1.1.0 | Cognitive support | Synaptic bridge + Mesh runtime | External runtime required |
| N07 | `divibisoul/Orquestrador-` | 1.1.0 | Neocortex / federation / SuperGPU control plane | Neural federation + adaptive protocol capability | External runtime required |

## Current governance capabilities

- N01 peer client routes to N02–N07 and validates `correlationId`, protocol, contract and route identity.
- N02 exposes the canonical existing `api/soul-mesh.ts`; the compatibility adapter accepts legacy action envelopes without replacing the secured Mesh runtime.
- N02–N06 expose local `SynapticNodeBridge` modules over the existing Soul Mesh.
- N07 contains adaptive protocol translation and federation controls.
- Integrity verifiers are present in all seven nuclei and report correlation, response validity and latency.
- External deployment URLs/secrets are never fabricated; unavailable commissioning remains explicitly uncommissioned.

## Automatic governance

N07 now owns `scripts/generate-crossfront-state.mjs` and `.github/workflows/crossfront-state-sync.yml`. The workflow checks out N01–N07 repository state, regenerates this report and publishes changes only when the generated content differs.

## Acceptance rule

A repository-level green state does not prove live seven-runtime connectivity. Final ONLINE status requires successful authenticated real-runtime federation/E2E evidence through the existing N07 commissioning gates.
