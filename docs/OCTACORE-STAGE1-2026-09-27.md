# Octacore — Stage 1 (2026-09-27)

## Frozen boundary

The previously implemented Soul Admin Plus remains a separate preserved front. This Stage 1 adds Octacore on top of the current N07/SARA code; it does not replace or rewrite the frozen Plus branch.

## Definition

Octacore is the SOUL/SARA system GPU: a federated software execution processor over eight domain slots G0-G7.

It is not a silicon octa-core CPU and makes no claim of physical CUDA/GPU/NPU hardware.

## Stage 1 inventory

| Slot | Nucleus | Current evidence | Stage 1 state |
|---|---|---|---|
| G0 | SARA | Existing SARA HTTP + RegenerativeLoop + ARA/ETR/ITR + VagusNerveBus | IMPLEMENTED kernel boundary |
| G1 | N01 | Current Mesh host and native SOUL runtime | REPO_PRESENT_RUNTIME_UNVERIFIED |
| G2 | N02 | Current Soul Mesh endpoint and N02 capability runtime | REPO_PRESENT_RUNTIME_UNVERIFIED |
| G3 | N03 | Current Soul Mesh endpoint and audio/multimodal runtime | REPO_PRESENT_RUNTIME_UNVERIFIED |
| G4 | N04 | Current N04 Mesh/runtime/tool boundary | ADAPTER_READY |
| G5 | N05 | Current N05 Mesh gateway/inference runtime | REPO_PRESENT_RUNTIME_UNVERIFIED |
| G6 | N06 | Current N06 dispatcher/session/tool runtime | ADAPTER_READY |
| G7 | N07 | Existing SuperGPU Runtime + Mesh peer client + dynamic router + federated gateway | PROCESSOR/SCHEDULER |

Repository presence is not runtime proof.

## Stage 1 implementation

### N07

- octacore/types.go: canonical Job, Result, Slot and VagusEnvelope contracts.
- octacore/processor.go: processor boundary using the existing N07 SuperGPU Runtime and Mesh PeerClient.
- octacore/scheduler.go: real concurrent waves, barriers, global inflight bound, token bucket, per-slot circuit breaker, TTL, backend preference handling and deterministic unavailable states.
- octacore/operations.go: registration on the existing N07 Engine; no second public ingress.
- backend/sara_proxy.go: Vagus publication through the existing SARA HTTP boundary.
- backend/backend.go: existing /v1/execute now preserves an explicitly supplied correlationId.
- Existing SuperGPU registration remains active and was corrected to be engine-local instead of process-global sync.Once.

### G0 / SARA

- SistemaVivo.process and RegenerativeLoop.run accept optional federated context.
- The context is preserved inside canonical CycleContext.artifacts; it does not bypass ARA/ETR/ITR.
- OctacoreG0Kernel adds serial admission, bounded queue, backpressure, throttle and health over the canonical SARA runtime.
- /v1/cycle routes through G0 and exposes a context-presence summary.
- /v1/vagus publishes the defined GPU/control envelopes to the existing VagusNerveBus.
- Capability discovery advertises octacore.g0@1.0.0.

### Host integrity

- N01 soul-supergpu.mjs no longer reports a successful local execution when there is no real local executor.
- N01 telemetry no longer reports a hard-coded AGI 17/17, 60 FPS or false memory values.
- Arquitetura Quadrangular is no longer presented as runtime-active without an executable health source.

## Scheduler rules

G0 is single-concurrency because SARA remains the sole regenerative authority.

For all other slots, backend_prefs are honored in order. REMOTE_MESH reuses the existing Soul Mesh and PeerClient HMAC/correlation path.

WEBASSEMBLY and WEBGPU return explicit unavailable errors until a real backend is detected. The implementation never substitutes a fake CPU result for an unavailable WebGPU backend.

parallel_group creates a concurrently executable wave. A job with a barrier and a different consumer group waits until the producer group using that barrier leaves the pending set.

## Engineering rationale

The design follows established distributed-systems practices:

- W3C Trace Context defines standardized propagation of distributed request context.
- Go Context is propagated across service boundaries to carry cancellation/deadlines.
- AWS guidance recommends bounded retries/backoff, token buckets, jitter and idempotency for side-effecting retries.
- WebGPU must be detected from the actual runtime; requestAdapter() may return null, and WebGPU is only exposed in supported secure contexts.

## Validation state

Stage 1 source changes are additive. The current repositories must still pass their own CI after this overlay is applied. No source-level implementation is labeled production-operational until build, automated tests, live Mesh transport and SARA commissioning produce evidence.

## Stage 2 boundary

Stage 2 will deepen the individual kernel adapters for G1-G6, wire N04/N06 high-throughput job generation, add the N03/N05 mappings without replacing their native runtimes, connect session pre-context/barrier flow end-to-end, and execute cross-repository integration tests.

No Stage 2 semantics are being invented in Stage 1.
