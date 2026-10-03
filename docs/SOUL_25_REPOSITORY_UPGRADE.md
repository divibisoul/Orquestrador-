# SOUL — 25-Repository Upgrade Fabric

## What changed

SOUL now treats the 9 native repositories and the 25 attached upstream capability sources as one **capability federation of 25 repositories**.

The purpose is not to merge 25 codebases into one runtime. The purpose is to let the 9 native SOUL components consume complementary capabilities from the 25 upstream sources through explicit contracts, adapters, the canonical Mesh and evidence gates.

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

## The 25 upstream capability sources

Superpowers, SuperAGI, LangGraph, CrewAI, Microsoft Agent Framework, OpenHands, MetaGPT, AgentScope, Letta Code, Browser Use, smolagents, Pydantic AI, LlamaIndex, DSPy, Whisper, Kokoro, ECC, SwarmClaw, Mem0, Letta, Langfuse, vLLM, SGLang, Ray and Megatron-LM.

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

The resolver returns all registered upstream sources and marks direct-affinity sources for a requested native component/capability.

## Current 34-repository state\n\nThe historical 25-repository name is preserved for continuity; the current composition is 9 native + 25 upstream = 34 repositories.\n\n## Engineering agents

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

## Native-function reinforcement layer

The 25 upstream sources are now mapped against concrete native capability families in `integrations/native-capability-augmentation.json`.

This creates a reinforcement relationship for the existing 9 components:

**native capability -> compatible upstream providers -> canonical N07 resolver -> explicit adapter -> evidence-gated execution**

The resolver does not replace the native implementation. It identifies which upstream capability sources can strengthen the native function and which of those sources have an adapter boundary that is actually available.

At the current stage, all 25 repository sources are structurally integrated and the augmentation graph is real; external provider runtime execution remains PROJECTED until provider-specific runtime adapters and end-to-end evidence exist.
