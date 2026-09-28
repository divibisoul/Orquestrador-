# Archaeological Audit — Orquestrador

> **PROGRESSO SUPER AGI: 15% (anterior: 15%)**
>
> Indicador interno provisório: a auditoria não adiciona implementação funcional; ela transforma arquitetura histórica e divergências de branches em estado rastreável. O percentual **não é uma medida científica de AGI** e não substitui os gates de execução real.

**Data do levantamento:** 2026-09-28  
**Repositório:** `divibisoul/Orquestrador-`  
**Base auditada:** `main` @ `a84508b930b691fad9d900d1761263f090df4d17`  
**PR de fundação em análise:** #41, head observado durante o levantamento  
**Princípio de preservação:** nenhum subsistema é considerado descartável só porque está fora do `main`.

---

## 1. Escopo e método

Esta auditoria cruza quatro superfícies:

1. árvore e código do `main`;
2. branches históricos e de trabalho visíveis no repositório;
3. divergência de branches relevantes contra `main`;
4. histórico de commits pesquisável por subsistema/tema.

### Limitação de ferramenta que afeta a palavra “todos”

O conector GitHub disponível expõe busca/listagem de **branches** e comparação de commits, além de busca de commits, mas **não expõe uma operação de enumeração de tags** nem uma paginação completa de todos os objetos de commit de um repositório.

Consequentemente:

- os **88 branches visíveis** foram inventariados;
- commits históricos foram cruzados por busca e por divergência de branches relevantes;
- não é tecnicamente honesto declarar “todos os commits e todas as tags individualmente verificados” com esta interface;
- tags permanecem **NOT MEASURED** como universo de refs;
- nenhum conteúdo ausente foi inferido como inexistente.

Essa limitação é registrada no relatório em vez de ser escondida.

---

## 2. Regra de classificação

| Estado | Definição usada |
|---|---|
| **ATIVO** | Existe no `main` e há caminho funcional identificável para execução/integração. |
| **ÓRFÃO** | Existe em histórico/branch, mas não possui proprietário/caminho atual claro no `main`. |
| **SOTERRADO** | Existe em branch/commit histórico, mas foi deixado fora da linha principal atual. |
| **INVALIDADO** | O nome/alegação aparece em documentação ou intenção, mas não existe evidência de implementação executável correspondente. |
| **CONGELADO** | Existe contrato/estrutura preservada, mas sua promoção/execução foi deliberadamente mantida fechada por gate. |

**Regra adicional:** “declarado” nunca equivale a “executável”. Capabilities somente entram como executáveis quando houver handler/runtime demonstrável.

---

## 3. Estado de referência do `main`

### 3.1 O PR #41 ainda não é `main`

No momento desta auditoria, o `main` está em:

`a84508b930b691fad9d900d1761263f090df4d17`

O PR #41 é uma linha separada. Portanto, os seguintes artefatos ainda **não estão integrados ao `main)** apenas por existirem no PR:

- E2E N07→N04/N05/N06 real;
- workflow de Quick Tunnel;
- contratos de preservação;
- matriz de comunicação revisada.

### 3.2 E2E sintético ainda encontrado no `main`

A busca no `main` encontrou `httptest.NewServer` em:

`mesh/n01_n07_federation_e2e_test.go`

O PR #41 substitui esse conteúdo por um teste `integration` que usa o `PeerClient` real. Isso é uma **correção proposta**, não uma correção já integrada ao `main`.

Há ainda outros usos de `httptest.NewServer` em testes auxiliares do `main`, por exemplo:

- `jev/client_test.go`
- `backend/backend_test.go`
- `cmd/release-doctor/main_test.go`
- `backend/sara_proxy_test.go`

Esses usos são compatíveis com testes unitários tradicionais, mas entram em conflito com a diretriz global atual do projeto (“NADA deve usar httptest/mocks/stubs/fake servers”). Na Tarefa 1 eles ficam **registrados como dívida de conformidade**; não serão removidos nesta auditoria.

---

# 4. Inventário dos subsistemas

## 4.1 Clareira

**Estado no `main`: SOTERRADO**

