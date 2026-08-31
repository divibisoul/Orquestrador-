# API Reference

## Internal contracts

Protobuf definitions live in `api/proto/` for Orchestrator, Prefrontal and Super AGI.

## TaskSpec

Carries objective, subtasks, priority, precision requirement, cost budget, speculation flag, confidence threshold and compatibility metadata.

## NodeAnnounce

Advertises node status, capabilities, transports, compute resources and compatible systems.

## EventEnvelope

Carries event identity, trace/correlation IDs, source/target, payload, timestamp and causal metadata.

REST/WebSocket adapters and gRPC generated bindings are intentionally provider-neutral and can be added without changing core decision logic.
