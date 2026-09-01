# SOUL cumulative execution queue

This queue is the cross-front coordination artifact for the cumulative SOUL directives. It is evidence-oriented: a task is not marked complete from documentation alone; executable code and validation evidence are required.

## Current priority
1. Re-read current N01–N07 heads before every mutation.
2. Preserve peer ownership; N07 owns orchestration/neural federation and peers expose adapters.
3. Reconcile protocol, correlation, authentication, timeout/cancellation, capability ownership and tool/agent contracts before fusion.
4. Keep N07 last-stage fusion compatible with the six base nuclei and do not create duplicate runtimes.

## 40-function execution order

| # | Plane | Function | Closure requirement | Status |
|---:|---|---|---|---|
| 1 | Orchestrator | New | executable construction + dependency validation | VERIFIED STRUCTURAL |
| 2 | Orchestrator | Register | semantic versioning + duplicate isolation | VERIFIED STRUCTURAL |
| 3 | Orchestrator | Route | exact/newest-compatible routing + cache | VERIFIED STRUCTURAL |
| 4 | Orchestrator | Submit | deadline/rate/breaker/trace/cleanup | VERIFIED STRUCTURAL |
| 5 | Orchestrator | Execute | message + schema gate | VERIFIED STRUCTURAL |
| 6 | Orchestrator | Cancel | trace cancellation | VERIFIED STRUCTURAL |
| 7 | Orchestrator | Status | runtime state | VERIFIED STRUCTURAL |
| 8 | Orchestrator | Health | component/circuit health | VERIFIED STRUCTURAL |
| 9 | Orchestrator | Stats | telemetry aggregation | VERIFIED STRUCTURAL |
| 10 | Orchestrator | Shutdown | cancellation + compute shutdown | VERIFIED STRUCTURAL |
| 11 | Neural | New | bounded learning/runtime init | VERIFIED STRUCTURAL |
| 12 | Neural | AddEdge | graph/cycle/bounds validation | VERIFIED STRUCTURAL |
| 13 | Neural | RemoveEdge | deletion + cache invalidation | VERIFIED STRUCTURAL |
| 14 | Neural | Activate | finite input + batch semantics | VERIFIED STRUCTURAL |
| 15 | Neural | Forward | context + bounded cache | VERIFIED STRUCTURAL |
| 16 | Neural | Learn | SGD/RMSProp/Adam + regularization/clipping | VERIFIED STRUCTURAL |
| 17 | Neural | Normalize | stable finite normalization | VERIFIED STRUCTURAL |
| 18 | Neural | Attention | scaled attention + finite validation | VERIFIED STRUCTURAL |
| 19 | Neural | Backprop | derivatives + clipping | VERIFIED STRUCTURAL |
| 20 | Neural | Health | topology/learning/cache/gradient metrics | VERIFIED STRUCTURAL |
| 21 | Prefrontal | New | threshold/capacity/policy init | VERIFIED STRUCTURAL |
| 22 | Prefrontal | Evaluate | multicriteria scoring | VERIFIED STRUCTURAL |
| 23 | Prefrontal | Plan | Pareto filtering + ordering | VERIFIED STRUCTURAL |
| 24 | Prefrontal | Prioritize | score/urgency ordering | VERIFIED STRUCTURAL |
| 25 | Prefrontal | Inhibit | risk/utility gate | VERIFIED STRUCTURAL |
| 26 | Prefrontal | Select | policy-aware selection | VERIFIED STRUCTURAL |
| 27 | Prefrontal | ValidateAction | pre-commit policy validation | VERIFIED STRUCTURAL |
| 28 | Prefrontal | Commit | decision record + justification | VERIFIED STRUCTURAL |
| 29 | Prefrontal | Recall | bounded history retrieval | VERIFIED STRUCTURAL |
| 30 | Prefrontal | Health | evaluation/inhibition/latency metrics | VERIFIED STRUCTURAL |
| 31 | SuperGPU | New | backend/resource initialization | VERIFIED STRUCTURAL |
| 32 | SuperGPU | Discover | cached backend discovery | VERIFIED STRUCTURAL |
| 33 | SuperGPU | Select | preference-aware device selection | VERIFIED STRUCTURAL |
| 34 | SuperGPU | Reserve | exclusive reservation + expiry | VERIFIED STRUCTURAL |
| 35 | SuperGPU | Release | owner-checked release | VERIFIED STRUCTURAL |
| 36 | SuperGPU | Execute | context + capability enforcement | VERIFIED STRUCTURAL |
| 37 | SuperGPU | Batch | ordered bounded execution | VERIFIED STRUCTURAL |
| 38 | SuperGPU | MemoryStats | explicit memory-support state | VERIFIED STRUCTURAL |
| 39 | SuperGPU | Health | device/reservation/discovery health | VERIFIED STRUCTURAL |
| 40 | SuperGPU | Shutdown | admission stop + drain | VERIFIED STRUCTURAL |

## Cross-front workstream

- Neural federation: N01–N06 adapters target the single N07 neural surface.
- Parallelism: federated admission must remain bounded and cancellation-aware.
- Fusion: capabilities/tools/agents are composed through ownership and adapters, not copied into N07.
- Runtime proof: structural status is distinct from live six-runtime commissioning.
- CI proof: exact-head format/vet/test/race/build remains mandatory before release.

## Current execution evidence

- N07 neural federation parallel admission was hardened with a 32-task concurrency bound and parent-context propagation.
- Legacy `Assign`, `Parallel` and `Broadcast` APIs remain compatible through context-aware implementations.
- Result ordering remains deterministic.
- Cancellation is represented in per-task results rather than silently dropped.

## Next actions

- Reconcile latest N01–N06 neural adapter heads and response-signature semantics against N07.
- Re-run exact-head CI and record the observed result.
- Commission live bidirectional Mesh only after peer runtimes are concurrently reachable.
- Continue function/tool/agent fusion only after ownership and contract compatibility are confirmed.
