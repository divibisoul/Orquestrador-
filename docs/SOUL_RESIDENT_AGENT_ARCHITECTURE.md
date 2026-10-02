# SOUL Resident Agents + External Capability Federation

## Purpose

SOUL keeps one canonical authority per responsibility. External OSS repositories provide implementation options or engineering methodology; they do not become parallel control planes.

The integration path is:

human/CI -> coding agent + Superpowers skills -> nucleus Resident Agent -> soul-mesh/1 (1.1.0) -> adjacent peer -> N07 federation/control plane -> explicit OSS adapter -> evidence -> re-audit.

## Resident Agent decision

The safe default is an **embedded-local worker**, not seven new always-on network daemons. Each N01-N07 process owns exactly one Resident Agent object/worker bound to the nucleus runtime. The worker is discoverable through Mesh but reuses the nucleus HTTP/runtime boundary.

This avoids creating an eighth control plane, avoids process explosion, and keeps failure isolated to the owning nucleus. A future deployment may host the worker as a separate process only when a concrete isolation requirement is demonstrated and the same Mesh contract is retained.

## Auth and storage

Resident Agents reuse the existing nucleus Mesh authenticator. HMAC is the primary mode; Bearer remains only where an existing nucleus route already requires it. Each nucleus keeps its existing secret configuration. The contract is fail-closed and introduces no new secret namespace.

Plans are persisted in nucleus-owned state keyed by nucleus + correlationId and are not Git-tracked. Test/runtime evidence uses the existing CI artifact/runtime evidence path with schema `soul-evidence/1`. Resident Agents do not commit these records into repository history.

## Seven agents

| Nucleus | Agent | Native owner | Superpowers focus | Published scope |
|---|---|---|---|---|
| N01 | Agent-Clareira/Gateway Steward | N01 | systematic-debugging, verification-before-completion, requesting-code-review | mesh.health, mesh.discovery, registry, correlation continuity |
| N02 | Agent-Neural/Inference Steward | N02 | brainstorming, verification-before-completion | inference.*, model-routing evidence |
| N03 | Agent-Audio Steward | N03 | TDD, systematic-debugging, verification | audio.*, speech.*, adapter health |
| N04 | Agent-Tools Steward | N04 | writing-plans, executing-plans, verification | tools.*, documents.*, artifacts.* |
| N05 | Agent-Conversation Steward | N05 | subagent-driven-development, review, verification | chat.*, dispatch, retrieval/browser boundary |
| N06 | Agent-Cognition Steward | N06 | writing-plans, TDD, verification | composition.*, planning, summarize, session |
| N07 | Agent-Orchestrator Steward | N07 | dispatching-parallel-agents, subagent-driven-development, review | execute, federation, routing, OSS capability facade |

SARA receives its own **Agent-Regen** because SARA is a transversal canonical authority; it is not counted as N08.

## Superpowers boundary

Superpowers remains a development methodology/skill pack pinned to SHA `8ca22dba9a94f28898bbce59f2537ff4d87c747d`.

It is not exposed as a HTTP runtime, not copied as a second Mesh, and not used as a policy engine. Coding agents that modify a nucleus use the same workflow: plan -> TDD -> execute -> review -> verify.

## Capability ownership rule

The registry `integrations/external-capabilities.json` remains unchanged. `integrations/external-capability-ownership.json` adds the missing distinction:

- **primaryOwner** = canonical SOUL nucleus responsible for the capability slice;
- **consumers** = nuclei allowed to request/consume the capability;
- **affinity** = documented reason the upstream is useful there;
- **status** = structural-only, adapter-planned, or adapter-live.

No upstream project can be the authority for more than one SOUL responsibility merely because it contains multiple agent abstractions.

## Answers to engineering questions

