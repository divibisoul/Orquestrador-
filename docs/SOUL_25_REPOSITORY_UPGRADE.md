# SOUL — 25-Repository Upgrade Fabric

## What changed

SOUL now treats the 9 native repositories and the 16 attached upstream repositories as one **capability federation of 25 repositories**.

The purpose is not to merge 25 codebases into one runtime. The purpose is to let the 9 native SOUL components consume complementary capabilities from the 16 upstream sources through explicit contracts, adapters, the canonical Mesh and evidence gates.

## The 9 native SOUL components

N01 keeps coordination, state, identity and memory authority.

N02 keeps multi-agent execution and collaboration authority.

N03 keeps audio perception, speech and multimodal voice authority.

N04 keeps conversational interface, tools, artifacts and documents authority.

N05 keeps inference, chat, knowledge and retrieval authority.

N06 keeps cognition, reasoning, planning and agent-execution authority.

N07 remains the canonical federation, routing, SuperGPU, Clareira and Octacore control plane.

SARA keeps regeneration, governance, provenance, rollback and resilience authority.

JEV keeps typed decision and guardrail authority.

## The 16 upstream capability sources

Superpowers, SuperAGI, LangGraph, CrewAI, Microsoft Agent Framework, OpenHands, MetaGPT, AgentScope, Letta Code, Browser Use, smolagents, Pydantic AI, LlamaIndex, DSPy, Whisper and Kokoro.

Their original repositories, licenses and maintainers remain authoritative.

## Upgrade path

Each capability follows:

**upstream source → provenance → capability contract → adapter → canonical Mesh/N07 routing → native SOUL owner → controlled execution → verification → observability → re-audit**

A provider can be directly affine to a nucleus and also remain globally discoverable through N07.

## Evidence

**REAL** means the structure/contract exists and the corresponding CI or source-level integration has verified it.

**PROJECTED** means the integration path is specified and addressable, but real external runtime execution still requires the provider adapter, credentials/configuration and an end-to-end result.

**BLOCKED** means execution is deliberately prevented because required infrastructure is absent.

**UNMEASURABLE** means the available environment cannot establish runtime truth.

No status is upgraded by documentation alone.

## Current runtime surface

N07 exposes:

- `soul.capability.upgrade.describe@1.0.0`
- `soul.capability.upgrade.resolve@1.0.0`
- `GET|POST /v1/capability-upgrade`

The resolver returns all 16 upstream sources and marks direct-affinity sources for a requested native component/capability.

## Engineering agents

The upgrade fabric is maintained by:

- `soul.engineer.fabric`
- `soul.engineer.provenance`
- `soul.engineer.adapter`
- `soul.engineer.runtime`
- `soul.engineer.verification`
- `soul.engineer.resilience`

Superpowers provides methodology at explicit boundaries through the previously integrated crossfront agents.

## Safety of integration

Native SOUL functions are preserved. External providers do not become new SOUL nuclei, do not receive Mesh authority, and do not silently replace a native capability.

Runtime success is never synthesized when a provider is unavailable.