**Localização histórica confirmada:**

- branch `integrate/clareira-octapla-2026-09-25`
- `clareira/bridge.go`
- `shared/clareira-contract.ts`
- `shared/clareira_contract.go`
- alterações em `mesh/http.go`
- alteração em `orchestrator/topology.go`

**Evidência adicional atual:** o `main` contém documentação que referencia Clareira, mas a busca de código não encontrou uma implementação equivalente desses elementos na linha principal.

**Dependências identificadas:**

- Soul Mesh 1.1.0;
- contratos Clareira;
- topologia/orquestração;
- N01 como destino da integração;
- EventBus/bridge no ecossistema N01.

**Bypass pelo PR #41:** **SIM, funcionalmente indireto.** O gate real do PR #41 não exerce o caminho Clareira; ele prova apenas native capabilities N04/N05/N06.

**Ação necessária:**

- preservar `clareira/bridge.go` e contratos históricos;
- reconciliar o bridge com o Mesh canônico;
- ligar Clareira à matriz de comunicação sem criar segundo barramento;
- criar prova real N05→N01 para `clareira.ingest`;
- validar observabilidade de ingestão/processamento/erro.

---

## 4.2 EventBus

**Estado no `main`: SOTERRADO/EXTERNO AO N07**

**Localização confirmada no ecossistema:**

- N01: `src/core/EventBus.ts`
- referências arquiteturais no sistema N01;
- integrações históricas de Clareira usam esse caminho.

**Dependências:**

- N01 runtime;
- Clareira;
- canais internos do núcleo;
- adaptação para eventos federados somente quando requerido.

**Bypass pelo PR #41:** **NÃO diretamente.** O PR #41 não substitui EventBus. Porém o novo gate Mesh não prova a cadeia EventBus→Mesh.

**Ação:**

- manter EventBus como mecanismo interno de N01;
- adicionar bridge Mesh somente no limite externo;
- provar propagação com correlação;
- não substituí-lo por um “event bus” paralelo no N07.

---

## 4.3 ProcessingNodes

**Estado:** **SOTERRADO**

**Localização:** estrutura associada ao Clareira/N01; não localizada como implementação equivalente no `main` do N07.

**Dependências:**

- Clareira;
- EventBus;
- HomeostasisManager;
- canais internos.

**Bypass pelo PR #41:** **SIM, por escopo do gate.**

**Ação:**

- reconstruir inventário fonte no N01;
- registrar cada ProcessingNode;
- estabelecer adaptador Mesh somente quando a unidade precisar sair do N01;
- adicionar regressão real.

---

## 4.4 InformationChannels

**Estado:** **SOTERRADO**

**Localização:** arquitetura Clareira/N01; não encontrado como runtime equivalente no `main` N07.

**Dependências:**

- ProcessingNodes;
- EventBus;
- Clareira;
- homeostase.

**Bypass pelo PR #41:** **SIM, por escopo.**

**Ação:**

- preservar canais internos;
- mapear entrada/saída para canais Mesh quando federados;
- não duplicar transporte.

---

## 4.5 HomeostasisManager

**Estado:** **SOTERRADO / NÃO MEDIDO**

**Localização:** pertencente à arquitetura interna de N01/Clareira; não encontrado como implementação correspondente no `main` N07.

**Dependências:**

- sinais internos;
- estado dos nós;
- EventBus;
- Clareira.

**Bypass pelo PR #41:** **SIM, por escopo.**

**Ação:**

- localizar implementação canônica em N01;
- preservar estado e ciclo;
- expor somente telemetria/controle necessária via Mesh;
- adicionar prova de preservação após integração.

---

## 4.6 SARA

**Estado:** **ATIVO; handoff externo CONGELADO**

**Localização atual:**

- `backend/sara_proxy.go`
- `orchestrator/engine.go`
- `cmd/nexus/main.go`
- contratos/documentação SARA;
- testes de correlação e transporte.

**Dependências:**

- `SARA_SERVICE_URL`
- `SARA_SERVICE_TOKEN` para operações protegidas;
- correlação;
- Mesh HMAC quando atravessa Mesh.

