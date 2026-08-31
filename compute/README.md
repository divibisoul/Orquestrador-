# Compute + Neural Fabric implementation contract

This directory defines the provider-neutral execution substrate. Device telemetry is normalized into capabilities; the neural control layer scores routes; the prefrontal controller owns goals and constraints; the orchestrator owns execution.

## Cost function

`score = alpha*latency + beta*energy + gamma*(1-quality)`.

Weights are configuration, not magic constants. Predictions must be compared with measured results and persisted as experience before learning is enabled.

## Hardware truth

CPU fallback is always available. GPU/NPU adapters are optional and must report their actual capability. No benchmark target is considered achieved without measured evidence on declared hardware.
