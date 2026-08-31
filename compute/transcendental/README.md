# Transcendental Compute Engine

`compute/transcendental` is a pure-software, deterministic simulation layer for five reference accelerator classes: NVIDIA Blackwell B200, NVIDIA Vera Rubin, AMD Instinct MI400/CDNA5, Google TPU v6e Trillium, and Huawei Atlas 950.

It provides three interface-level roles:

- `ComputeBackend` for simulated execution and workload estimates.
- `CostEstimator` for Prefrontal planning.
- `MetricsProvider` for Neural Fabric observability.

## Isolation

The package depends only on the Go standard library and its own packages below `compute/transcendental`. It does not import `runtime.go`, `router.go`, `cortex.go`, application packages, providers, databases, APIs, or hardware drivers.

## Deterministic model

Performance is estimated with a simple roofline model:

`computeTime = flops / (pFLOPS * 1e15)`

`memoryTime = dataBytes / (bandwidthGBs * 1e9)`

`estimated = max(computeTime, memoryTime) / efficiencyFactor`

The default efficiency factor is `0.7`. No random source and no real `time.Sleep` are used for simulated execution. Metrics contain an observational timestamp, but performance calculations are deterministic for identical inputs and configuration.

## Reference data caveat

Vera Rubin values used here are preliminary 2026 marketing/reference figures. They must be updated when official specifications are published.

Atlas 950 is modeled as a scale-out reference, not as a single GPU. Its figures are preliminary/reference estimates and must be replaced by official published specifications when available.

## Auto selection

The selector filters insufficient memory first, applies precision support/fallback penalties, restricts FP64 to the supplied Blackwell/Rubin reference family, prefers low latency for small workloads, and favors high-bandwidth/high-memory models for large workloads. Explicit modes are also supported.

## Activation

The feature is disabled by default. A future Orquestrador integration should inject interfaces only, keep pointers nil-safe, and preserve the legacy path when `compute.enabled` is false. No existing integration files are modified by this package until those files exist and are inspected.

## Extension

Add a model by implementing `models.PerformanceModel`, adding it to `models.DefaultCatalog`, and adding deterministic model/selector tests. Do not import the rest of the repository into this package.
