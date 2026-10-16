# GRF/GRCE — Instalação e Integração Sistêmica do SOUL
Data: 2026-10-06

## Estado de instalação

O Golden Rule Framework (GRF) 2.0 e o Golden Rule Cycle Executor (GRCE) 2.0 foram adicionados ao N07 como camada transversal de governança/regeneração, sem criar um oitavo núcleo nem um segundo Soul Mesh.

### Componentes instalados

- `grf/types.go`: estados epistêmicos, Context, Evidence, Failure, Opposition, Artifact, State, Capability, Provenance e contrato `GoldenRuleParticipant`.
- `grf/invariants.go`: conjunto canônico I1–I18 e validação obrigatória.
- `grf/participants.go`: boundary adapters dos 10 novos contratos; eles preservam estado epistêmico e não alegam execução dos upstreams.
- `grce/executor.go`: executor da sequência PREFLIGHT → DETECT → EVIDENCE → CHARACTERIZE → DUALIZE → ETR → ANALYZE → MMD → ARA → ETR → ITR → FORM → VALIDATE → FEEDBACK.
- `orchestrator/grce_ops.go`: capacidades Mesh/API para descoberta, bindings, participant ingest e execução do GRCE.
- `integrations/soul-29-capability-fabric.json`: Soul-25 preservado como base e ampliado com quatro nós complementares.
- `integrations/grf/soul-29-participant-contracts.json`: dez contratos GoldenRuleParticipant.
- `integrations/grf/system-binding.json`: conexão com N01–N07, SARA, JEV, HortaCore, VagusBus, Clareira, SuperGPU, Octacore e Orbitador.

## Fontes complementares

### PROJECTED — fonte real materializada
- Autogenesis — `DVampire/Autogenesis`
- octos — `lispking/octos`
- hora-graph-core — `Vivien83/hora-graph-core`
- mycelium — `mycelium-io/mycelium`
- prime-agent — `PrimeIntellect-ai/prime-agent`
- cuda-oxide — `SuperInstance/cuda-oxide`

### BLOCKED — fonte não verificável no momento da instalação
- CogniFold — `OpenNerve/CogniFold`
- belel-protocol — `TTOPM/belel-protocol`
- OpenSIN-Neural-Bus — `OpenSIN-AI/OpenSIN-Neural-Bus`
- functional-graph-agi — `kexi-bq/functional-graph-agi`

Esses quatro permanecem registrados. Nenhuma implementação substituta foi inventada.

## Separação de responsabilidades

- HortaCore transporta/persiste estado.
- VagusBus transporta sinal.
- Soul Mesh distribui mensagens e capacidades.
- SARA mantém a autoridade regenerativa existente de ERU/ARA/ETR/ITR/RGO.
- JEV mantém decisão tipada/guardrails.
- N07 hospeda o GRCE control-plane.
- Os upstreams permanecem fontes de implementação e não se tornam núcleos SOUL.

## Estado operacional

O GRCE está instalado e agora possui um executor com hooks conectados às fronteiras reais de SARA/ERU-ARA-ETR, SuperGPU e RGO, além do canal Nervo Vago unificado para feedback. O código do executor não reproduz internamente as autoridades ARA/ETR/ITR/ERU; ele orquestra e registra as evidências dessas autoridades.

A classificação sistêmica permanece `PROJECTED` até que o E2E do GitHub execute C1→C7→C1 e produza evidência observável. Não há promoção artificial para `REAL` apenas pela existência do código.

## Prova exigida antes de ACTIVE

1. contexto C completo em cada ciclo;
2. proveniência com parent_hash/input_hash/output_hash/sequence_index;
3. SuperGPU e Octacore com resultados mergeáveis;
4. ARA → ETR → ITR sem atalho;
5. validação ETR + ITR + RGO;
6. rollback preservando evidências em caso de falha;
7. feedback HortaCore + VagusBus + Mesh;
8. um E2E real C1→C7→C1 sem doubles.