**Histórico relevante:**

Há uma sequência de commits de 2026-09-22 a 2026-09-27 cobrindo:

- registro das operações SARA;
- proxy real;
- trace;
- correlação;
- integração com OctaCore;
- propagação de contexto consolidado;
- classificação determinística da ausência do backend.

**Bypass pelo PR #41:** **NÃO substituído.** O commissioning N07→SARA continua separado do gate N04/N05/N06.

**Ação:**

- preservar o proxy;
- abrir gate real N07↔SARA;
- provar health/cycle/audit/regenerate/state/capabilities/trace conforme contratos existentes.

---

## 4.7 Agentes autônomos

**Estado:** **ATIVO**

**Localizações principais identificadas:**

- `octacore/`
- registries de agentes/capabilities em N07;
- `orchestrator/`
- superfícies Mesh de discovery/execução.

**Dependências:**

- Engine N07;
- capability registry;
- PeerClient;
- OctaCore;
- SARA em operações de controle quando configurado.

**Bypass pelo PR #41:** **NÃO removido.** O PR prova apenas um subconjunto native capability dos peers.

**Ação:**

- manter registries;
- alinhar discovery com executabilidade real;
- transformar capacidades declaradas sem handler em estado explícito, nunca em PASS.

---

## 4.8 Registries

**Estado:** **ATIVO**

**Localizações:**

- registries de capabilities em `orchestrator/`;
- registros de operações;
- registry de peers em `mesh/peers.go`;
- estruturas de execução OctaCore.

**Dependências:**

- protocolo Mesh;
- engine;
- PeerClient;
- capability discovery.

**Bypass pelo PR #41:** **NÃO.**

**Ação:** nenhuma remoção. A auditoria posterior deve verificar consistência entre “registered”, “discovered” e “executable”.

---

## 4.9 Memória de longo prazo

**Estado:** **CONGELADO / PARCIAL**

**Localizações atuais/históricas:**

- persistência/backend Supabase;
- `backend/supabase.go`;
- `migrations/...`;
- superfícies de identidade/estado.

**Evidência importante:** a base atual tem persistência real, mas a separação cognitiva explícita entre memória episódica, semântica e procedural foi encontrada em branches de evolução cognitiva, não como camada consolidada no `main`.

**Branch histórico relevante:**

`feat/soul-cognitive-capabilities-r1`

contém:

- `cognitive/memory.go`
- `cognitive/types.go`
- `cognitive/loop.go`
- `cognitive/planner.go`
- `cognitive/planned_run.go`
- `cognitive/runtime.go`

**Bypass pelo PR #41:** **SIM.**

**Ação:**

- preservar persistência atual;
- reconciliar o modelo cognitivo histórico com o backend existente;
- não criar memória fictícia;
- provar escrita/leitura real e recuperação por contexto.

---

## 4.10 Meta-cognição

**Estado:** **SOTERRADO**

**Localização histórica:**

`feat/soul-cognitive-capabilities-r1` e `feat/reconcile-mesh-cognitive-octacore-2026-09-28`

Arquivos candidatos:

- `cognitive/loop.go`
- `cognitive/runtime.go`
- `cognitive/types.go`

**Dependências:**

- memória;
- execução;
- planner;
- estado do runtime.

**Bypass pelo PR #41:** **SIM.**

**Ação:**

- extrair contratos concretos;
- reconciliar com N07 Engine;
- nunca declarar “system knows what it knows” apenas por documentação.

---

## 4.11 Goal-setting

**Estado:** **SOTERRADO**

**Localização histórica:**

- `cognitive/planner.go`
- `cognitive/planned_run.go`
- `cognitive/operations.go`

**Branch:** `feat/soul-cognitive-capabilities-r1`

**Dependências:**

- memória;
- planner;
- execution engine;
- observação de resultado.

**Bypass pelo PR #41:** **SIM.**

**Ação:**

- reconectar planner ao Engine;
- criar objetivos como artefatos reais;
- provar geração → planejamento → execução → observação.

---

## 4.12 Auto-modificação

**Estado:** **INVALIDADO COMO CAPACIDADE OPERACIONAL ATUAL**

