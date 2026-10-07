# SOUL Engineering Execution Policy

This is the standing engineering policy for every future GitHub modification across the six base nuclei N01-N06 and the N07 orchestration/fusion layer.

## 1. Truth before narration

Never infer system state from prior claims, task text, percentages, filenames, or expected architecture. Before changing code, inspect the current branch, current HEAD, relevant files, recent commits from parallel fronts, and the latest applicable CI/runtime evidence.

## 2. Repair before reporting

When a defect, inactive area, inconsistency, incomplete capability, failing test, broken integration, stale contract, or false-positive validation is found, the default action is to repair it in the current state. Do not stop because a failure is severe. Diagnose the failure, research viable alternatives when necessary, implement the safest concrete correction, then validate again.

## 3. External research is an engineering tool

When the repository, compiler, runtime, platform, protocol, dependency, or CI environment blocks progress, use authoritative external documentation or current technical sources to identify alternative solutions. Do not treat an unfamiliar failure as a stopping condition.

## 4. Parallel-front reconciliation

Before creating a new implementation, inspect recent work from other fronts. Preserve correct concurrent changes, adapt to their current interfaces, and avoid duplicate implementations, duplicate owners, duplicated registries, or conflicting protocol definitions. A six-tab workflow is expected; GitHub history, commits, branches, and live files are the shared memory between fronts.

## 5. Write -> read back -> execute -> re-check

Every substantive code change must follow this evidence chain:

1. Write the correction to GitHub.
2. Read the changed file back from the resulting revision.
3. Run or trigger the narrowest meaningful test.
4. Run broader CI where required.
5. Inspect the resulting evidence against the exact commit that was changed.
6. If the state moved underneath the change, reconcile again instead of assuming the result still applies.

## 6. Runtime truth

A build proves compilation. A unit test proves the tested behavior. An integration test proves the exercised boundary. A real end-to-end transaction is required before marking a cross-nucleus communication path as operational.

## 7. Time and delivery dashboard

Every long-running audit must track elapsed audit time, current revision, current state, validation evidence, remaining blockers, and the next executable correction. Use a compact progress graph when reporting substantial work. Do not use percentage completion as evidence of correctness. Stalled stages must record why they are stalled and what concrete action resumes them.

## 8. Failure does not terminate the audit

A failing command is input to diagnosis, not the end of the task. Retry with the corrected hypothesis, change approach when appropriate, reduce scope to isolate the failure, research an alternative, or use another supported implementation path. Never fake success, suppress a meaningful error, or silently downgrade a failed validation.

## 9. Ownership and composition

Each executable capability has one authoritative owner. Other nuclei delegate through the canonical Mesh contract rather than copying specialized runtimes. New functions should emerge from safe composition of existing capabilities where useful, not from parallel duplicate implementations.

## 10. Six base IAs plus N07 orchestration/fusion layer

The SOUL base architecture remains six independent AI nuclei: N01, N02, N03, N04, N05, and N06. N07 is an additional technical nucleus/layer introduced after the original six-nucleus mission: it is itself executable, observable, testable, and capable of receiving and producing Mesh traffic, but it is not counted as a replacement seventh base AI in the six-way mission model.

N07 inherits all engineering rules that apply to N01-N06. N07 must therefore have explicit identity, agents, capabilities, tools, context, memory interfaces, execution, inputs, outputs, discovery, delegation, response, observability, security, resilience, tests, CI, and performance characteristics.

N07's role is orchestration, protocol mediation, distributed execution, SuperGPU coordination, federation, validation, and final composition. Its implementation may contain its own specialized compute/cognitive components, including the SuperGPU runtime, while preserving single ownership for each executable capability.

## 11. N07 parallel development and final activation gate

N07 is engineered continuously with all other fronts: protocol, security, observability, telemetry, orchestration, compute, tests, adapters, performance, discovery, delegation, and composition all remain active work. N07 must not be postponed until the end as a development task.

The final N07 activation/fusion gate remains closed until the six base nuclei have had their ingress/egress, capability, tool, authorization, correlation, error, retry, resource, shutdown, discovery, delegation, and response surfaces inventoried and validated. This does not prevent N07 development, isolated tests, adapters, or pairwise integration work.

## 12. Mesh topology

Each base nucleus N01-N06 must support five bidirectional peer relationships to the other five base nuclei. N07 participates in the broader Mesh as an additional orchestration endpoint and therefore must support validated traffic with every base nucleus. The six-by-five base topology and the N07 orchestration links must not be conflated.

The canonical flow is:

SOUL MESSAGE -> TRANSPORT RESOLUTION -> DISCOVERY -> CAPABILITY ROUTING -> DELEGATION -> EXECUTION -> RESPONSE -> CORRELATION -> COMPOSITION -> VALIDATION

Use existing Mesh mechanisms first. Add adapters or transport implementations only when they close a verified interoperability gap.

## 13. Hybrid transport

The architecture may use HTTP, REST, WebSocket, realtime, loopback, internal calls, events, Pub/Sub, or other justified transports. Agents interact with a canonical SOUL message abstraction; transport selection remains an infrastructure concern.

## 14. SuperGPU

SuperGPU means distributed parallel computation, not a claim that every host has a physical GPU. The architecture should support decomposition, scheduling, capability routing, parallel execution, aggregation, validation, fallback, resource accounting, and observability across N01-N06 and N07, including intra-nucleus worker pools where they are real.

N04 remains an important execution/parallelism nucleus, but SuperGPU ownership belongs to the distributed architecture and N07 orchestration layer rather than to N04 alone.

## 15. Synergy and emergent capabilities

For every relevant pairing and composition, inspect agents x tools x capabilities x context x execution. Ask what becomes possible jointly that is not possible independently. Only create an emergent capability when a technically coherent contract, useful behavior, ownership model, implementation path, and validation strategy exist.

Track pairwise, four-nucleus, and six-nucleus discoveries in a live architectural matrix. Composition must preserve traceability to the component capabilities and participating agents/tools.

## 16. Delegation and failure recovery

When a nucleus cannot satisfy a task locally, it should first discover peer capabilities and delegate when an authorized, suitable route exists. A missing local implementation is not automatically a system-level NOT_IMPLEMENTED result.

Connections should use appropriate timeout, retry, backoff, circuit-breaking, correlation, tracing, validation, rate limiting, size controls, authentication, worker recovery, and fallback mechanisms.

## 17. Architectural memory between fronts

Every substantial front must leave durable evidence in GitHub of what changed, what was found, what remains, exact commit/branch information, dependencies, blockers, and the next recommended action. This record is part of the system architecture and is not replaced by conversational memory.

## 18. Closure criteria

A nucleus is closed only when its structure, capabilities, agents, tools, Mesh integration, ingress, egress, discovery, delegation preparation, security, resilience, CI, documentation, and synergy surfaces are validated. When end-to-end validation requires the whole network, mark the item STRUCTURALLY VALIDATED — INTEGRATED TEST PENDING instead of claiming operational status prematurely.

The system may be declared online only when the relevant code exists on the claimed revision and the required validation evidence is green for that same revision.

## 19. Required final questions

For every nucleus and composition stage, ask:

1. What can this nucleus do alone?
2. What can it do with one peer?
3. What can four nuclei do together?
4. What can the six base nuclei do together that none can do alone?
5. What new, technically grounded capability emerges from that composition?
