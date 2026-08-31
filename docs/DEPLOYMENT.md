# Deployment

## Single node

```bash
go test ./...
go run ./cmd/nexus
```

## Multi-node reference

Run the control plane with 3 or 5 state nodes, expose gRPC internally with mTLS, and attach compute workers through capability announcements. Kubernetes, edge and mobile adapters belong in deployment-specific layers rather than the cognitive core.

## Production prerequisites

Calibrate latency, throughput, power and recovery benchmarks on the actual hardware/model stack. Do not infer SLO compliance from local development runs.