Não encontrei evidência suficiente no `main` para afirmar que exista um runtime de auto-modificação autônoma, seguro e verificável.

Existem documentos e branches de “Super AGI foundation”, mas isso não prova auto-modificação executável.

**Bypass pelo PR #41:** não aplicável; não estava estabelecido como runtime operacional.

**Ação:**

- não implementar por declaração;
- primeiro localizar mecanismo fonte e limites;
- exigir sandbox/rollback/proveniência/teste antes de abrir o gate.

---

## 4.13 Meta-aprendizado

**Estado:** **SOTERRADO**

**Localização histórica confirmada no branch:**

`feature/trinity-production-hardening-v1`

Arquivos:

- `core/prefrontal/meta_rl.go`
- `core/prefrontal/prefrontal.go`
- `core/prefrontal/executive.go`

**Dependências:**

- Prefrontal;
- aprendizagem/recompensa;
- estado;
- execução.

**Bypass pelo PR #41:** **SIM.**

**Ação:**

- preservar o código histórico;
- determinar compatibilidade real com N07;
- integrar somente após contrato e teste de execução.

---

## 4.14 HortaCore

**Estado:** **SOTERRADO**

**Branch:** `feat/aeternum-hortacore-sara-chimera`

**Localização:**

- `aeternum/hortacore.go`
- `aeternum/hortacore_test.go`
- alterações em `cmd/nexus/main.go`

**Dependências:**

- SARA;
- estado/ciclo;
- orquestração.

**Bypass pelo PR #41:** **SIM.**

**Ação:**

- reconciliar HortaCore com a autoridade N07/SARA existente;
- manter seu código e testes;
- provar integração por caminho real.

---

## 4.15 Neural Fabric

**Estado:** **SOTERRADO / HISTÓRICO**

**Localização histórica confirmada em:**

`feature/trinity-production-hardening-v1`

Arquivos:

- `core/neuralfabric/fabric.go`
- `core/neuralfabric/router.go`
- `core/neuralfabric/runtime.go`
- `core/neuralfabric/feedback.go`
- `core/neuralfabric/decision_tree.go`

**Bypass pelo PR #41:** **SIM.**

**Ação:**

- comparar com `neural/` atual do N07;
- determinar duplicação, compatibilidade ou complementação;
- integrar sem substituir o motor atual.

---

## 4.16 OctaCore

**Estado:** **ATIVO**

**Localização atual:**

- `octacore/processor.go`
- `octacore/operations.go`
- `octacore/scheduler.go`
- `octacore/types.go`
- `octacore/scheduler_test.go`

**Evidência histórica:** Stage 1 foi integrado ao `main` pelo PR #40.

**Dependências:**

- Engine N07;
- Mesh;
- SARA/Vagus;
- scheduler;
- backend quando disponível.

**Bypass pelo PR #41:** **NÃO.** O PR #41 adiciona commissioning da federação ao lado do OctaCore.

**Ação:** manter ativo e posteriormente incluir no seven-core smoke real.

---

## 4.17 JEV

**Estado:** **ATIVO**

**Localização atual:**

- `jev/client.go`
- `jev/client_test.go`
- `orchestrator/jev.go`
- `orchestrator/topology.go`
- `backend/backend.go`
- `cmd/nexus/main.go`
- `docs/JEV_INTEGRATION.md`

O branch histórico `feat/jev-soul-integration` contém a implementação que posteriormente chegou ao `main`.

**Bypass pelo PR #41:** **NÃO removido.** O JEV permanece fora do gate Phase-1 específico.

**Ação:** preservar e abrir seu próprio runtime/e2e gate depois da fundação Mesh.

---

## 4.18 ERU / RGO / MMD

### ERU

**Estado:** **ATIVO / PARCIAL**

Há referências funcionais/documentais e integração com operações/backend, mas a auditoria não encontrou evidência suficiente para afirmar que o papel completo do ERU conceituado anteriormente está consolidado como runtime independente.

### RGO

**Estado:** **INVALIDADO COMO RUNTIME ATUAL**

