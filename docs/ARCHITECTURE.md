# ORCHESTRATOR-NEXUS — Architecture

## Integrated cognitive loop

```text
Goal -> Neo-Cortex -> Neural Control Fabric -> Orchestrator -> Compute Fabric
  ^                                                           |
  |----------- SuperAGI / verification / memory / telemetry ---|
```

## Six planes

0. **State Fabric:** durable control-plane metadata, checkpoints, cache and future Raft/CRDT adapters.
1. **Neo Cortex Prefrontal:** perception → reasoning → planning → decision → evaluation → adaptation.
2. **Orchestrator:** workflow execution, resilience, distribution and fractal scaling.
3. **Super AGI:** model-provider abstraction, verification, memory, learning and inference.
4. **Mesh:** capability discovery, routing and transport adapters.
5. **Cross-cutting:** security, observability, API and compatibility.
6. **Compute + Neural Control:** heterogeneous CPU/GPU/NPU execution plus learned routing and prediction.

## Function registry

**Orchestrator 1–15:** CreateWorkflow, ExecuteStep, GetWorkflowStatus, PauseWorkflow, ResumeWorkflow, RollbackWorkflow, RetryFailedStep, ExecuteParallel, ExecuteDistributed, CircuitBreaker, Bulkhead, RateLimiter, SpawnSubOrchestrator, KillSubOrchestrator, RebalanceTasks.

**Neo Cortex 16–29:** ReadContext, FuseSignals, DetectAnomalies, CausalReason, ProbabilisticReason, GeneratePlan, SimulatePlan, PrioritizeGoals, Decide, Delegate, EvaluateOutcome, LearnFromFeedback, OptimizePolicy, ExplainDecision.

**Super AGI 30–60:** GenerateText, GenerateEmbedding, GenerateImage, GenerateCode, Classify, Summarize, Translate, VerifyFact, VerifySafety, VerifyCoherence, VerifyCode, WorkingMemory, EpisodicMemory, SemanticMemory, ProceduralMemory, VectorMemory, TrainOnline, FineTuneLoRA, PredictLoRADemand, SwapLoRA, ReplayExperience, Inference, BatchInference, DynamicQuantization, SelectBestModel, CacheTensor, ProfileModel, EstimateCost, ExplainInference, MonitorDrift, AutoRetry.

## Compute Fabric

The Compute Fabric abstracts CPU/GPU/NPU capabilities, precision, memory, thermal state, power and utilization. Scheduling is capability-driven rather than hardware-coupled. Batching, microbatching, quantization, migration and telemetry are exposed as replaceable policies.

## Neural Control Fabric

The Neural Control Fabric contains state/task encoders, latency/cost/energy/quality/failure predictors, route selection and an experience-learning boundary. The repository starts with deterministic baselines so the system remains testable; learned models can be plugged in behind stable interfaces.

## Performance contracts

The specification targets <15 ms P95 decision-to-node-start, 1,200 RPS on a six-node consumer-GPU reference cluster, up to 40% power reduction for eligible non-critical workloads, 99.99% control-plane availability, and 1→1,000-node scale-out under 30 seconds. These are **measurable targets**, not guarantees. Benchmarks must establish hardware, model, payload and network conditions before claiming compliance.

## Design rules

- Provider and substrate agnostic.
- No mandatory model vendor.
- No cognitive module directly owns transport or storage.
- External effects pass through policy and execution boundaries.
- Every production performance claim requires benchmark evidence.
