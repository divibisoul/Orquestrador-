# N07 fusion state

Current target: N01 + N06 capability/tool federation into N07 without duplicating native ownership.

N01 owner: native Android capabilities.
N06 owner: contextual support capabilities and registered tools.
N07 owner: discovery, route selection, delegation, composition, correlation and aggregation.

The live GitHub state currently contains multiple concurrent N07 federation PRs. This branch intentionally remains isolated until its exact-head CI proves correctness and a maintainer chooses the strongest non-duplicating implementation.

Concurrent PR observed: #15 `upgrade/n07-hybrid-supercompute-v2` adds broad federation/supercompute. Its `orchestrator/federation.go` imports `mesh`, while the current `mesh` package imports `orchestrator` through the gateway, creating an import-cycle hazard. This branch avoids that architecture by keeping the orchestrator dependent only on the `PeerInvoker` interface.

Current branch implements:
- verified N01/N06 ownership catalog;
- Mesh-based discovery before delegation;
- correlation-preserving N01/N06 invocation;
- peer retry/backoff/circuit handling inherited from the live Mesh client;
- sequential capability composition with required/optional steps;
- N07 control-plane advertisement of ownership and transports.

No claim of live six-runtime E2E is made until exact-head CI and runtime federation evidence are available.