Busca no `main` não encontrou implementação operacional inequívoca com o nome/contrato esperado.

### MMD

**Estado:** **INVALIDADO COMO RUNTIME ATUAL**

Busca no `main` não encontrou implementação operacional inequívoca.

**Ação comum:** preservar documentação/artefatos históricos, localizar implementações reais em branches anteriores e só então criar integração verificável.

---

# 5. 72 nódulos arquiteturais

**Estado:** **CONGELADO / NÃO MEDIDO NODE-BY-NODE**

A documentação atual afirma um inventário de 72 nódulos arquiteturais protegidos, mas a pesquisa do código do `main` não produziu um inventário fonte-a-fonte de 72 unidades individuais.

Logo:

- o **escopo é preservado**;
- a **contagem não é promovida para prova de execução**;
- o estado de cada nódulo ainda precisa ser enumerado;
- nenhum nódulo será removido para simplificar o commissioning.

**Bypass pelo PR #41:** **SIM.** O gate real testa poucas capacidades de três peers e não exercita os 72 nódulos.

**Ação necessária:**

1. construir índice fonte dos 72;
2. atribuir proprietário/núcleo;
3. localizar dependências;
4. verificar estado;
5. mapear canal Mesh quando federado;
6. criar teste de regressão por grupo;
7. somente então classificar cada nódulo.

---

# 6. Branch archaeology

Foram identificados **88 branches visíveis** no repositório.

## 6.1 Branches diretamente relevantes para a arqueologia

| Branch | Relação com `main` | Achado |
|---|---|---|
| `feat/aeternum-8-modules-2026-09-23` | divergente | GenesisModule / OmniModeModule históricos |
| `feat/aeternum-hortacore-sara-chimera` | divergente | HortaCore real + integração SARA |
| `feat/expose-neural-runtime-parameters-2026-09-28` | ahead 3 | exposição de parâmetros do runtime neural |
| `feat/forensic-reconcile-n07-2026-09-28` | ahead 4 | reconciliação/scheduler |
| `feat/jev-soul-integration` | divergente | JEV |
| `feat/octacore-stage1-current-2026-09-27` | atrás | já absorvido em main |
| `feat/reconcile-mesh-cognitive-octacore-2026-09-28` | ahead 32 | camada cognitiva + OctaCore |
| `feat/soul-cognitive-capabilities-r1` | ahead 18 | memória/planner/loop cognitivo |
| `feat/soul-federation-expansion` | divergente | JEV + backend + expansão Mesh |
| `feat/soul-langgraph-nemo-bridge` | divergente | bridges LangGraph/NVIDIA |
| `feat/soul-super-agi-foundation` | ahead 31 | contratos/gates/fundação Super AGI |
| `integrate/clareira-octapla-2026-09-25` | divergente | Clareira + contratos |
| `fusion/n01-n06-n07-mainline-live` | divergente | catálogo de fusão N01/N06 |
| `feature/trinity-production-hardening-v1` | fortemente divergente | Prefrontal, NeuralFabric, Trinity, SuperAGI e múltiplos subsistemas |
| `feature/unified-backend-supabase-storage` | divergente | backend, storage, Supabase e SuperGPU |
| `feature/nexus-finalization-v1` | histórico | linha anterior do núcleo |
| `feature/nexus-runtime-hardening-batch1` | histórico | endurecimento do runtime |
| `feature/readiness-dashboard` | histórico | observabilidade/readiness |
| `feature/trinity-fundamental-v1` | histórico | primeira camada Trinity |
| `audit/2026-09-01-hardening` | histórico | hardening |
| `forensic/reconciliation-2026-09-27` | histórico | reconciliação |
| `forensic/reconciliation-current-2026-09-27` | histórico | reconciliação corrente |
| `recovery/consolidation` | histórico | consolidação |
| `release/v1-online-hardening-2026-09-01` | histórico | release/hardening |
| `soul/lei-zero-f0-integrity` | histórico | integridade/lei zero |
| `upgrade/soul-federation-live-2026-09-01` | histórico | federação |
| `verify/n07-supercompute*` | histórico | séries de verificação SuperCompute |

