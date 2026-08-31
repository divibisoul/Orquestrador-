# Architecture — Orchestrator Nexus

## Six planes

1. **State Fabric** — versioned state boundary; production consensus can be supplied through `RaftBoundary` without coupling the core to one vendor.
2. **Neo-Cortex Prefrontal** — context, signal fusion, anomaly detection, causal/probabilistic reasoning, planning, decision and feedback.
3. **Orchestrator** — DAG validation, step execution, parallel/distributed execution, retry, circuit breaker, bulkhead, rate limiting and fractal workers.
4. **Super AGI** — provider-neutral generation/inference, memory systems, verification, learning boundaries, model selection and tensor cache.
5. **Mesh** — capability discovery, registration, routing and health endpoints.
6. **API/Control** — HTTP control plane with clean boundaries for future gRPC/Protobuf, mTLS and external adapters.

## Function contract map

### Orchestrator (1–15)
1 CreateWorkflow; 2 ExecuteStep; 3 GetWorkflowStatus; 4 PauseWorkflow; 5 ResumeWorkflow; 6 RollbackWorkflow; 7 RetryFailedStep; 8 ExecuteParallel; 9 ExecuteDistributed; 10 CircuitBreaker; 11 Bulkhead; 12 RateLimiter; 13 SpawnSubOrchestrator; 14 KillSubOrchestrator; 15 RebalanceTasks.

### Neo-Cortex (16–29)
16 ReadContext; 17 FuseSignals; 18 DetectAnomalies; 19 CausalReason; 20 ProbabilisticReason; 21 GeneratePlan; 22 SimulatePlan; 23 PrioritizeGoals; 24 Decide; 25 Delegate; 26 EvaluateOutcome; 27 LearnFromFeedback; 28 OptimizePolicy; 29 ExplainDecision.

### Super AGI (30–60)
30 GenerateText; 31 GenerateEmbedding; 32 GenerateImage; 33 GenerateCode; 34 Classify; 35 Summarize; 36 Translate; 37 VerifyFact; 38 VerifySafety; 39 VerifyCoherence; 40 VerifyCode; 41 WorkingMemory; 42 EpisodicMemory; 43 SemanticMemory; 44 ProceduralMemory; 45 VectorMemory; 46 TrainOnline; 47 FineTuneLoRA; 48 PredictLoRADemand; 49 SwapLoRA; 50 ReplayExperience; 51 Inference; 52 BatchInference; 53 DynamicQuantization; 54 SelectBestModel; 55 CacheTensor; 56 ProfileModel; 57 EstimateCost; 58 ExplainInference; 59 MonitorDrift; 60 AutoRetry.

## Engineering truthfulness

The foundation implements deterministic and provider-neutral behavior where a real external model/runtime is not present. Image generation, translation, online weight training, LoRA training/swapping and accelerator execution therefore expose explicit extension boundaries rather than pretending to execute a model that is not installed. This makes tests honest and production integration possible.

## Performance

Targets such as P95 decision latency <15 ms, 1,200 RPS, 40% energy reduction, 99.99% control-plane availability and 1→1,000 nodes/30 s are acceptance targets. They require hardware, topology, workload and benchmark evidence before being marked achieved.

## Failure model

DAG cycles are rejected; unknown dependencies are rejected; node selection requires `ready` status and declared capability; retries have bounded backoff; workflow state is explicit; and the control plane shuts down gracefully.
