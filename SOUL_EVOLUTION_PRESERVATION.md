# SOUL — Conservation and Uniform Evolution Contract

## Purpose

This contract adds a preservation invariant to Phase 1: improving the communication substrate must raise the system's integration level **without lowering the integrity or capability of any existing internal subsystem**.

The Mesh is the communication medium. It is not a license to flatten, hide, disable, rename, or replace the systems carried by that medium.

## Conservation invariant

For every nucleus N01–N07 and for SARA:

- Native runtime ownership remains intact.
- Existing modules, agents, tools, capabilities, data paths and interfaces remain part of the system unless a separate, evidence-backed compatibility decision explicitly retires one.
- Integration work is additive/minimal-diff whenever possible.
- A new capability must connect to the existing canonical bus/authority rather than creating a parallel hidden bus.
- Observability must preserve enough identity to tell which nucleus, subsystem, capability and correlation produced an observation.
- A failure in one subsystem becomes a finding to correct; it is not evidence that the subsystem should be removed.
- "Not measured" and "blocked" are valid evidence states; they are never replaced by a declaration of readiness.

## Uniform-evolution rule

"Grow equally" means **no layer is left behind functionally**. It does not mean every subsystem receives identical code or identical performance targets.

When a federation phase advances:

1. The communication contract advances.
2. Each nucleus keeps its native internal systems active and compatible with that contract.
3. Existing subsystems receive the minimum bridge/adapter/telemetry changes required to participate.
4. New cognitive or orchestration behavior is not allowed to become a second implementation of functionality already owned by a nucleus.
5. The phase evidence must identify what was preserved, what was adapted, what was measured, and what remains blocked.

## Protected internal scope

The protected scope includes, at minimum:

- N01: Projeto Clareira, its neural processing nodes, InformationChannel, HomeostasisManager, NucleoRaizAlma, EventBus/bridge, ModuleRegistry and other native runtime systems.
- N02–N06: their native agents, tools, capabilities, runtimes, registries, memory/context systems and transport adapters.
- N07: Mesh, PeerClient, orchestration, routing, SuperGPU, observability, storage and governance paths.
- SARA: its independent Python runtime and regenerative/governance authority.
- The broader architecture's **72 nódulos** named in the system design are treated as protected scope. Their individual enumeration is a separate evidence task; this contract intentionally does not invent a node-by-node inventory that has not been proven from source.

## Clareira-specific rule

Clareira is an active subsystem, not a decorative UI and not disposable transport glue.

Phase 1 may add Mesh/telemetry integration around Clareira, but must not replace its internal processing graph, homeostasis, channels, event semantics or native ownership with a simplified substitute.

At the current N01 source snapshot, ProjetoClareira constructs a real internal graph of one central root plus primary and secondary ProcessingNodes. That concrete source evidence is preserved separately from the broader 72-node architectural inventory.

## Change classification

Every integration change must be classifiable as one of:

- **PRESERVE** — existing behavior retained unchanged.
- **ADAPT** — existing behavior retained with a compatibility bridge, telemetry, authentication or routing adaptation.
- **EXTEND** — new additive capability connected through the canonical contract.
- **RETIRE-WITH-REPLACEMENT** — allowed only with evidence of the old behavior, explicit replacement, compatibility analysis, regression coverage and documentation.

Unclassified removal of native functionality is a release defect.

## Phase-1 consequence

Phase 1 remains about real communication proof. This contract does not open Planner, Working Memory, RGO or a Super AGI gate.

The next phase may consume the proven Mesh channel, but it must inherit this conservation invariant and may not create a second communication substrate.

## Evidence

A phase handoff should record:

- exact source HEAD;
- changed and preserved components;
- native capabilities touched;
- communication paths measured;
- correlation/trace evidence;
- failures and diagnostics;
- compatibility assumptions;
- tests and CI runs;
- explicitly unmeasured or blocked areas.

No component is considered "evolved" merely because a document lists it.
