# N07 — SOUL Orquestrador

N07 is the independent SOUL orchestration runtime. It coordinates N01–N06 through the canonical Soul Mesh and composes neural, prefrontal, compute, federation and SuperCompute capabilities without copying peer runtimes.

```text
N01 ─┐
N02 ─┤
N03 ─┤
N04 ─┼──► SOUL MESH ──► N07 ORQUESTRATOR ──► Neural ──► Prefrontal ──► Compute
N05 ─┤                         │
N06 ─┘                         ├──► Discovery / Routing / Delegation
                              ├──► Capability Composition / Fusion
                              └──► Federated SuperCompute / Parallel Execution
```

## Runtime contract

- Language: Go 1.25.
- Native N07 protocol: `N07.v1`.
- Canonical Soul Mesh contract: `soul-mesh/1`, contract `1.1.0` where supported by the peer.
- Messages carry source, target, message identity, correlation identity, timestamp and operation metadata.
- Authentication uses the repository's configured Mesh HMAC boundary; replay, timestamp and payload limits are enforced at the relevant ingress.
- Local execution is preferred. Unsupported capabilities may be discovered and delegated to an available peer.
- A logical SuperGPU is distributed parallel execution across nuclei; no physical GPU capability is claimed without a concrete accelerator backend.
- Shutdown and execution paths are context-aware.

## N07 capability planes

### Orchestrator
`New`, `Register`, `Route`, `Submit`, `Execute`, `Cancel`, `Status`, `Health`, `Stats`, `Shutdown`.

### Neural
`New`, `AddEdge`, `RemoveEdge`, `Activate`, `Forward`, `Learn`, `Normalize`, `Attention`, `Backprop`, `Health`.

The neural runtime validates finite inputs, prevents graph cycles, supports cache invalidation after topology/learning changes, tracks learning/activation metrics and exposes health state.

### Prefrontal
`New`, `Evaluate`, `Plan`, `Prioritize`, `Inhibit`, `Select`, `ValidateAction`, `Commit`, `Recall`, `Health`.

### SuperGPU
`New`, `Discover`, `Select`, `Reserve`, `Release`, `Execute`, `Batch`, `MemoryStats`, `Health`, `Shutdown`.

### Federation
N07 also provides the peer-facing federation boundary for N01–N06: peer registration, discovery, capability resolution, health-aware routing, delegation, retry/backoff, circuit protection, correlation and federation metrics.

## Mesh topology

N07 is structurally connected to all six base nuclei:

```text
           N01
          ↗   ↘
      N02       N06
       ↑   N07  ↑
      N03       N05
          ↖   ↙
           N04
```

The topology is logical and transport-neutral. Concrete transports are selected only when implemented/configured. Supported architectural modes include in-process, loopback HTTP, HTTP, realtime and event adapters.

## Neural federation

The neural fabric exposes N01–N06 as remote members through one transport-neutral `RemotePeer` boundary. It supports:

- explicit target routing;
- correlation propagation;
- deadline-aware execution;
- finite-number payload validation;
- broadcast;
- parallel task execution;
- deterministic member ordering;
- compatibility with the historical `Source` routing hint.

This is a federation boundary, not a second neural implementation inside each peer.

## Capability discovery and routing

The routing lifecycle is:

`REQUEST → DISCOVERY → CAPABILITY MATCH → PEER SCORING → ROUTE → EXECUTE/DELEGATE → RESULT → CORRELATION`

Peer selection can use capability ownership, health, latency, failure history, current load and transport availability. Registered operation versions are canonicalized as `name@semver` so explicit versions route exactly and unversioned requests can resolve to the newest compatible registration.

Discovery is cached for a bounded TTL and invalidated after relevant peer failure so the cache cannot silently mask an unavailable runtime.

## Capability composition

N07 composes existing capabilities rather than inventing duplicate implementations. A composed operation preserves component provenance and may run sequentially or in parallel depending on dependency constraints.

The architecture supports the cumulative composition model:

`Agent × Tool × Capability × Context × Execution → Emergent Capability`

Only combinations with an executable, technically grounded path are eligible for registration.

## Federated SuperCompute

The distributed execution pipeline is:

`TASK → DECOMPOSITION → SCHEDULER → CAPABILITY ROUTER → PARALLEL EXECUTION → AGGREGATION → VALIDATION → FINAL RESULT`

Independent tasks receive child correlation identities while preserving the parent request identity. Per-task duration and partial failures are retained so the orchestrator can measure real parallel behavior instead of only returning a boolean success.

## Control plane

The runnable N07 service exposes the canonical Mesh ingress plus operational inspection endpoints. The exact endpoint set is implementation-controlled and must be verified from the current runtime before deployment; documentation must not be treated as proof of liveness.

## Observability

The runtime tracks request/failure/success state, peer health, latency and federation activity. Release validation must use observed measurements for latency, concurrency, failure recovery and cancellation rather than relying on documentation claims.

## Security boundary

Secrets are not embedded in source. Mesh authentication, nonce/replay protection, timestamp/TTL controls, payload-size limits and correlation validation are treated as protocol responsibilities. Provider/API credentials remain deployment configuration and are never considered proof of execution until the provider path returns a real result.

## Validation policy

A capability is considered complete only when its executable path exists, its input/output contract is validated, its failure behavior is explicit and tests exercise the relevant boundary. GitHub Actions on the exact revision are the authoritative CI gate.

Final release requires:

1. exact-head format, vet, test, race and build evidence;
2. compatible N01–N06 runtime configuration;
3. live bidirectional Mesh commissioning;
4. neural federation E2E;
5. capability composition/fusion E2E;
6. federated SuperCompute concurrency/failure measurements;
7. final operational smoke test.

Until those gates pass, N07 is `STRUCTURALLY READY — RUNTIME COMMISSIONING PENDING`, not falsely declared production-complete.
