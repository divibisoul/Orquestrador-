# Octacore — Certification R1

## G2 / G3 / G5 gates

| Slot | Runtime owner | Adapter | Contract test | CI execution | Current state |
|---|---|---|---|---|---|
| G2 | N02 / Eternium- | `octacore.execute` over canonical Soul Mesh | Real N02 handler: `mesh.ping`, correlation, deterministic 501 | Recent Actions run failed with `steps: null` | CODE/CONTRACT CERTIFIED; CI UNMEASURABLE |
| G3 | N03 / nexus-aeternum-fusion | `octacore.execute` over existing SoulMeshRouter | Real N03 handler: `mesh.ping`, correlation, deterministic 501 | Recent Actions runs failed with `steps: null` | CODE/CONTRACT CERTIFIED; CI UNMEASURABLE |
| G5 | N05 / nextjs-ai-chatbot | `octacore.execute` over existing N05 Mesh gateway | Real Next route: `mesh.ping`, correlation, deterministic 501 | No observable successful run for the Octacore branch | CODE/CONTRACT CERTIFIED; CI UNMEASURABLE |

No existing file was removed by these branches. Existing nucleus ownership and Mesh boundaries remain intact.

## Full federated path

The first-cycle implementation now exposes:

`G6 -> G4/G3 (parallel_group=pre) -> barrier=pre -> G0 SARA audit -> G0 SARA cycle`

N07/G7 owns scheduling, correlation, backpressure, circuit breaking and backend selection.

G0 remains serial and is the only regenerative authority.

G4 research failure is fail-closed unless `allowResearchSkip=true`. G3 is optional for text-only sessions and is activated when N06 supplies a perception payload.

The federated cycle accepts real context fields:

- `research_payload`
- `perception_payload`
- `pipeline_status`
- `correlation_id`

The consolidated context is forwarded through the existing SARA HTTP contract by `CycleWithContext`; no local regeneration is introduced.

## Mesh/Vagus/HortaCore fusion

- VagusBus remains the control plane.
- Existing Soul Mesh remains the data/execution plane.
- HortaCore is an additive N07/G7 composition layer.
- HortaCore keeps five processor identities; missing authoritative repositories are explicitly `PENDING_REPO`.
- ERU/audit identities delegate to SARA instead of creating duplicate authority.
- HortaCore Mesh discovery is performed by the existing N07 peer client.
- HortaCore state/result/error signals are published on VagusBus.
- Existing N07 SuperGPU is connected once to Octacore G7; Octacore does not create another SuperGPU runtime.

## Plus integration

Soul Admin Plus is no longer isolated from Octacore. Its existing N07 client, encrypted endpoint store, DashboardStateStore, WorkManager/Watchdog and cockpit were preserved and extended with an additive Octacore client.

The Android Plus layer now observes:

- N07 health
- Octacore/G7 health
- G0–G7 inventory
- authenticated Octacore transport
- explicit Octacore unavailable/unauthorized/invalid-response errors

No second Mesh is created on Android.

## Evidence rule

Repository source and contract tests can establish implementation/contract state. They cannot prove live distributed execution without a real configured N07/SARA endpoint and an observable build/runtime environment.

Current Actions infrastructure remains the hard validation blocker: several recent workflows complete as failures with no recorded steps/logs. Therefore this document intentionally does not label live E2E as PASS.

## Validation update 2026-09-23

N07 CI reached real runner execution after the normalization gate was corrected. The initial Octacore formatting failure was an automation defect (`git push HEAD:` with an empty ref), not a source compile failure. The normalizer now runs on push events and targets all Go files touched by the Octacore/HortaCore fusion.

The next verification gate must observe `gofmt`, `go vet`, unit tests, race tests, build and the existing N01->N07 federation E2E step before promoting N07 to runtime-certified.

## Validation update 2026-09-23 — current evidence

### Focused nucleus gates

- G2/N02 — PASS: Octacore G2 Focused Gate run 35816233796.
- G3/N03 — PASS: Octacore G3 Focused Gate run 35816195012.
- G5/N05 — PASS: Octacore G5 Focused Gate run 35816741470.
- G6/N06 — existing N06 Channel Contract and Soul Mesh CI are PASS on its Octacore branch.
- G4/N04 — N04 CI and Soul Mesh CI are PASS; the Octacore adapter is present and uses the canonical N04 runtime boundary.

### G7/N07 full Round-1 gate

The latest successful Octacore Focused Gate is run 35819219405. Its observed steps all passed:

`gofmt` → `go vet ./octacore ./hortacore` → `go test ./octacore ./hortacore` → race tests → federated preflow tests → full build.

The federated test passed the complete contract flow:

`G6 → G4/G3 parallel pre-stage → barrier → G0 SARA audit → G0 SARA cycle`.

The same gate also passed HortaCore tests, including shared Octacore control and canonical Mesh discovery.

### HortaCore fusion

HortaCore is now an additive G7 composition layer over the canonical Octacore processor. It does not create a second scheduler or Mesh. Mesh discovery is parallelized through the existing N07 peer client; HortaCore publishes capability/result/error/health events through the Vagus control boundary and exposes throttle/degrade/halt/resume signals against the same Octacore processor.

### Soul Admin Plus thaw

Plus is UNFROZEN for this integration stage. Its Android source now includes an authenticated HortaCoreClient over the existing N07 `/v1/execute` contract, persistent HortaCore health in DashboardStateStore, and cockpit visibility. This is implementation-level integration; APK/runtime commissioning remains unverified.

### Remaining gate

SARA G0 validation is still OPEN. The most recent SARA validation run exposed missing methods in OctaCoreG0Kernel that were themselves traced to an earlier accidental line-removal during fusion; those methods are being restored additively. No SARA production certification is declared until its own CI returns PASS.
