# SOUL Agent Arsenal Gateway

N07's Agent Arsenal Gateway integrates the three upstream ecosystems identified in the external research:

- `obra/superpowers`: agentic development skills, parallel subagent dispatch and verification workflows.
- `affaan-m/ECC`: Everything Claude Code assets including agents, skills, hooks, commands, MCP configuration and harness tooling.
- `ruvnet/ruflo`: executable multi-agent/swarm runtime, agent spawning, coordination, memory and MCP surfaces.

## Provenance

All upstream sources are checked out at exact immutable Git commits:

| Source | Repository | Commit |
| --- | --- | --- |
| Superpowers | `obra/superpowers` | `8ca22dba9a94f28898bbce59f2537ff4d87c747d` |
| ECC | `affaan-m/ECC` | `c05b2d6614f62f6db0047669aa4eefb223d478f9` |
| Ruflo | `ruvnet/ruflo` | `27982983ea6cdc4767c0b6614a4ad9a9d9497cce` |

The gateway never silently replaces upstream source with a mutable branch.

## SOUL capability surface

N07 registers:

- `agent.arsenal.inventory@1.0.0`
- `agent.arsenal.catalog@1.0.0`
- `agent.arsenal.resolve@1.0.0`
- `agent.arsenal.swarm@1.0.0`
- `agent.arsenal.agent.spawn@1.0.0`

The catalog inventories every checked-out file and classifies it as `agent`, `skill`, `hook`, `command`, `mcp`, `plugin`, `workflow` or `source`. Pagination permits the whole upstream trees to be inspected without a fixed static copy.

## Runtime boundary

Superpowers and ECC are primarily Claude Code/harness assets rather than independent model runtimes. SOUL therefore connects their actual artifacts, instructions and configuration surfaces to its discovery and retrieval plane.

Ruflo is the executable multi-agent backend exposed by the gateway. N07 can request swarm orchestration and agent spawning through the typed operations above. The Ruflo CLI is version-pinned to `3.50.0` and execution is disabled by default in the gateway image.

Production deployment explicitly enables Ruflo execution and supplies a shared bearer token between N07 and the sidecar. The gateway has no public host port.

## Files

- `gateway.mjs`: pinned source provisioning, inventory, artifact resolution and bounded Ruflo execution.
- `bootstrap.sh`: deterministic local/CI provisioning of the three upstream repositories.
- `Dockerfile`: isolated Node runtime for the gateway.
- `package.json`: Node runtime metadata.
