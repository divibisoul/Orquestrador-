# N07 concurrent-front coordination

The N07 repository currently has multiple open federation fronts. This front must not overwrite concurrent changes. Before touching a shared path, re-read its current GitHub SHA. On SHA conflict, stop the write, fetch the latest file, rebase the intended change conceptually, and write against the latest state.

## Current concurrent work observed

- PR #13 `upgrade/soul-federation-live-2026-09-01`: hybrid federation/discovery/capability execution.
- PR #14 `fusion/n01-n06-n07-live`: N01/N06 ownership-aware fusion boundary.
- PR #15 `upgrade/n07-hybrid-supercompute-v2`: hybrid federation, topology and distributed SuperCompute.
- PR #10 / #9 / #6 / #5 / #4 / #3 are additional historical/concurrent Orquestrador work on other base branches.

## Ownership rule for this front

N01 remains owner of native Android execution. N06 remains owner of N06 tools/capabilities. N07 owns discovery, routing, federation, composition, correlation, policy and aggregation. Do not duplicate executor implementations in N07 when delegation through the canonical Mesh is sufficient.

## Current risk to coordinate

The current PR #15 federation implementation imports the `mesh` package from `orchestrator`, while the existing HTTP gateway imports `orchestrator`. This creates a potential Go package import cycle if merged unchanged. The safe boundary is to keep orchestration transport-neutral and inject a peer/federation interface from the mesh layer.

## Next

Reconcile the strongest federation primitives from the concurrent fronts through a single transport-neutral interface, then run exact-head CI before merging anything into main.
