# N01 + N06 + N07 — Fusion Contract

## Ownership

N01 remains the executor of native Android capabilities. N06 remains the executor and owner of its capabilities, tools, agents and user-context runtime. N07 owns orchestration, discovery-driven routing, correlation, composition and aggregation.

## Verified N01 capabilities

`android.device_info`, `android.battery`, `android.memory`, `android.network`, `android.events`, `shizuku.bridge`, `brightness`, `wifi`, `bluetooth`, `airplane`, `background-process`.

## Verified N06 capabilities and tools

Capabilities: `support.context`, `support.artifacts`, `support.documents`, `support.tool-execution`, `support.streaming`, `support.mesh`, `support.ai-pilot`.

Tools: `createDocument`, `updateDocument`, `getWeather`, `requestSuggestions`.

## Runtime rule

N07 never copies the implementation of an owned N01/N06 capability. It discovers the executable owner through the canonical Mesh endpoint and delegates using the same `correlationId`.

## Composition

`mesh.fusion.execute` accepts an ordered `steps` list. Each step contains a real Mesh capability, payload, and optional `required` flag. N07 resolves the owner, delegates execution, records duration/owner/result, and stops on a failed required step. Optional failures remain visible in the aggregate response.

## Control plane

`mesh.describe`, `mesh.discovery`, `mesh.capabilities` and `mesh.fusion.describe` expose the N07 control plane and verified ownership catalog. Live capability execution still requires peer discovery to confirm that the owner is actually executable.

## Security and resilience

All delegated calls retain Mesh contract `1.1.0`, HMAC, nonce replay protection, correlation, payload limits, context cancellation, retry/backoff and peer circuit isolation.

## Runtime readiness

The implementation is structurally integrated when the branch compiles and tests pass. Full live E2E remains dependent on N01/N06 instances being deployed with compatible `/api/soul-mesh` endpoints and the same credential configuration. This is a validation state, not an implementation gap.
