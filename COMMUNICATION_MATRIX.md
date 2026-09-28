# SOUL Communication Matrix — Phase 1

## Scope

This is the current evidence matrix for communication among N01–N07 and SARA.

A cell describes an **observed communication path**, not merely the existence of code. Source-level capability declarations do not become PASS.

Allowed states:

- **PASS** — a real transaction was executed and evidence is preserved.
- **FAIL** — a real transaction executed and failed.
- **BLOCKED** — the transaction cannot currently be commissioned because a required dependency/configuration is missing.
- **NOT MEASURED** — no real end-to-end transaction has yet been observed.
- **LOCAL** — same-runtime/local operation; not a federated edge.

All off-diagonal cells are directional: A→B is separate evidence from B→A.

## Current matrix

| From \ To | N01 | N02 | N03 | N04 | N05 | N06 | N07 | SARA |
|---|---|---|---|---|---|---|---|---|
| **N01** | LOCAL | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED |
| **N02** | NOT MEASURED | LOCAL | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED |
| **N03** | NOT MEASURED | NOT MEASURED | LOCAL | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED |
| **N04** | NOT MEASURED | NOT MEASURED | NOT MEASURED | LOCAL | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED |
| **N05** | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | LOCAL | NOT MEASURED | NOT MEASURED | NOT MEASURED |
| **N06** | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | LOCAL | NOT MEASURED | NOT MEASURED |
| **N07** | NOT MEASURED | NOT MEASURED | NOT MEASURED | BLOCKED | BLOCKED | BLOCKED | LOCAL | NOT OPEN |
| **SARA** | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT MEASURED | NOT OPEN | LOCAL |

### Blocked commissioning edges

The current Phase-1 N07 real integration gate explicitly commissions:

- N07 → N04: native core.health
- N07 → N05: native core.health
- N07 → N06: native support.mesh

The commissioning workflow now creates its own ephemeral HMAC secret and public Quick Tunnel URLs, so external Mesh URL/secrets are no longer a prerequisite for this CI path.

The three edges remain BLOCKED / NOT MEASURED until a real workflow run executes the transactions successfully against the checked-out peer runtimes. The status is deliberately not upgraded by code inspection or workflow definition alone.

### Why reverse directions remain NOT MEASURED

The current integration test is an N07-originated transaction. It does not prove the reverse directional path, so reverse cells remain NOT MEASURED rather than being inferred from symmetric source code.

### SARA

N07↔SARA handoff remains **NOT OPEN**. N07 already contains SARA proxy/capability paths in its broader runtime surface, but a live transaction is not claimed without commissioning evidence.

### SARA capability handoff status

Source evidence confirms that N07 has native SARA proxy/capability paths, but this is **source-level availability, not live federation evidence**.

| Capability | Source status | Live N07↔SARA transaction |
|---|---|---|
| sara.cycle@1.0.0 | REGISTERED WHEN SARA CONFIGURED | NOT OPEN |
| sara.audit@1.0.0 | REGISTERED WHEN SARA CONFIGURED | NOT OPEN |
| sara.regenerate@1.0.0 | REGISTERED WHEN SARA CONFIGURED | NOT OPEN |
| sara.state@1.0.0 | REGISTERED WHEN SARA CONFIGURED | NOT OPEN |
| sara.capabilities@1.0.0 | REGISTERED WHEN SARA CONFIGURED | NOT OPEN |
| sara.trace@1.0.0 | REFERENCED BY N07 SARA/backend surface | NOT OPEN |

Required configuration for live SARA handoff is SARA_SERVICE_URL plus SARA_SERVICE_TOKEN. These values are not invented or embedded in this branch.

The SARA path must continue through the canonical Mesh/PeerClient authority where applicable; no second public Mesh or replacement SARA core is permitted.

## Internal-subsystem preservation

The matrix measures federation edges; it does not replace the inventories inside each nucleus.

For every future phase, communication evidence must be accompanied by a preservation/evolution check covering the native internal systems of the participating nucleus. This includes Clareira and its ProcessingNodes/InformationChannels/HomeostasisManager/EventBus path in N01, plus the native agents, tools and capabilities of all other nuclei.

The broader architecture's 72 nódulos are protected scope. Their per-node status must only be marked after source-backed enumeration; no unproven PASS is permitted.

## Evidence record format

Each upgraded cell should carry:

`STATE | date | source HEAD | workflow/run | capability | correlationId | latency/error`

Do not fabricate example values in an actual status cell.

## Gate

Phase 1 communication is not fully commissioned while required real edges remain BLOCKED or NOT MEASURED.