1. **Resident lifecycle:** default = embedded-local worker bound to each nucleus process; no new always-on daemon fleet. ASSUMPTION chosen to minimize blast radius.
2. **N05 agent:** no UX and no competition with the Chat SDK. It is a Mesh/capability operator behind the existing N05 route.
3. **Multi-agent software process:** canonical facade/control remains N07; MetaGPT/CrewAI/OpenHands/SuperAGI are backends. N06 owns local cognitive planning/execution where its native capability boundary is authoritative.
4. **SARA:** gets Agent-Regen, as already being integrated by SARA PR35. It remains transversal, not N08.
5. **Superpowers runtime role:** development skill pack only. No runtime policy-engine import.
6. **Mandatory skills:** writing-plans, test-driven-development, systematic-debugging, verification-before-completion. Role-specific skills are additive, never substitutes.
7. **Versioning:** pin upstream Superpowers SHA; do not fork unless a concrete incompatibility is demonstrated. ASSUMPTION.
8. **Health/discovery:** N05 canonical entry is `/api/soul-mesh` GET/POST. N07 canonical control entry is the existing Mesh endpoint plus `api/health/dashboard.go` for aggregate probes. The crossfront scanner must inspect source entrypoints first and runtime health separately.
9. **Secrets:** retain current `SOUL_MESH_HMAC_SECRET` for compatibility. Future directional/per-edge secrets are a later schema change, not a silent retrofit.
10. **Correlation:** ingress/N07 creates correlationId when absent; resident agents propagate it and never silently replace it on cross-front calls.
11. **First adapter:** Whisper -> N03. It minimizes authority clash because N03 already owns audio and Whisper is a provider implementation, not a replacement for the canonical `audio.transcribe` capability.
12. **CPU/model weights:** CPU-only execution is allowed as DEGRADED when configured but below production performance expectations. Model weights live in a local volume/cache and never in Git.
13. **Heavy submodules:** fetch only at the nucleus that owns the adapter; N07 keeps the federation catalog. Do not recursively clone all 16 in all nuclei.
14. **Repo writes:** Resident Agents are read-only toward Git by default; they emit evidence and proposed patch artifacts. CI/human-controlled workflows create PRs. ASSUMPTION.
15. **HITL:** mandatory for delete/payment/deploy/credential rotation.
16. **Superpowers telemetry:** disable in SOUL environments when that option is supported by the pinned upstream integration. Do not rely on an undocumented env var without verification.
17. **Deploy order:** N01 -> N02 -> N03 -> N04 -> N05 -> N06 -> N07. N07 can exist separately, but its federation worker remains DEGRADED until peer prerequisites are verified.
18. **System ONLINE:** all seven health/discovery contracts verified + current crossfront evidence + one authenticated real E2E N01 <-> N07 <-> N03. This is the minimum gate, not proof that every optional OSS adapter is live.
19. **DEGRADED evidence:** use `soul-evidence/1` with explicit check states and `runtime.attempted=false/true`; dependency absence is DEGRADED, execution error is FAIL.
20. **Product persona:** single SOUL persona at the UX boundary. Nuclei remain internal; N05/N07 are the public operational boundaries.

## Required evidence JSON

```json
{
  "schema": "soul-evidence/1",
  "nucleus": "N03",
  "agent": "N03.resident",
  "capability": "audio.transcribe.whisper@1.0.0",
  "state": "DEGRADED",
  "correlationId": "real-request-id",
  "observedAt": "2026-10-02T00:00:00Z",
  "checks": [
    {"name":"adapter-config","state":"PASS"},
    {"name":"model-weights","state":"DEGRADED","code":"WHISPER_MODEL_NOT_AVAILABLE"}
  ],
  "runtime": {"attempted":false,"success":false}
}
```

The example is a schema example only; it is not runtime evidence.

## Mesh diagnosis N05/N07

The prior crossfront generator used `sources/{id}` for every nucleus. That is wrong for N07 because N07 is the repository root, not `sources/N07`. N05 had a second problem: its real Mesh implementation lives at `app/api/soul-mesh/route.ts` and the generated state was stale, so a previous scan reported "not detected" even though the code was present.

The corrective PR is **#104**. It resolves N07 at repository root and strengthens Mesh source markers while adding a read-only evidence assertion. The assertion must fail when the source is absent; it must never be weakened to force green.

## Phase rollout

### Phase 0 — discovery
Merge the N05/N07 discovery correction; regenerate crossfront evidence and require N05/N07 OBSERVED with real source paths.

### Phase 1 — first adapter
Whisper -> N03 through an explicit provider-specific capability `audio.transcribe.whisper@1.0.0`. The canonical public `audio.transcribe` capability remains N03-owned; the provider-specific adapter does not create a second authority.

### Phase 2 — resident agents
Install the seven embedded Resident Agents using `mesh.resident.describe@1.0.0`. Start with read-only evidence and native capability delegation. Do not add a second bus or HTTP framework.

### Phase 3 — additional OSS
Promote one OSS adapter at a time from structural-only -> adapter-planned -> adapter-live only after a real authenticated E2E. Each promotion gets its own contract tests, runtime evidence, and CI gate.

## Anti-patterns

- copying Superpowers skills as HTTP endpoints;
- treating a Git submodule as runtime activation;
- placing provider logic inside N07 instead of an explicit adapter;
- creating duplicate multi-agent authorities in N07;
- reporting model-missing as PASS;
- using mocks to make an E2E green;
- letting a Resident Agent write Git history directly without human/CI review.
