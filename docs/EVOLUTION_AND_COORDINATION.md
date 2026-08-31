# Evolution and Cross-Front Coordination

This document records objective repository state so parallel work can be reconciled from GitHub rather than inferred from chat history.

## Operating rule

Before every mutation: inspect the current target branch, inspect active parallel branches/PRs, compare overlapping paths, preserve compatible work, merge only when the architecture and contracts are compatible, and correct incomplete or broken areas immediately when safe.

## Current topology

```text
                         SOUL ORCHESTRATOR
                                |
                 +--------------+--------------+
                 |                             |
          Control / Cognitive              Compute Plane
                 |                             |
        +--------+--------+              +-----+------+
        |        |        |              |            |
     Planner   Registry  Health         TCE         Fabric
        |        |        |              |            |
        +--------+--------+              +-----+------+
                 |                             |
                 +-------------+---------------+
                               |
                          SOUL MESH
                               |
                 +------+------+------+------+------+
                 |      |      |      |      |      |
                N01    N02    N03    N04    N05    N06
```

## Objective evolution

```text
Initial isolated TCE
       |
       v
Contracts + models + selector + executor
       |
       v
Validation + nil safety + error propagation
       |
       v
Parallel-front coordination
       |
       v
Trinity / Fabric / Mesh architecture discovered on parallel branches
       |
       v
NEXT: reconcile overlapping compute abstractions without destructive merge
```

## Objective progress ledger

| Prompt stage | Repository action | Evidence | Status |
|---|---|---|---|
| Initial TCE | isolated contracts/models/executor | Git history | complete |
| Refinement 1 | precision and plan validation | commit `105bd4b97a79d898c140c4454ac9f5b084d09a57` | complete |
| Refinement 2 | test coverage for validation | commit `974fd19ccabda1944369e0b695439f96c4221cfc` | complete |
| Refinement 3 | FP64 selector constraint restored | commit `fe180e90f4374004a52883c02fc9245ee5a7eeb3` | complete |
| Current stage | config safety + parallel cap tests | commits on `main` after `974fd19...` | applied |
| Cross-front audit | compare `main` with `integration/soul-six-nuclei` | 164 ahead / 58 behind, divergent | requires reconciliation |

## Parallel branches observed

The repository currently contains independent branches including `feature/nexus-finalization-v1`, `feature/trinity-fundamental-v1`, `feature/trinity-production-hardening-v1`, `feature/nexus-runtime-hardening-batch1`, `feature/readiness-dashboard`, `feature/gemini-provider`, `foundation/orchestrator-nexus-v1`, `integration/soul-six-nuclei`, and `recovery/consolidation`.

These branches must not be treated as if their contents are already in `main`.

## Reconciliation rule

The `integration/soul-six-nuclei` branch is structurally much more advanced than `main` but has diverged. Its Compute/Fabric/Trinity work is therefore an input to architectural reconciliation, not permission to duplicate or blindly merge code into the isolated TCE.

## Work-quality metric

No subjective percentage is assigned to the assistant's "performance". Progress is measured objectively by repository commits, changed paths, CI results, defects discovered, defects fixed, and successful post-write reads. A failed CI is progress only when its defect is subsequently corrected and revalidated.
