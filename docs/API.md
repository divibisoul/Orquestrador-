# N07 API — Production Contract

N07 exposes four execution domains with ten public functions in each domain. Public signatures remain stable while configuration, validation, observability and resource safety are enforced internally.

## Production gate

A production change is accepted only after the repository CI passes `go vet ./...`, `go test ./...`, `go test -race ./...` and `go build ./...`. A failing gate is a defect to be corrected before merge.

## Orquestrador
`New` builds the runtime and rejects missing dependencies. `Register` accepts `name` or `name@semver` and rejects invalid versions. `Route` resolves registered handlers and caches them with TTL. `Submit` validates messages, propagates context deadlines, enforces source rate limits, tracks active traces and updates the circuit breaker. `Execute` validates operation payloads and dispatches a traced message. `Cancel` cancels the active context associated with a trace. `Status` reports lifecycle state. `Health` returns component and circuit health. `Stats` returns aggregate counters and p95 latency. `Shutdown` stops new work, cancels active traces and shuts down compute.

## Neural
`New` validates learning rate and installs production defaults. `AddEdge` validates bounds, rejects self-cycles and detects graph cycles. `RemoveEdge` deletes only an existing edge. `Activate` supports one input vector or an exact flattened batch. `Forward` propagates through configured layers while honoring cancellation. `Learn` performs gradient clipping and configurable optimizer math. `Normalize` standardizes finite numeric vectors. `Attention` performs scaled attention with configured head adjustment. `Backprop` returns bounded finite gradients. `Health` reports density, learning and activation metrics.

## Prefrontal
`New` initializes a weighted decision policy and bounded history. `Evaluate` computes the best valid candidate against policy weights. `Plan` performs Pareto filtering before score ordering. `Prioritize` adds urgency and impact. `Inhibit` enforces safety constraints. `Select` chooses only valid candidates above threshold. `ValidateAction` is the admission gate. `Commit` writes a justified pending decision to bounded history. `Recall` retrieves recent decisions. `Health` reports evaluation/inhibition/latency metrics.

## SuperGPU
`New` creates the compute runtime around a concrete backend. `Discover` detects the host CPU and vendor tooling and only marks a GPU executable when the injected backend declares support. `Select` honors explicit device requests and never silently substitutes an unavailable requested device. `Reserve` creates a time-bounded exclusive reservation. `Release` verifies ownership. `Execute` runs through the compatible backend with context cancellation. `Batch` executes a cancellation-aware batch. `MemoryStats` reports known device metadata without fabricating hardware counters. `Health` reports backend/discovery state. `Shutdown` blocks new work and drains in-flight execution safely.

## Mesh
The `protocol.MeshEnvelope` follows the active SOUL Fusion Contract v1.2 fields: contract version, operation, payload, correlation identity, source, target and metadata. The `mesh.Adapter` provides real HTTP registration and dispatch hooks for N01/registry integration. Unknown or malformed contracts fail closed.

## Observability
`/health`, `/status` and `/metrics` are exposed by `cmd/nexus`. Metrics include request/success/error/cancel counts, in-flight work and p95 latency. Logs use structured JSON fields including component, operation and trace identity.
