# SOUL Agent Arsenal

Bounded N07 sidecar for external agent ecosystems referenced by the SOUL integration contract.

## Authority

N07 remains the single canonical orchestration and federation control plane. This sidecar does not expose a second SOUL Mesh or a second orchestrator.

- Superpowers: development methodology and skills.
- ECC: engineering harness and skills.
- Ruflo: external execution provider, only through the pinned checkout.

## Provenance pins

- Superpowers: obra/superpowers@8ca22dba9a94f28898bbce59f2537ff4d87c747d
- ECC: affaan-m/ECC@c05b2d6614f62f6db0047669aa4eefb223d478f9
- Ruflo: ruvnet/ruflo@27982983ea6cdc4767c0b6614a4ad9a9d9497cce

The gateway verifies each checkout with git rev-parse HEAD before exposing artifacts. Ruflo execution uses the checked-out bin/cli.js; it does not use floating npx packages.

## State semantics

Process health (/health) is not execution proof. The default is AGENT_ARSENAL_EXECUTE=false.

Structural attachment is PROJECTED until the sidecar endpoint is configured and a real authenticated execution path produces evidence. Missing configuration remains BLOCKED/DEGRADED, never synthetic PASS.

## Configuration

N07:
- SOUL_AGENT_ARSENAL_URL
- SOUL_AGENT_ARSENAL_TOKEN
- SOUL_AGENT_ARSENAL_TIMEOUT

Gateway:
- AGENT_ARSENAL_ROOT
- AGENT_ARSENAL_TOKEN
- AGENT_ARSENAL_EXECUTE
- AGENT_ARSENAL_COMMAND_TIMEOUT_MS
- AGENT_ARSENAL_INSTALL_TIMEOUT_MS

## Operations

N07 exposes:
- agent.arsenal.inventory@1.0.0
- agent.arsenal.catalog@1.0.0
- agent.arsenal.resolve@1.0.0
- agent.arsenal.activate@1.0.0
- agent.arsenal.swarm@1.0.0
- agent.arsenal.agent.spawn@1.0.0

The Go proxy preserves trace and correlation identifiers and returns upstream failures as errors.

## Deployment boundary

The sidecar fetches pinned upstream source into its own volume. External source trees are not copied into the N07 Go runtime and do not receive canonical capability ownership.
