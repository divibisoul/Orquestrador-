# Octacore — System Processor

## Definição obrigatória

**Octacore é uma GPU de SISTEMA:** uma fabric federada de execução paralela sobre oito núcleos de domínio **G0–G7**, implementada como software distribuído e integrada ao SOUL Mesh e ao VagusBus.

Ela **não é uma CPU octa-core de silício**, não cria hardware inexistente e não afirma existir CUDA, GPU/NPU física ou acelerador que não tenha evidência.

Octacore **não remove funções existentes**. Ele amplia, otimiza e conecta os runtimes existentes através de adapters e contratos verificáveis.

## Domínios

| Slot | Núcleo | Função |
|---|---|---|
| G0 | SARA | kernel regenerativo autoritativo |
| G1 | N01 | ingress/host gateway |
| G2 | N02 | conversation turns |
| G3 | N03 | perception/multimodal preparation |
| G4 | N04 | tools/documents/research boundary |
| G5 | N05 | dispatch |
| G6 | N06 | cognition/session batching |
| G7 | N07 | scheduler, SuperGPU software runtime, Mesh router, correlation |

A presença de um repositório não é considerada prova de runtime online. O inventory registra a diferença entre implementação, adapter disponível e runtime ainda não verificado.

## Planos de execução

Os jobs usam o contrato OctaCoreJob, compatível por alias com GpuJob, e produzem OctaCoreResult, compatível por alias com GpuResult.

Backends de job:

- IN_PROCESS
- WEBASSEMBLY
- WEBGPU
- REMOTE_MESH

WEBASSEMBLY e WEBGPU só podem ser escolhidos quando um backend real estiver disponível. Ausência resulta em erro determinístico; nunca em simulação.

G0 é especial: operações regenerativas usam a fronteira HTTP real do SARA (sara.cycle, sara.audit, sara.regenerate, sara.state, sara.trace). A preferência genérica de backend do job não pode deslocar a autoridade regenerativa para outro núcleo.

## Paralelismo

parallel_group forma uma frente concorrente real. barrier é uma junção nomeada: jobs posteriores com o mesmo barrier e fora do grupo produtor aguardam os produtores.

O scheduler mede latency_ms e queue_wait_ms, mantém limite de inflight, token-bucket para admissão, circuit breaker com half-open probe e sinais de throttle, halt e resume.

## VagusBus e Mesh

**VagusBus é control plane. Mesh é data/execution plane.**

O Octacore não cria um segundo Mesh. REMOTE_MESH reutiliza o transporte, circuitos e correlation existentes.

Eventos Vagus do Octacore usam:

gpu.submit, gpu.result, gpu.barrier

além de health.*, capability.*, signal.*, sara.*, session.* e research.*.

## Segurança sem simulação

Um job expirado pelo TTL não pode iniciar uma execução nova. Falhas externas são propagadas como erro explícito. Uma capacidade ausente, não registrada ou ainda sem adapter recebe estado PENDING_* ou erro determinístico.

## Estado desta rodada

Round 1 implanta o processador, contrato de job/result, scheduler N07, Vagus control envelope, kernel G0, adapters G4/G6 e testes estruturais. A validação E2E depende dos runners/ambientes reais de cada núcleo e não é inferida a partir da existência dos arquivos.