### Branches de evolução repetitiva

Há uma família grande:

- `upgrade/n07-final-v1` … `v8`
- `upgrade/n07-production-v7` … `v18`
- `verify/n07-supercompute` … `-9`

Essas linhas não devem ser tratadas como “código lixo” automaticamente. São **arqueologia de evolução**. Precisam ser comparadas contra o estado consolidado antes de qualquer decisão de descarte.

---

# 7. Commit archaeology

A busca histórica produziu evidência forte para dois grandes eixos.

## 7.1 SARA

A sequência de commits inclui, entre outros:

- `eb7a3b93...` — proxy real N07→SARA;
- `9297def4...` — registro de operações regenerativas;
- `1de4022e...` — inventário completo;
- `399ac874...` / `8abce150...` — trace;
- `d00902b4...` — roteamento das operações por intent/execute;
- `83ece1af...` / `b33cf720...` — propagação e teste de correlação;
- `d8748d24...` — contexto federado consolidado;
- `e0b220c4...` — conexão OctaCore/SuperGPU/SARA;
- `24d9a84c...` — classificação determinística da ausência do backend.

Conclusão: SARA não é um artefato órfão; é um subsistema real que chegou ao `main`.

## 7.2 OctaCore

A sequência entre `33391328...` e `aa1faa49...` mostra:

- contratos de job/Vagus;
- boundary sobre runtime N07;
- scheduler bounded-parallel;
- backend/SARA;
- barreiras;
- testes;
- Stage 1 integrado.

Conclusão: OctaCore é **ATIVO**.

## 7.3 Cérebro cognitivo

Os branches `feat/soul-cognitive-capabilities-r1` e `feat/reconcile-mesh-cognitive-octacore-2026-09-28` contêm uma camada coerente:

- `memory.go`;
- `loop.go`;
- `planner.go`;
- `planned_run.go`;
- `operations.go`;
- `runtime.go`;
- testes.

Conclusão: não é seguro recriar Tarefa 4 do zero sem antes reconciliar esses artefatos. Eles são o principal candidato de **SOTERRADO → REINTEGRÁVEL**.

---

# 8. Bypass específico do PR #41

O PR #41 não remove os subsistemas. Ele altera principalmente o **mecanismo de prova da federação**.

### O que o PR #41 realmente bypassa

| Subsistema | Bypass pelo novo gate |
|---|---|
| Clareira | SIM, por não ser capability alvo |
| EventBus | SIM, por não ser capability alvo |
| ProcessingNodes | SIM |
| InformationChannels | SIM |
| HomeostasisManager | SIM |
| SARA | NÃO; continua em gate separado |
| Agentes | NÃO; discovery/execution continuam válidos |
| Registries | NÃO |
| Memória cognitiva | SIM |
| Meta-cognição | SIM |
| Goal-setting | SIM |
| Auto-modificação | não operacional no gate |
| Meta-learning | SIM |
| HortaCore | SIM |
| NeuralFabric | SIM |
| OctaCore | não substituído |
| JEV | não substituído |
| 72 nódulos | SIM, por escopo |

**Interpretação:** o bypass é **de cobertura de teste**, não de substituição arquitetural.

---

# 9. Contratos que devem permanecer preservados

A arqueologia não autoriza remoção de:

- Clareira;
- EventBus;
- ProcessingNodes;
- InformationChannels;
- HomeostasisManager;
- registries;
- agentes;
- memória/persistência;
- OctaCore;
- SARA;
- JEV;
- NeuralFabric;
- Prefrontal;
- Trinity;
- SuperGPU/SuperCompute;
- contratos ERU;
- artefatos históricos que ainda tenham dependentes;
- os 72 nódulos protegidos.

Qualquer eventual aposentadoria futura exigirá:

**evidência → dependências → substituto → compatibilidade → regressão → migração.**

---

# 10. Estado consolidado da arqueologia

