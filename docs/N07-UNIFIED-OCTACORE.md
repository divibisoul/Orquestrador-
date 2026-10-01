# N07 Unified Executive OctaCore

Revision: 2026-10-01

## Canonical composition

The existing N07 Octacore processor now contains a single unified logical execution core named `N07.executive-octacore`.

It composes four existing authorities without creating a new nucleus:

1. Orchestrator — route validation, correlation and execution control.
2. Prefrontal Neocortex — neural-signal evaluation, risk inhibition and decision commitment.
3. Orbital Reasoning / Transcendental Compute Engine — deterministic resource estimation; simulation only.
4. SuperGPU — real bounded execution through the existing runtime and device backend.

The logical topology is eight lanes: two control/execution lanes for each of the four component authorities. This is a software composition, not a claim of physical eight-core silicon.

## Mesh contract

The unified core is exposed through the canonical N07 Mesh ingress:

- `octacore.core.describe@1.0.0`
- `octacore.core.health@1.0.0`
- `octacore.core.execute@1.0.0`

A request preserves the original Mesh `correlationId` through the entire execution chain.

The external-agent bridge uses the existing N02 canonical adapters. SuperAGI-derived `mlfg` and `skill_acquisition`, Xun-derived `emergent_cognition`, Hermes-derived `uci`, and CrewAI-derived `scre` are invoked only through canonical Mesh capabilities; no upstream runtime is copied into N07.

## Execution order

`orchestrator.resolve -> external-agent advisory (optional) -> orbital.reasoning.estimate -> prefrontal.neocortex.admission -> supergpu.execution`

A failed required collaborator blocks the composite request. Optional collaborator failures remain explicit in the returned evidence.

## Evidence semantics

`REAL` means the execution path and exact capability contract were reached.

`PROJECTED` means a reachable target exists but does not advertise the exact executable capability.

`BLOCKED` means policy/configuration prevents the execution.

`UNMEASURABLE` means the transport or capability evidence could not be established.

TCE output is always marked simulation-only. SuperGPU hardware acceleration is reported from the actual backend state; unsupported GPU hardware is not fabricated.

## Upstream agent utilization

The strongest multi-agent contribution currently wired into this core is the SuperAGI family:

- `mlfg` -> meta-learning design advisory -> N02 adapter -> Mesh -> N07 core.
- `skill_acquisition` -> reusable skill extraction from demonstrations -> N02 adapter -> Mesh -> N07 core.

This preserves N02 canonical adapter ownership while allowing the new core to use those functions cooperatively.
