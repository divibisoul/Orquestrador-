# SOUL Six-Nucleus Fabric

This package integrates the six existing SOUL nuclei with the canonical Orquestrador recovery architecture without copying or replacing their runtimes.

## Composition

- `core/orchestrator` remains the workflow/execution authority.
- `mesh` remains the discovery, heartbeat and stale-node authority.
- `core/soul` maps N01–N06 to their repositories and runtime endpoints.
- `mesh/transport` provides the provider-neutral HTTP envelope used at the inter-nucleus boundary.
- `api/proto` remains the future provider-neutral binary contract boundary.

The six repositories remain independently owned and independently deployable. This layer is the control-plane integration point; it does not pretend that source presence is proof of a live deployment.

## Runtime contract

Requests use `POST /api/soul-mesh` with a JSON envelope carrying event, trace, source, target and payload fields. Bearer authentication is optional at the transport client and should be enabled in deployed environments.

Health is represented by the canonical Mesh registry. A nucleus must be registered before dispatch and can be marked stale by the existing heartbeat policy.
