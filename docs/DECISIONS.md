# Architecture decisions

## Go

Go is the foundation language because the first implementation emphasizes concurrency, simple deployment, explicit interfaces and low operational overhead. Model-specific numerical kernels remain behind adapters and may use specialized runtimes.

## No hard vendor dependency

The cognitive core depends on contracts, not Kubernetes, Redis, NATS, a particular LLM or GPU vendor. Infrastructure integrations are replaceable adapters.

## Honest capability boundaries

A stub/provider boundary is never represented as a completed production capability. This prevents architecture diagrams from becoming false implementation claims.
