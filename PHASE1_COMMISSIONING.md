# Phase 1 — Real Mesh Commissioning

## Scope

This procedure commissions the current N07 real federation gate against live N04, N05 and N06 runtimes.

It does **not** declare SOUL fully ONLINE. N01 live execution, N02/N03 external execution and SARA handoff remain separate evidence gates.

## Automated GitHub Actions commissioning

The canonical CI path now self-provisions the Phase-1 Mesh transport for the real federation gate:

1. N07 checks out the current N04, N05 and N06 repositories from their `main` branches.
2. Each real peer application is installed and production-built on the GitHub-hosted runner.
3. N04, N05 and N06 start as real Next.js runtimes on isolated local ports.
4. A fresh `SOUL_MESH_HMAC_SECRET` is generated for the run.
5. Three Cloudflare Quick Tunnels create temporary public HTTPS origins for those local runtimes.
6. The generated origins are injected into N07 as `SOUL_MESH_N04_URL`, `SOUL_MESH_N05_URL` and `SOUL_MESH_N06_URL`.
7. The existing integration-only N07 test executes the real Mesh requests and validates discovery, native capability execution, correlation, source/target and response HMAC.
8. Runtime logs and the participating source HEADs are retained as workflow artifacts.

No Mesh URL, tunnel token, paid domain or permanent Mesh secret is required for this CI path.

Cloudflare Quick Tunnels are intended for temporary development/testing and generate random `trycloudflare.com` origins. They are not the production ingress mechanism.

## Manual external commissioning

For an externally deployed environment, the original explicit variables remain supported:

- `SOUL_MESH_N04_URL`
- `SOUL_MESH_N05_URL`
- `SOUL_MESH_N06_URL`
- `SOUL_MESH_HMAC_SECRET`

The URLs must be the base deployment origins. The test appends `/api/soul-mesh`.

Do not store URLs or secrets in source files.

## Runtime contract

The live gate expects:

- protocol: `soul-mesh/1`
- contractVersion: `1.1.0`
- N04 native probe: `core.health`
- N05 native probe: `core.health`
- N06 native probe: `support.mesh`

Before execution, N07 sends `mesh.describe` to the target peer. A declared capability is not accepted as proof by itself; the subsequent native transaction must also return a correlated response.

## Exact integration command

From the N07 repository checkout:

```bash
go test -tags=integration ./mesh \
  -run '^TestN07FederatesToRealN04N05N06NativeCapabilities$' \
  -count=1 -v
```

Expected evidence is a separate subtest result for N04, N05 and N06 with preserved `correlationId`, correct `source`, correct `target=N07`, and a valid response payload.

## Fail-closed behavior

Missing or unusable runtime configuration must fail the gate. An absent peer is never converted into PASS.

In automated CI, a failed Quick Tunnel, unreachable peer, authentication failure or capability failure is a real failure of the commissioning job; it is never replaced with a synthetic peer.

## Evidence to preserve

Preserve the complete Actions log for:

1. dependency/toolchain setup;
2. peer source HEAD capture;
3. real peer build/start;
4. Mesh tunnel creation and reachability;
5. Mesh discovery;
6. each native capability execution;
7. correlation/source/target assertions;
8. response HMAC verification;
9. final test summary.

The log must not contain secret values.

## Current evidence boundary

| Gate | Current state |
|---|---|
| N04 real native E2E | NOT MEASURED until the new self-provisioned real federation job completes successfully |
| N05 real native E2E | NOT MEASURED until the new self-provisioned real federation job completes successfully |
| N06 real native E2E | NOT MEASURED until the new self-provisioned real federation job completes successfully |
| N01 health/discovery | Source endpoint exists; live cross-runtime result is NOT MEASURED |
| N02 | Mesh health endpoint exists; live cross-runtime result is NOT MEASURED |
| N03 | Mesh API and native capabilities exist; live cross-runtime result is NOT MEASURED |
| SARA | Local CI/self-check evidence exists; live N07↔SARA handoff is NOT OPEN |

## Release rule

Do not change `PHASE 1 PARTIAL` to full ONLINE until the corresponding live transactions have produced reproducible evidence.

Do not open the Goal→Plan→Act→Observe cognitive loop until the Phase 1 release gates are satisfied.

## Preservation and uniform evolution invariant

Commissioning the Mesh must not reduce the internal capability of any participating nucleus.

Before a phase is promoted, record the participating nucleus' preserved native systems, adapted bridges, extended capabilities, measured communication paths and explicit blockers. The Mesh is the canonical communication substrate; it must not be duplicated by a second hidden bus.

This includes Projeto Clareira and its native neural graph, ProcessingNodes, InformationChannels, HomeostasisManager and EventBus/bridge in N01, as well as native agents, tools, capabilities, registries and runtime services in N02–N07. SARA remains independent and authoritative for its regenerative/governance responsibilities.

The broader architecture's 72 nódulos are protected scope. Their individual statuses must be source-enumerated before they are marked measured; no count is converted into fabricated execution evidence.

See SOUL_EVOLUTION_PRESERVATION.md and COMMUNICATION_MATRIX.md.

### Current CI evidence

A completed PASS exists for commit b139ac488137544e5a814be73305313603bf8ada (Actions run #875). That run is historical evidence for that exact commit; it is not a PASS claim for the current PR head.

The current PR head is still not marked VALIDATED until the self-provisioned real Mesh federation job has executed successfully.