| Subsistema | Estado | Necessita reintegração |
|---|---|---|
| Clareira | SOTERRADO | SIM |
| EventBus | SOTERRADO/EXTERNO N01 | SIM, via bridge |
| ProcessingNodes | SOTERRADO | SIM |
| InformationChannels | SOTERRADO | SIM |
| HomeostasisManager | SOTERRADO/NOT MEASURED | SIM |
| SARA | ATIVO / CONGELADO no handoff | SIM, gate real |
| Agentes autônomos | ATIVO | expansão/teste |
| Registries | ATIVO | reconciliação |
| Long-term memory | CONGELADO/PARCIAL | SIM |
| Meta-cognição | SOTERRADO | SIM |
| Goal-setting | SOTERRADO | SIM |
| Auto-modificação | INVALIDADO como runtime atual | primeiro localizar contrato real |
| Meta-learning | SOTERRADO | SIM |
| HortaCore | SOTERRADO | SIM |
| NeuralFabric | SOTERRADO | SIM |
| OctaCore | ATIVO | expandir cobertura |
| JEV | ATIVO | expandir cobertura |
| ERU | ATIVO/PARCIAL | verificar contrato |
| RGO | INVALIDADO como runtime atual | arqueologia adicional |
| MMD | INVALIDADO como runtime atual | arqueologia adicional |
| 72 nódulos | CONGELADO / NÃO MEDIDO | inventário fonte-a-fonte |

---

# 11. Decisões arquiteturais decorrentes

### D1 — Não recriar o cognitivo sem reconciliar o histórico

Os branches cognitivos possuem uma implementação concreta suficientemente extensa para serem considerados patrimônio arquitetural.

### D2 — Clareira deve voltar pelo contrato existente

O branch `integrate/clareira-octapla-2026-09-25` fornece uma base muito mais segura que inventar uma Clareira nova.

### D3 — N07 continua autoridade de orquestração

Nenhum subsistema histórico deve ganhar um segundo orquestrador concorrente.

### D4 — Mesh continua sendo o transporte federativo canônico

EventBus permanece interno quando for interno. Clareira não ganha um transporte paralelo.

### D5 — Declaração de capability precisa ser auditável

O caso N04 demonstra o risco: execução existente e discovery inconsistente. O estado canônico deve ser:

`DECLARED + HANDLER + EXECUTED`

antes de considerar uma capability operacional.

### D6 — “ONLINE” continua sendo evidência, não documentação

Branch, código, README, workflow ou capability declaration nunca substituem uma transação real.

---

# 12. Próximas ações liberadas pela Tarefa 1

A Tarefa 1 libera a preparação das seguintes frentes, **mas não as executa neste PR**:

1. reabrir Clareira e canais internos através do Mesh;
2. reconciliar a camada cognitiva histórica;
3. comissionar N07→N01/N02/N03;
4. abrir gate real de SARA;
5. construir inventário fonte dos 72 nódulos;
6. mapear NeuralFabric/Prefrontal/Trinity para o runtime atual;
7. separar “declared”, “registered”, “executable”, “measured”.

---

# 13. Limitações explícitas

Não foi possível, com a interface GitHub instalada nesta execução:

- enumerar tags do repositório;
- baixar/reconstruir localmente o grafo completo de todos os objetos Git;
- garantir inspeção individual de cada commit antigo.

Isso significa **NOT MEASURED**, não “inexistente”.

A auditoria também não declara que um subsistema externo a `Orquestrador-` é inválido apenas porque seu código está em N01/N02/N03. A autoridade do núcleo proprietário deve ser preservada.

---

# 14. Resultado da Tarefa 1

**STATUS: PASS — AUDITORIA ARQUEOLÓGICA DOCUMENTAL CONCLUÍDA, COM LIMITAÇÕES EXPLICITAMENTE REGISTRADAS.**

**Não houve exclusão, simplificação ou substituição de código.**

O principal resultado é uma mudança de modelo: a arquitetura histórica foi separada em:

- **ATIVO**
- **SOTERRADO**
- **CONGELADO**
- **INVALIDADO**
- **ÓRFÃO**

e cada elemento recebeu uma ação concreta de reintegração.

**Tarefa 2, Tarefa 3, Tarefa 4 e Tarefa 5 NÃO foram executadas neste PR.**
