# SOUL Ecosystem

**SOUL component:** N07 — Orquestrador, federation and external capability control plane

This repository is the **canonical public integration point for the SOUL ecosystem**. It preserves the native SOUL nuclei and records the external open-source capability sources attached by functional similarity.

## External capability ecosystem
The canonical registry is:

**integrations/external-capabilities.json**

It identifies source repository, pinned revision, target SOUL nucleus and declared capabilities. The attached source repositories remain maintained by their original upstream authors.

## SOUL repositories
- N01: https://github.com/divibisoul/aeternum-core-29
- N02: https://github.com/divibisoul/Eternium-
- N03: https://github.com/divibisoul/nexus-aeternum-fusion
- N04: https://github.com/divibisoul/nextjs-ai-chatbots
- N05: https://github.com/divibisoul/nextjs-ai-chatbot
- N06: https://github.com/divibisoul/nextjs-ai-chatbot-2000
- N07: https://github.com/divibisoul/Orquestrador-
- SARA: https://github.com/divibisoul/SARA
- Jev API: https://github.com/divibisoul/jev-api

## Functional integration boundary

The binding contract for this component is recorded in `integrations/capability-boundary.json`. It states why each upstream capability is present, the canonical routing boundary, the engineering agent responsible, and the evidence gate before runtime activation.

## 25-repository capability upgrade

This component participates in the shared SOUL 25-repository capability fabric. The local binding is recorded in `integrations/soul-25-augmentation.json`; external capabilities are consumed through the canonical N07 federation and remain evidence-gated.

## 25-repository upgrade fabric

Architecture and functional roles are documented in `docs/SOUL_25_REPOSITORY_UPGRADE.md`. N07 exposes the canonical capability-upgrade surface at `/v1/capability-upgrade`, with runtime activation remaining evidence-gated.

## Runtime truth
Git links and registry entries establish structural integration. Runtime activation is reported only when an adapter, configuration and end-to-end verification exist.
